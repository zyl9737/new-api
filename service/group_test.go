package service

import (
	"encoding/json"
	"testing"

	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
)

func cloneGroupMap(input map[string]string) map[string]string {
	cloned := make(map[string]string, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}

func cloneSpecialGroupRules(input map[string]map[string]string) map[string]map[string]string {
	cloned := make(map[string]map[string]string, len(input))
	for userGroup, rules := range input {
		cloned[userGroup] = cloneGroupMap(rules)
	}
	return cloned
}

func TestGetUserUsableGroupsOnlyRules(t *testing.T) {
	originalUsableGroups := setting.GetUserUsableGroupsCopy()
	originalSpecialRules := ratio_setting.GetGroupRatioSetting().GroupSpecialUsableGroup.ReadAll()

	t.Cleanup(func() {
		originalUsableGroupsJSON, err := json.Marshal(cloneGroupMap(originalUsableGroups))
		if err != nil {
			t.Fatalf("failed to marshal original usable groups: %v", err)
		}
		if err := setting.UpdateUserUsableGroupsByJSONString(string(originalUsableGroupsJSON)); err != nil {
			t.Fatalf("failed to restore original usable groups: %v", err)
		}
		ratio_setting.GetGroupRatioSetting().GroupSpecialUsableGroup.Clear()
		ratio_setting.GetGroupRatioSetting().GroupSpecialUsableGroup.AddAll(
			cloneSpecialGroupRules(originalSpecialRules),
		)
	})

	usableGroupsJSON, err := json.Marshal(map[string]string{
		"default": "Default group",
		"vip":     "VIP group",
		"studio":  "Studio group",
		"preview": "Preview group",
	})
	if err != nil {
		t.Fatalf("failed to marshal usable groups: %v", err)
	}
	if err := setting.UpdateUserUsableGroupsByJSONString(string(usableGroupsJSON)); err != nil {
		t.Fatalf("failed to set usable groups: %v", err)
	}

	ratio_setting.GetGroupRatioSetting().GroupSpecialUsableGroup.Clear()
	ratio_setting.GetGroupRatioSetting().GroupSpecialUsableGroup.AddAll(
		map[string]map[string]string{
			"vip": {
				"=:studio":   "Studio only",
				"+:preview":  "Preview access",
				"-:default":  "remove",
				"temporary":  "Temporary group",
				"=:preview2": "Preview only",
			},
		},
	)

	usableGroups := GetUserUsableGroups("vip")

	if _, ok := usableGroups["default"]; ok {
		t.Fatalf("default group should be removed when only rules are active: %#v", usableGroups)
	}

	for _, group := range []string{"studio", "preview", "preview2", "temporary", "vip"} {
		if _, ok := usableGroups[group]; !ok {
			t.Fatalf("expected group %q to be available, got %#v", group, usableGroups)
		}
	}
}
