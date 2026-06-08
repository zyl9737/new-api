package model

import (
	"fmt"

	"github.com/QuantumNous/new-api/common"
)

var mediaKitDefaultModelPrices = map[string]float64{
	"volc-mediakit-enhance-video-standard":        0.75,
	"volc-mediakit-enhance-video-professional":    0.75,
	"volc-mediakit-erase-video-subtitle-standard": 1,
	"volc-mediakit-erase-video-subtitle-pro":      1,
}

func migrateMediaKitModelPriceDefaults() {
	common.OptionMapRWMutex.RLock()
	modelPriceRaw := common.OptionMap["ModelPrice"]
	common.OptionMapRWMutex.RUnlock()

	nextModelPrice, changed, err := buildMediaKitModelPriceMigration(modelPriceRaw)
	if err != nil {
		common.SysLog("mediakit model price migration skipped: " + err.Error())
		return
	}
	if !changed {
		return
	}
	if err := UpdateOption("ModelPrice", nextModelPrice); err != nil {
		common.SysLog("failed to persist migrated MediaKit ModelPrice: " + err.Error())
		return
	}
	common.SysLog("migrated MediaKit default model prices into ModelPrice")
}

func buildMediaKitModelPriceMigration(modelPriceRaw string) (string, bool, error) {
	modelPrices := make(map[string]float64)
	if err := unmarshalJSONMap(modelPriceRaw, &modelPrices); err != nil {
		return "", false, fmt.Errorf("parse ModelPrice: %w", err)
	}

	changed := false
	for modelName, defaultPrice := range mediaKitDefaultModelPrices {
		if _, ok := modelPrices[modelName]; ok {
			continue
		}
		modelPrices[modelName] = defaultPrice
		changed = true
	}
	if !changed {
		return modelPriceRaw, false, nil
	}

	nextModelPrice, err := common.Marshal(modelPrices)
	if err != nil {
		return "", false, fmt.Errorf("marshal ModelPrice: %w", err)
	}
	return string(nextModelPrice), true, nil
}
