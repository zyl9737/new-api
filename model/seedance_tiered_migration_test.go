package model

import (
	"strings"
	"testing"
)

func TestBuildLegacySeedanceTieredPricingMigration_FromLegacyRatios(t *testing.T) {
	modelRatioRaw := `{"doubao-seedance-2-0-260128":23,"doubao-seedance-2-0-fast-260128":18.5}`
	billingModeRaw := `{}`
	billingExprRaw := `{}`

	nextModelRatio, nextBillingMode, nextBillingExpr, changed, err := buildLegacySeedanceTieredPricingMigration(
		modelRatioRaw,
		billingModeRaw,
		billingExprRaw,
	)
	if err != nil {
		t.Fatalf("buildLegacySeedanceTieredPricingMigration: %v", err)
	}
	if !changed {
		t.Fatal("expected migration to change legacy seedance settings")
	}
	if !strings.Contains(nextModelRatio, `"doubao-seedance-2-0-260128":46`) {
		t.Fatalf("expected sd2 ratio to migrate to 46, got %s", nextModelRatio)
	}
	if !strings.Contains(nextModelRatio, `"doubao-seedance-2-0-fast-260128":37`) {
		t.Fatalf("expected sd2-fast ratio to migrate to 37, got %s", nextModelRatio)
	}
	if !strings.Contains(nextBillingMode, `"doubao-seedance-2-0-260128":"tiered_expr"`) {
		t.Fatalf("expected sd2 billing mode to become tiered_expr, got %s", nextBillingMode)
	}
	if !strings.Contains(nextBillingMode, `"doubao-seedance-2-0-fast-260128":"tiered_expr"`) {
		t.Fatalf("expected sd2-fast billing mode to become tiered_expr, got %s", nextBillingMode)
	}
	if !strings.Contains(nextBillingExpr, `c * 46`) {
		t.Fatalf("expected sd2 billing expr to be injected, got %s", nextBillingExpr)
	}
	if !strings.Contains(nextBillingExpr, `c * 37`) {
		t.Fatalf("expected sd2-fast billing expr to be injected, got %s", nextBillingExpr)
	}
}

func TestBuildLegacySeedanceTieredPricingMigration_FromBaseRatioStillSelectsTemplate(t *testing.T) {
	modelRatioRaw := `{"doubao-seedance-2-0-260128":46}`
	billingModeRaw := `{}`
	billingExprRaw := `{}`

	_, nextBillingMode, nextBillingExpr, changed, err := buildLegacySeedanceTieredPricingMigration(
		modelRatioRaw,
		billingModeRaw,
		billingExprRaw,
	)
	if err != nil {
		t.Fatalf("buildLegacySeedanceTieredPricingMigration: %v", err)
	}
	if !changed {
		t.Fatal("expected migration to select classic tiered template")
	}
	if !strings.Contains(nextBillingMode, `"doubao-seedance-2-0-260128":"tiered_expr"`) {
		t.Fatalf("expected sd2 billing mode to become tiered_expr, got %s", nextBillingMode)
	}
	if !strings.Contains(nextBillingExpr, `c * 46`) {
		t.Fatalf("expected sd2 billing expr to be injected, got %s", nextBillingExpr)
	}
}

func TestBuildLegacySeedanceTieredPricingMigration_LeavesCustomRatioUntouched(t *testing.T) {
	modelRatioRaw := `{"doubao-seedance-2-0-260128":50}`
	billingModeRaw := `{}`
	billingExprRaw := `{}`

	_, _, _, changed, err := buildLegacySeedanceTieredPricingMigration(
		modelRatioRaw,
		billingModeRaw,
		billingExprRaw,
	)
	if err != nil {
		t.Fatalf("buildLegacySeedanceTieredPricingMigration: %v", err)
	}
	if changed {
		t.Fatal("expected custom ratio configuration to remain untouched")
	}
}
