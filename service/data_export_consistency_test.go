package service

import (
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cacheSize() int {
	model.CacheQuotaDataLock.Lock()
	defer model.CacheQuotaDataLock.Unlock()
	return len(model.CacheQuotaData)
}

func TestRecordTaskBillingLog_NoQuotaExportWhenLogInsertFails(t *testing.T) {
	truncate(t)
	seedUser(t, 1001, 10000)

	require.NoError(t, model.DB.Exec("DROP TABLE logs").Error)
	t.Cleanup(func() {
		require.NoError(t, model.DB.AutoMigrate(&model.Log{}))
	})

	model.RecordTaskBillingLog(model.RecordTaskBillingLogParams{
		UserId:           1001,
		LogType:          model.LogTypeConsume,
		Content:          "test",
		ChannelId:        1,
		ModelName:        "seedance-2.0",
		Quota:            123,
		PromptTokens:     0,
		CompletionTokens: 88,
		QuotaDelta:       123,
	})

	assert.Equal(t, 0, cacheSize(), "log insert failed, so quota_data cache should stay untouched")
}

func TestRecordConsumeLog_NoQuotaExportWhenLogInsertFails(t *testing.T) {
	truncate(t)
	seedUser(t, 1002, 10000)

	require.NoError(t, model.DB.Exec("DROP TABLE logs").Error)
	t.Cleanup(func() {
		require.NoError(t, model.DB.AutoMigrate(&model.Log{}))
	})

	gin.SetMode(gin.ReleaseMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", "test_user")

	model.RecordConsumeLog(c, 1002, model.RecordConsumeLogParams{
		ChannelId:        1,
		PromptTokens:     0,
		CompletionTokens: 66,
		ModelName:        "seedance-2.0",
		TokenName:        "tk",
		Quota:            99,
		Content:          "test",
		TokenId:          1,
		UseTimeSeconds:   1,
		IsStream:         false,
		Group:            "default",
	})

	assert.Equal(t, 0, cacheSize(), "log insert failed, so quota_data cache should stay untouched")
}

func TestSaveQuotaDataCache_RollbackFailedItemsAndRetry(t *testing.T) {
	truncate(t)
	seedUser(t, 1003, 10000)

	// Queue one dashboard delta in cache.
	model.LogQuotaData(1003, "test_user", "seedance-2.0", 321, 1_700_000_000, 77)
	assert.Equal(t, 1, cacheSize())

	// Simulate transient DB failure: quota_data table unavailable.
	require.NoError(t, model.DB.Exec("DROP TABLE quota_data").Error)
	t.Cleanup(func() {
		require.NoError(t, model.DB.AutoMigrate(&model.QuotaData{}))
	})

	model.SaveQuotaDataCache()
	assert.Equal(t, 1, cacheSize(), "failed flush should keep data in cache for retry")

	// Restore schema and flush again.
	require.NoError(t, model.DB.AutoMigrate(&model.QuotaData{}))
	model.SaveQuotaDataCache()
	assert.Equal(t, 0, cacheSize(), "successful flush should drain cache")

	var rows int64
	require.NoError(t, model.DB.Table("quota_data").Count(&rows).Error)
	assert.Equal(t, int64(1), rows)

	var qd model.QuotaData
	require.NoError(t, model.DB.Table("quota_data").First(&qd).Error)
	assert.Equal(t, 1003, qd.UserID)
	assert.Equal(t, "seedance-2.0", qd.ModelName)
	assert.Equal(t, 321, qd.Quota)
	assert.Equal(t, 77, qd.TokenUsed)
	assert.Equal(t, 1, qd.Count)
}
