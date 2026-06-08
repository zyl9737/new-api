package model

import (
	"strings"
	"testing"
)

func TestBuildMediaKitModelPriceMigration_AddsMissingDefaults(t *testing.T) {
	raw := `{"dall-e-3":0.04,"volc-mediakit-enhance-video-standard":0.8}`
	next, changed, err := buildMediaKitModelPriceMigration(raw)
	if err != nil {
		t.Fatalf("buildMediaKitModelPriceMigration error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true when MediaKit model prices are missing")
	}
	for _, key := range []string{
		`"volc-mediakit-enhance-video-standard":0.8`,
		`"volc-mediakit-enhance-video-professional":0.75`,
		`"volc-mediakit-erase-video-subtitle-standard":1`,
		`"volc-mediakit-erase-video-subtitle-pro":1`,
	} {
		if !strings.Contains(next, key) {
			t.Fatalf("migrated ModelPrice missing %s: %s", key, next)
		}
	}
}

func TestBuildMediaKitModelPriceMigration_NoChangeWhenAllPresent(t *testing.T) {
	raw := `{"volc-mediakit-enhance-video-standard":0.75,"volc-mediakit-enhance-video-professional":0.75,"volc-mediakit-erase-video-subtitle-standard":1,"volc-mediakit-erase-video-subtitle-pro":1}`
	next, changed, err := buildMediaKitModelPriceMigration(raw)
	if err != nil {
		t.Fatalf("buildMediaKitModelPriceMigration error: %v", err)
	}
	if changed {
		t.Fatalf("expected changed=false, got next=%s", next)
	}
	if next != raw {
		t.Fatalf("expected raw string to be preserved, got %s", next)
	}
}
