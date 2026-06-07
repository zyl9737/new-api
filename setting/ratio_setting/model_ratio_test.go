package ratio_setting

import "testing"

func TestUpdateModelRatioByJSONStringNormalizesLegacySeedanceHalfPrices(t *testing.T) {
	original := ModelRatio2JSONString()
	t.Cleanup(func() {
		if err := UpdateModelRatioByJSONString(original); err != nil {
			t.Fatalf("restore model ratios: %v", err)
		}
	})

	err := UpdateModelRatioByJSONString(`{
		"doubao-seedance-2-0-260128": 23,
		"doubao-seedance-2-0-fast-260128": 18.5
	}`)
	if err != nil {
		t.Fatalf("UpdateModelRatioByJSONString: %v", err)
	}

	if got, ok, _ := GetModelRatio("doubao-seedance-2-0-260128"); !ok || got != 46 {
		t.Fatalf("doubao-seedance-2-0-260128 ratio = %v, ok=%v, want 46", got, ok)
	}
	if got, ok, _ := GetModelRatio("doubao-seedance-2-0-fast-260128"); !ok || got != 37 {
		t.Fatalf("doubao-seedance-2-0-fast-260128 ratio = %v, ok=%v, want 37", got, ok)
	}
}
