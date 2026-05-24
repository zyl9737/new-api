package model

import (
	"errors"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"

	"github.com/bytedance/gopkg/util/gopool"
	"gorm.io/gorm"
)

const (
	BatchUpdateTypeUserQuota = iota
	BatchUpdateTypeTokenQuota
	BatchUpdateTypeUsedQuota
	BatchUpdateTypeChannelUsedQuota
	BatchUpdateTypeRequestCount
	BatchUpdateTypeCount // if you add a new type, you need to add a new map and a new lock
)

var batchUpdateStores []map[int]int
var batchUpdateLocks []sync.Mutex
var batchUpdateRunLock sync.Mutex

func init() {
	for i := 0; i < BatchUpdateTypeCount; i++ {
		batchUpdateStores = append(batchUpdateStores, make(map[int]int))
		batchUpdateLocks = append(batchUpdateLocks, sync.Mutex{})
	}
}

func InitBatchUpdater() {
	gopool.Go(func() {
		for {
			time.Sleep(time.Duration(common.BatchUpdateInterval) * time.Second)
			batchUpdate()
		}
	})
}

// FlushBatchUpdates force flushes all in-memory batch update deltas to DB.
// It is intended for graceful shutdown to minimize counter drift.
func FlushBatchUpdates() {
	batchUpdate()
}

func addNewRecord(type_ int, id int, value int) {
	batchUpdateLocks[type_].Lock()
	defer batchUpdateLocks[type_].Unlock()
	if _, ok := batchUpdateStores[type_][id]; !ok {
		batchUpdateStores[type_][id] = value
	} else {
		batchUpdateStores[type_][id] += value
	}
}

func batchUpdate() {
	batchUpdateRunLock.Lock()
	defer batchUpdateRunLock.Unlock()

	// check if there's any data to update
	hasData := false
	for i := 0; i < BatchUpdateTypeCount; i++ {
		batchUpdateLocks[i].Lock()
		if len(batchUpdateStores[i]) > 0 {
			hasData = true
			batchUpdateLocks[i].Unlock()
			break
		}
		batchUpdateLocks[i].Unlock()
	}

	if !hasData {
		return
	}

	common.SysLog("batch update started")
	for i := 0; i < BatchUpdateTypeCount; i++ {
		batchUpdateLocks[i].Lock()
		store := batchUpdateStores[i]
		batchUpdateStores[i] = make(map[int]int)
		batchUpdateLocks[i].Unlock()

		if len(store) == 0 {
			continue
		}

		failed := make(map[int]int)
		// TODO: maybe we can combine updates with same key?
		for key, value := range store {
			switch i {
			case BatchUpdateTypeUserQuota:
				err := increaseUserQuota(key, value)
				if err != nil {
					common.SysLog("failed to batch update user quota: " + err.Error())
					failed[key] += value
				}
			case BatchUpdateTypeTokenQuota:
				err := increaseTokenQuota(key, value)
				if err != nil {
					common.SysLog("failed to batch update token quota: " + err.Error())
					failed[key] += value
				}
			case BatchUpdateTypeUsedQuota:
				if err := DB.Model(&User{}).Where("id = ?", key).Updates(
					map[string]interface{}{"used_quota": gorm.Expr("used_quota + ?", value)},
				).Error; err != nil {
					common.SysLog("failed to batch update user used quota: " + err.Error())
					failed[key] += value
				}
			case BatchUpdateTypeRequestCount:
				if err := DB.Model(&User{}).Where("id = ?", key).Update("request_count", gorm.Expr("request_count + ?", value)).Error; err != nil {
					common.SysLog("failed to batch update user request count: " + err.Error())
					failed[key] += value
				}
			case BatchUpdateTypeChannelUsedQuota:
				if err := DB.Model(&Channel{}).Where("id = ?", key).Update("used_quota", gorm.Expr("used_quota + ?", value)).Error; err != nil {
					common.SysLog("failed to batch update channel used quota: " + err.Error())
					failed[key] += value
				}
			}
		}

		if len(failed) > 0 {
			batchUpdateLocks[i].Lock()
			for key, value := range failed {
				batchUpdateStores[i][key] += value
			}
			batchUpdateLocks[i].Unlock()
		}
	}
	common.SysLog("batch update finished")
}

func RecordExist(err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func shouldUpdateRedis(fromDB bool, err error) bool {
	return common.RedisEnabled && fromDB && err == nil
}
