package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetBatchStoresForTest() {
	for i := 0; i < BatchUpdateTypeCount; i++ {
		batchUpdateLocks[i].Lock()
		batchUpdateStores[i] = make(map[int]int)
		batchUpdateLocks[i].Unlock()
	}
}

func getBatchStoreValueForTest(typeID int, id int) int {
	batchUpdateLocks[typeID].Lock()
	defer batchUpdateLocks[typeID].Unlock()
	return batchUpdateStores[typeID][id]
}

func TestFlushBatchUpdates_ApplyUsedQuotaDelta(t *testing.T) {
	truncateTables(t)
	resetBatchStoresForTest()

	user := &User{Id: 10001, Username: "batch_used_quota_user", Status: common.UserStatusEnabled, Quota: 1000}
	require.NoError(t, DB.Create(user).Error)

	oldBatch := common.BatchUpdateEnabled
	common.BatchUpdateEnabled = true
	t.Cleanup(func() { common.BatchUpdateEnabled = oldBatch })

	addNewRecord(BatchUpdateTypeUsedQuota, user.Id, 123)
	addNewRecord(BatchUpdateTypeRequestCount, user.Id, 2)
	FlushBatchUpdates()

	var got User
	require.NoError(t, DB.Select("used_quota", "request_count").Where("id = ?", user.Id).First(&got).Error)
	assert.Equal(t, 123, got.UsedQuota)
	assert.Equal(t, 2, got.RequestCount)
	assert.Equal(t, 0, getBatchStoreValueForTest(BatchUpdateTypeUsedQuota, user.Id))
	assert.Equal(t, 0, getBatchStoreValueForTest(BatchUpdateTypeRequestCount, user.Id))
}

func TestFlushBatchUpdates_RetryFailedTokenDelta(t *testing.T) {
	truncateTables(t)
	resetBatchStoresForTest()

	user := &User{Id: 10002, Username: "batch_token_user", Status: common.UserStatusEnabled, Quota: 1000}
	require.NoError(t, DB.Create(user).Error)
	token := &Token{
		Id:             10002,
		UserId:         user.Id,
		Name:           "batch-token",
		Key:            "sk-batch-token",
		Status:         common.TokenStatusEnabled,
		UnlimitedQuota: false,
		RemainQuota:    500,
		UsedQuota:      0,
	}
	require.NoError(t, DB.Create(token).Error)

	oldBatch := common.BatchUpdateEnabled
	common.BatchUpdateEnabled = true
	t.Cleanup(func() { common.BatchUpdateEnabled = oldBatch })

	addNewRecord(BatchUpdateTypeTokenQuota, token.Id, -50)

	// Force DB write error and ensure delta is re-queued instead of dropped.
	require.NoError(t, DB.Exec("DROP TABLE tokens").Error)
	t.Cleanup(func() {
		require.NoError(t, DB.AutoMigrate(&Token{}))
	})

	FlushBatchUpdates()
	assert.Equal(t, -50, getBatchStoreValueForTest(BatchUpdateTypeTokenQuota, token.Id))

	// Restore table and token row, then flush again; delta should be applied.
	require.NoError(t, DB.AutoMigrate(&Token{}))
	require.NoError(t, DB.Create(token).Error)
	FlushBatchUpdates()

	var got Token
	require.NoError(t, DB.Select("remain_quota", "used_quota").Where("id = ?", token.Id).First(&got).Error)
	assert.Equal(t, 450, got.RemainQuota)
	assert.Equal(t, 50, got.UsedQuota)
	assert.Equal(t, 0, getBatchStoreValueForTest(BatchUpdateTypeTokenQuota, token.Id))
}
