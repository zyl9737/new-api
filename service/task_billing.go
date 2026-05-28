package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// errSettleAlreadyApplied is returned by settleTaskQuotaInTransaction when the
// CAS predicate (WHERE id=? AND quota=preConsumedQuota) matches zero rows,
// meaning another worker already settled this task. Callers should treat this
// as an idempotent no-op rather than a real failure.
var errSettleAlreadyApplied = errors.New("task settlement already applied (CAS no-op)")

// errRefundAlreadyApplied is returned by refundTaskQuotaInTransaction when the
// refund CAS predicate matches zero rows, meaning another worker already
// refunded or otherwise reconciled this task's pre-consumed quota.
var errRefundAlreadyApplied = errors.New("task refund already applied (CAS no-op)")

// LogTaskConsumption 记录任务消费日志和统计信息（仅记录，不涉及实际扣费）。
// 实际扣费已由 BillingSession（PreConsumeBilling + SettleBilling）完成。
func LogTaskConsumption(c *gin.Context, info *relaycommon.RelayInfo) {
	tokenName := c.GetString("token_name")
	logContent := fmt.Sprintf("操作 %s", info.Action)
	// 支持任务仅按次计费
	if common.StringsContains(constant.TaskPricePatches, info.OriginModelName) {
		logContent = fmt.Sprintf("%s，按次计费", logContent)
	} else {
		if len(info.PriceData.OtherRatios) > 0 {
			var contents []string
			for key, ra := range info.PriceData.OtherRatios {
				if 1.0 != ra {
					contents = append(contents, fmt.Sprintf("%s: %.2f", key, ra))
				}
			}
			if len(contents) > 0 {
				logContent = fmt.Sprintf("%s, 计算参数：%s", logContent, strings.Join(contents, ", "))
			}
		}
	}
	other := make(map[string]interface{})
	other["is_task"] = true
	other["request_path"] = c.Request.URL.Path
	if info.PublicTaskID != "" {
		other["task_id"] = info.PublicTaskID
	}
	other["model_price"] = info.PriceData.ModelPrice
	if info.PriceData.ModelRatio > 0 {
		other["model_ratio"] = info.PriceData.ModelRatio
	}
	other["group_ratio"] = info.PriceData.GroupRatioInfo.GroupRatio
	if info.PriceData.GroupRatioInfo.HasSpecialRatio {
		other["user_group_ratio"] = info.PriceData.GroupRatioInfo.GroupSpecialRatio
	}
	if info.IsModelMapped {
		other["is_model_mapped"] = true
		other["upstream_model_name"] = info.UpstreamModelName
	}
	model.RecordConsumeLog(c, info.UserId, model.RecordConsumeLogParams{
		ChannelId: info.ChannelId,
		ModelName: info.OriginModelName,
		TokenName: tokenName,
		Quota:     info.PriceData.Quota,
		Content:   logContent,
		TokenId:   info.TokenId,
		Group:     info.UsingGroup,
		Other:     other,
	})
	model.UpdateUserUsedQuotaAndRequestCount(info.UserId, info.PriceData.Quota)
	model.UpdateChannelUsedQuota(info.ChannelId, info.PriceData.Quota)
}

// ---------------------------------------------------------------------------
// 异步任务计费辅助函数
// ---------------------------------------------------------------------------

// resolveTokenKey 通过 TokenId 运行时获取令牌 Key（用于 Redis 缓存操作）。
// 如果令牌已被删除或查询失败，返回空字符串。
func resolveTokenKey(ctx context.Context, tokenId int, taskID string) string {
	token, err := model.GetTokenById(tokenId)
	if err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("获取令牌 key 失败 (tokenId=%d, task=%s): %s", tokenId, taskID, err.Error()))
		return ""
	}
	return token.Key
}

// taskIsSubscription 判断任务是否通过订阅计费。
func taskIsSubscription(task *model.Task) bool {
	return task.PrivateData.BillingSource == BillingSourceSubscription && task.PrivateData.SubscriptionId > 0
}

// taskAdjustFunding 调整任务的资金来源（钱包或订阅），delta > 0 表示扣费，delta < 0 表示退还。
func taskAdjustFunding(task *model.Task, delta int) error {
	if taskIsSubscription(task) {
		return model.PostConsumeUserSubscriptionDelta(task.PrivateData.SubscriptionId, int64(delta))
	}
	if delta > 0 {
		return model.DecreaseUserQuota(task.UserId, delta, false)
	}
	return model.IncreaseUserQuota(task.UserId, -delta, false)
}

// taskAdjustTokenQuota 调整任务的令牌额度，delta > 0 表示扣费，delta < 0 表示退还。
// 需要通过 resolveTokenKey 运行时获取 key（不从 PrivateData 中读取）。
func taskAdjustTokenQuota(ctx context.Context, task *model.Task, delta int) {
	if task.PrivateData.TokenId <= 0 || delta == 0 {
		return
	}
	tokenKey := resolveTokenKey(ctx, task.PrivateData.TokenId, task.TaskID)
	if tokenKey == "" {
		return
	}
	var err error
	if delta > 0 {
		err = model.DecreaseTokenQuota(task.PrivateData.TokenId, tokenKey, delta)
	} else {
		err = model.IncreaseTokenQuota(task.PrivateData.TokenId, tokenKey, -delta)
	}
	if err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("调整令牌额度失败 (delta=%d, task=%s): %s", delta, task.TaskID, err.Error()))
	}
}

// taskBillingOther 从 task 的 BillingContext 构建日志 Other 字段。
func taskBillingOther(task *model.Task) map[string]interface{} {
	other := make(map[string]interface{})
	if bc := task.PrivateData.BillingContext; bc != nil {
		other["model_price"] = bc.ModelPrice
		if bc.ModelRatio > 0 {
			other["model_ratio"] = bc.ModelRatio
		}
		other["group_ratio"] = bc.GroupRatio
		if len(bc.OtherRatios) > 0 {
			for k, v := range bc.OtherRatios {
				other[k] = v
			}
		}
	}
	props := task.Properties
	if props.UpstreamModelName != "" && props.UpstreamModelName != props.OriginModelName {
		other["is_model_mapped"] = true
		other["upstream_model_name"] = props.UpstreamModelName
	}
	return other
}

// taskModelName 从 BillingContext 或 Properties 中获取模型名称。
func taskModelName(task *model.Task) string {
	if bc := task.PrivateData.BillingContext; bc != nil {
		if name := strings.TrimSpace(bc.OriginModelName); name != "" {
			return name
		}
	}
	if name := strings.TrimSpace(task.Properties.OriginModelName); name != "" {
		return name
	}
	if name := strings.TrimSpace(task.Properties.UpstreamModelName); name != "" {
		return name
	}
	action := strings.TrimSpace(task.Action)
	platform := strings.TrimSpace(string(task.Platform))
	if action != "" && platform != "" {
		return platform + ":" + action
	}
	if action != "" {
		return action
	}
	if platform != "" {
		return platform
	}
	return "task:unknown"
}

// taskSettleLogTokens derives prompt/completion tokens for settlement logs.
// We keep prompt+completion aligned with the billing token basis (effective
// total tokens) so log aggregation and token-based recalculation stay
// consistent.
func taskSettleLogTokens(taskResult *relaycommon.TaskInfo, effectiveTotalTokens int) (promptTokens int, completionTokens int) {
	if effectiveTotalTokens <= 0 {
		return 0, 0
	}
	if taskResult == nil {
		return 0, effectiveTotalTokens
	}

	// When completion_tokens is available and valid, split total into
	// prompt/output parts. Otherwise treat all effective tokens as output-side.
	if taskResult.CompletionTokens > 0 && taskResult.CompletionTokens <= effectiveTotalTokens {
		return effectiveTotalTokens - taskResult.CompletionTokens, taskResult.CompletionTokens
	}
	return 0, effectiveTotalTokens
}

// RefundTaskQuota 统一的任务失败退款逻辑。
// 当异步任务失败时，将预扣的 quota 退还给用户（支持钱包和订阅），并退还令牌额度。
func RefundTaskQuota(ctx context.Context, task *model.Task, reason string) {
	quota := task.Quota
	if quota == 0 {
		return
	}
	updatedAt := common.GetTimestamp()

	// Flush pending batch updates so the pre-charge is already applied to the
	// user's quota field before the refund transaction reads it.  Without this,
	// the refund can read a stale (higher) quota, and the pending batch update
	// will later subtract the pre-charge again — causing the total
	// (quota + used_quota) to silently shrink.
	model.FlushBatchUpdates()

	if err := refundTaskQuotaInTransaction(task, quota, updatedAt); err != nil {
		if errors.Is(err, errRefundAlreadyApplied) {
			logger.LogInfo(ctx, fmt.Sprintf("RefundTaskQuota: refund skipped (already applied) for task %s", task.TaskID))
			return
		}
		logger.LogWarn(ctx, fmt.Sprintf("RefundTaskQuota: refund transaction failed for task %s: %s", task.TaskID, err.Error()))
		return
	}

	taskAdjustTokenQuota(ctx, task, -quota)

	other := taskBillingOther(task)
	other["task_id"] = task.TaskID
	other["reason"] = reason
	model.RecordTaskBillingLog(model.RecordTaskBillingLogParams{
		UserId:     task.UserId,
		LogType:    model.LogTypeRefund,
		Content:    "",
		ChannelId:  task.ChannelId,
		ModelName:  taskModelName(task),
		Quota:      quota,
		QuotaDelta: -quota,
		TokenId:    task.PrivateData.TokenId,
		Group:      task.Group,
		Other:      other,
	})
	if !taskIsSubscription(task) {
		if err := model.InvalidateUserCache(task.UserId); err != nil {
			logger.LogWarn(ctx, fmt.Sprintf("RefundTaskQuota: invalidate user cache failed for task %s: %s", task.TaskID, err.Error()))
		}
	}
	task.Quota = 0
	task.UpdatedAt = updatedAt
}

func adjustFundingTx(tx *gorm.DB, task *model.Task, delta int, updatedAt int64) error {
	if delta == 0 {
		return nil
	}
	if taskIsSubscription(task) {
		var sub model.UserSubscription
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("id = ?", task.PrivateData.SubscriptionId).
			First(&sub).Error; err != nil {
			return err
		}
		newUsed := sub.AmountUsed + int64(delta)
		if newUsed < 0 {
			newUsed = 0
		}
		if sub.AmountTotal > 0 && newUsed > sub.AmountTotal {
			return fmt.Errorf("subscription used exceeds total, used=%d total=%d", newUsed, sub.AmountTotal)
		}
		return tx.Model(&model.UserSubscription{}).
			Where("id = ?", task.PrivateData.SubscriptionId).
			Updates(map[string]any{"amount_used": newUsed, "updated_at": updatedAt}).Error
	}
	if delta > 0 {
		return tx.Model(&model.User{}).
			Where("id = ?", task.UserId).
			Update("quota", gorm.Expr("quota - ?", delta)).Error
	}
	return tx.Model(&model.User{}).
		Where("id = ?", task.UserId).
		Update("quota", gorm.Expr("quota + ?", -delta)).Error
}

func adjustUsageStatsTx(tx *gorm.DB, task *model.Task, quotaDelta int) error {
	if quotaDelta == 0 {
		return nil
	}
	if err := tx.Model(&model.User{}).Where("id = ?", task.UserId).
		Update("used_quota", gorm.Expr("used_quota + ?", quotaDelta)).Error; err != nil {
		return err
	}
	if task.ChannelId > 0 {
		if err := tx.Model(&model.Channel{}).Where("id = ?", task.ChannelId).
			Update("used_quota", gorm.Expr("used_quota + ?", quotaDelta)).Error; err != nil {
			return err
		}
	}
	return nil
}

func refundTaskQuotaInTransaction(task *model.Task, refundedQuota int, updatedAt int64) error {
	return model.DB.Transaction(func(tx *gorm.DB) error {
		persistResult := tx.Model(&model.Task{}).
			Where("id = ? AND quota = ?", task.ID, refundedQuota).
			Updates(map[string]any{"quota": 0, "updated_at": updatedAt})
		if persistResult.Error != nil {
			return persistResult.Error
		}
		if persistResult.RowsAffected == 0 {
			return errRefundAlreadyApplied
		}

		if err := adjustFundingTx(tx, task, -refundedQuota, updatedAt); err != nil {
			return err
		}
		return adjustUsageStatsTx(tx, task, -refundedQuota)
	})
}

func settleTaskQuotaInTransaction(task *model.Task, preConsumedQuota int, actualQuota int, quotaDelta int, updatedAt int64) error {
	return model.DB.Transaction(func(tx *gorm.DB) error {
		persistResult := tx.Model(&model.Task{}).
			Where("id = ? AND quota = ?", task.ID, preConsumedQuota).
			Updates(map[string]any{"quota": actualQuota, "updated_at": updatedAt})
		if persistResult.Error != nil {
			return persistResult.Error
		}
		if persistResult.RowsAffected == 0 {
			return errSettleAlreadyApplied
		}

		if err := adjustFundingTx(tx, task, quotaDelta, updatedAt); err != nil {
			return err
		}
		if quotaDelta != 0 {
			// NOTE: settle deliberately bypasses common.BatchUpdateEnabled.
			// Async task settlement is low-frequency and these usage counters
			// must commit atomically with tasks.quota and wallet/subscription
			// accounting; routing through batch helpers would defer the writes
			// and break that guarantee.
			//
			// used_quota is updated for both positive and negative deltas so
			// that the statistic stays consistent with the final settled quota.
			//
			// NOTE: request_count is intentionally not adjusted in settlement.
			// For async tasks, the request was already counted at submit time via
			// LogTaskConsumption → UpdateUserUsedQuotaAndRequestCount; settle only
			// reconciles used_quota to the actual cost.
			if err := adjustUsageStatsTx(tx, task, quotaDelta); err != nil {
				return err
			}
		}
		return nil
	})
}

// RecalculateTaskQuota 通用的异步差额结算。
// actualQuota 是任务完成后的实际应扣额度，与预扣额度 (task.Quota) 做差额结算。
// reason 用于日志记录（例如 "token重算" 或 "adaptor调整"）。
func RecalculateTaskQuota(ctx context.Context, task *model.Task, actualQuota int, reason string) {
	recalculateTaskQuotaWithTokenUsage(ctx, task, actualQuota, reason, 0, 0)
}

func recalculateTaskQuotaWithTokenUsage(ctx context.Context, task *model.Task, actualQuota int, reason string, promptTokens int, completionTokens int) {
	if actualQuota <= 0 {
		return
	}

	// Flush pending batch updates so the pre-charge is already applied to the
	// user's quota field before the settlement transaction reads it.
	model.FlushBatchUpdates()

	preConsumedQuota := task.Quota
	quotaDelta := actualQuota - preConsumedQuota

	if quotaDelta == 0 {
		logger.LogInfo(ctx, fmt.Sprintf("任务 %s 预扣费准确（%s，%s）",
			task.TaskID, logger.LogQuota(actualQuota), reason))
		return
	}

	logger.LogInfo(ctx, fmt.Sprintf("任务 %s 差额结算：delta=%s（实际：%s，预扣：%s，%s）",
		task.TaskID,
		logger.LogQuota(quotaDelta),
		logger.LogQuota(actualQuota),
		logger.LogQuota(preConsumedQuota),
		reason,
	))

	updatedAt := common.GetTimestamp()
	logType := model.LogTypeRefund
	logQuota := -quotaDelta
	if quotaDelta > 0 {
		logType = model.LogTypeConsume
		logQuota = quotaDelta
	}
	other := taskBillingOther(task)
	other["task_id"] = task.TaskID
	other["pre_consumed_quota"] = preConsumedQuota
	other["actual_quota"] = actualQuota
	logParams := model.RecordTaskBillingLogParams{
		UserId:           task.UserId,
		LogType:          logType,
		Content:          reason,
		ChannelId:        task.ChannelId,
		ModelName:        taskModelName(task),
		Quota:            logQuota,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		QuotaDelta:       quotaDelta,
		TokenId:          task.PrivateData.TokenId,
		Group:            task.Group,
		Other:            other,
	}

	// Persist the task row, funding adjustment, and usage stats in one DB
	// transaction. Billing logs intentionally stay best-effort (matching
	// model.RecordTaskBillingLog) and are written only after the accounting
	// transaction commits, so a transient log-table issue cannot block
	// refunds/settles.
	if err := settleTaskQuotaInTransaction(task, preConsumedQuota, actualQuota, quotaDelta, updatedAt); err != nil {
		if errors.Is(err, errSettleAlreadyApplied) {
			// CAS guard: another worker already settled this task; this is an
			// expected idempotent skip, not a real failure.
			logger.LogInfo(ctx, fmt.Sprintf("RecalculateTaskQuota: settle skipped (already applied) for task %s", task.TaskID))
			return
		}
		logger.LogError(ctx, fmt.Sprintf("RecalculateTaskQuota: settle transaction failed for task %s, aborting settle: %s", task.TaskID, err.Error()))
		return
	}
	model.RecordTaskBillingLog(logParams)
	if !taskIsSubscription(task) {
		// Wallet settlements update users.quota through the transaction instead
		// of model.Increase/DecreaseUserQuota, so invalidate the Redis user cache
		// after commit to avoid serving a stale quota value.
		if err := model.InvalidateUserCache(task.UserId); err != nil {
			logger.LogWarn(ctx, fmt.Sprintf("RecalculateTaskQuota: invalidate user cache failed for task %s: %s", task.TaskID, err.Error()))
		}
	}

	task.Quota = actualQuota
	task.UpdatedAt = updatedAt

	// Token quota/cache adjustment stays best-effort and outside the DB
	// transaction because token cache updates are intentionally asynchronous.
	taskAdjustTokenQuota(ctx, task, quotaDelta)
}

// RecalculateTaskQuotaByTokens 根据实际 token 消耗重新计费（异步差额结算）。
// 当任务成功且返回了 totalTokens 时，根据模型倍率和分组倍率重新计算实际扣费额度，
// 与预扣费的差额进行补扣或退还。支持钱包和订阅计费来源。
//
// 注意：tiered_expr 模型的结算由 adaptor.AdjustBillingOnComplete 处理
// （在 task_polling.go:SettleTaskBillingOnComplete 中优先调用），
// 此函数仅作为倍率计费路径的回退。
func RecalculateTaskQuotaByTokens(ctx context.Context, task *model.Task, totalTokens int) {
	RecalculateTaskQuotaByTokensWithUsage(ctx, task, totalTokens, 0, totalTokens)
}

// RecalculateTaskQuotaByTaskResult 按任务结果中的 token 结算，并将 token
// 写入第二段任务日志，用于 token 统计与看板展示。
func RecalculateTaskQuotaByTaskResult(ctx context.Context, task *model.Task, taskResult *relaycommon.TaskInfo) {
	totalTokens := effectiveTokenCount(taskResult)
	if totalTokens <= 0 {
		return
	}
	promptTokens, completionTokens := taskSettleLogTokens(taskResult, totalTokens)
	if completionTokens == 0 {
		completionTokens = totalTokens
	}
	RecalculateTaskQuotaByTokensWithUsage(ctx, task, totalTokens, promptTokens, completionTokens)
}

// RecalculateTaskQuotaByTokensWithUsage applies token-based quota recalculation
// with explicit prompt/completion token fields for settlement logs.
func RecalculateTaskQuotaByTokensWithUsage(ctx context.Context, task *model.Task, totalTokens int, promptTokens int, completionTokens int) {
	if totalTokens <= 0 {
		return
	}

	modelName := taskModelName(task)

	// 获取模型价格和倍率
	modelRatio, hasRatioSetting, _ := ratio_setting.GetModelRatio(modelName)
	// 只有配置了倍率(非固定价格)时才按 token 重新计费
	if !hasRatioSetting || modelRatio <= 0 {
		return
	}

	// 获取用户和组的倍率信息
	group := task.Group
	if group == "" {
		user, err := model.GetUserById(task.UserId, false)
		if err == nil {
			group = user.Group
		}
	}
	if group == "" {
		return
	}

	groupRatio := ratio_setting.GetGroupRatio(group)
	userGroupRatio, hasUserGroupRatio := ratio_setting.GetGroupGroupRatio(group, group)

	var finalGroupRatio float64
	if hasUserGroupRatio {
		finalGroupRatio = userGroupRatio
	} else {
		finalGroupRatio = groupRatio
	}

	// 计算 OtherRatios 乘积（视频折扣、时长等）
	otherMultiplier := 1.0
	if bc := task.PrivateData.BillingContext; bc != nil {
		for _, r := range bc.OtherRatios {
			if r != 1.0 && r > 0 {
				otherMultiplier *= r
			}
		}
	}

	// 计算实际应扣费额度: totalTokens * modelRatio * groupRatio * otherMultiplier
	actualQuota := int(float64(totalTokens) * modelRatio * finalGroupRatio * otherMultiplier)

	reason := fmt.Sprintf("token重算：tokens=%d, modelRatio=%.2f, groupRatio=%.2f, otherMultiplier=%.4f", totalTokens, modelRatio, finalGroupRatio, otherMultiplier)
	recalculateTaskQuotaWithTokenUsage(ctx, task, actualQuota, reason, promptTokens, completionTokens)
}
