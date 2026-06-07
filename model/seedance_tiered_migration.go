package model

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/billing_setting"
)

type seedanceTieredPreset struct {
	baseRatio float64
	halfRatio float64
	expr      string
}

var legacySeedanceTieredPresets = map[string]seedanceTieredPreset{
	"doubao-seedance-2-0-260128": {
		baseRatio: 46,
		halfRatio: 23,
		expr:      `(tier("base", c * 46)) * (param("resolution") == "1080p" ? 1.108696 : 1) * (param("content.#(type==\"video_url\")") != nil ? 0.608696 : 1)`,
	},
	"doubao-seedance-2-0-fast-260128": {
		baseRatio: 37,
		halfRatio: 18.5,
		expr:      `(tier("base", c * 37)) * (param("content.#(type==\"video_url\")") != nil ? 0.594595 : 1)`,
	},
}

func migrateLegacySeedanceTieredPricing() {
	common.OptionMapRWMutex.RLock()
	modelRatioRaw := common.OptionMap["ModelRatio"]
	billingModeRaw := common.OptionMap["billing_setting.billing_mode"]
	billingExprRaw := common.OptionMap["billing_setting.billing_expr"]
	common.OptionMapRWMutex.RUnlock()

	nextModelRatio, nextBillingMode, nextBillingExpr, changed, err := buildLegacySeedanceTieredPricingMigration(
		modelRatioRaw,
		billingModeRaw,
		billingExprRaw,
	)
	if err != nil {
		common.SysLog("legacy seedance tiered migration skipped: " + err.Error())
		return
	}
	if !changed {
		return
	}

	if nextModelRatio != modelRatioRaw {
		if err := UpdateOption("ModelRatio", nextModelRatio); err != nil {
			common.SysLog("failed to persist migrated ModelRatio: " + err.Error())
			return
		}
	}
	if nextBillingMode != billingModeRaw {
		if err := UpdateOption("billing_setting.billing_mode", nextBillingMode); err != nil {
			common.SysLog("failed to persist migrated billing mode: " + err.Error())
			return
		}
	}
	if nextBillingExpr != billingExprRaw {
		if err := UpdateOption("billing_setting.billing_expr", nextBillingExpr); err != nil {
			common.SysLog("failed to persist migrated billing expr: " + err.Error())
			return
		}
	}

	common.SysLog("migrated legacy seedance pricing to classic tiered_expr presets")
}

func buildLegacySeedanceTieredPricingMigration(
	modelRatioRaw string,
	billingModeRaw string,
	billingExprRaw string,
) (string, string, string, bool, error) {
	modelRatios := make(map[string]float64)
	billingModes := make(map[string]string)
	billingExprs := make(map[string]string)

	if err := unmarshalJSONMap(modelRatioRaw, &modelRatios); err != nil {
		return "", "", "", false, fmt.Errorf("parse ModelRatio: %w", err)
	}
	if err := unmarshalJSONMap(billingModeRaw, &billingModes); err != nil {
		return "", "", "", false, fmt.Errorf("parse billing mode: %w", err)
	}
	if err := unmarshalJSONMap(billingExprRaw, &billingExprs); err != nil {
		return "", "", "", false, fmt.Errorf("parse billing expr: %w", err)
	}

	changed := false
	for modelName, preset := range legacySeedanceTieredPresets {
		currentRatio, hasRatio := modelRatios[modelName]
		currentMode := strings.TrimSpace(billingModes[modelName])
		currentExpr := strings.TrimSpace(billingExprs[modelName])

		shouldMigrate := false
		switch {
		case hasRatio && (currentRatio == preset.halfRatio || currentRatio == preset.baseRatio):
			shouldMigrate = true
		case currentMode == billing_setting.BillingModeTieredExpr:
			shouldMigrate = true
		case currentExpr != "":
			shouldMigrate = true
		}
		if !shouldMigrate {
			continue
		}

		if !hasRatio || currentRatio == preset.halfRatio {
			modelRatios[modelName] = preset.baseRatio
			changed = true
		}
		if currentMode != billing_setting.BillingModeTieredExpr {
			billingModes[modelName] = billing_setting.BillingModeTieredExpr
			changed = true
		}
		if currentExpr == "" {
			billingExprs[modelName] = preset.expr
			changed = true
		}
	}

	if !changed {
		return modelRatioRaw, billingModeRaw, billingExprRaw, false, nil
	}

	nextModelRatio, err := common.Marshal(modelRatios)
	if err != nil {
		return "", "", "", false, fmt.Errorf("marshal ModelRatio: %w", err)
	}
	nextBillingMode, err := common.Marshal(billingModes)
	if err != nil {
		return "", "", "", false, fmt.Errorf("marshal billing mode: %w", err)
	}
	nextBillingExpr, err := common.Marshal(billingExprs)
	if err != nil {
		return "", "", "", false, fmt.Errorf("marshal billing expr: %w", err)
	}

	return string(nextModelRatio), string(nextBillingMode), string(nextBillingExpr), true, nil
}

func unmarshalJSONMap[T any](raw string, target *T) error {
	if strings.TrimSpace(raw) == "" {
		raw = "{}"
	}
	return common.Unmarshal([]byte(raw), target)
}
