package mediakit

import (
	"bytes"
	"math"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	channelmediakit "github.com/QuantumNous/new-api/relay/channel/mediakit"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
)

func TestAdjustBillingOnComplete_EnhanceProfessional(t *testing.T) {
	task := &model.Task{
		Properties: model.Properties{OriginModelName: channelmediakit.ModelEnhanceVideoProfessional},
		PrivateData: model.TaskPrivateData{
			BillingContext: &model.TaskBillingContext{
				ModelPrice:      0.75,
				GroupRatio:      1,
				OriginModelName: channelmediakit.ModelEnhanceVideoProfessional,
				MediaKit: &model.TaskMediaKitParams{
					Tool:        actionEnhanceVideo,
					ToolVersion: "professional",
				},
			},
		},
		Data: []byte(`{
			"success": true,
			"task_id": "amk-task-upstream",
			"status": "success",
			"output": {
				"video_url": "https://example.com/output.mp4",
				"duration": 120,
				"resolution": "1080p",
				"fps": 60,
				"tool_version": "professional"
			}
		}`),
	}

	a := &TaskAdaptor{}
	got := a.AdjustBillingOnComplete(task, nil)
	want := int(math.Round(0.75 * 40 * 2 * common.QuotaPerUnit))
	if got != want {
		t.Fatalf("AdjustBillingOnComplete enhance = %d, want %d", got, want)
	}

	bc := task.PrivateData.BillingContext
	if bc == nil {
		t.Fatal("BillingContext should not be nil")
	}
	if bc.OtherRatios["billing_coefficient"] != 40 {
		t.Fatalf("billing_coefficient = %v, want 40", bc.OtherRatios["billing_coefficient"])
	}
	if !strings.Contains(bc.SettlementReason, "toolVersion=professional") {
		t.Fatalf("SettlementReason = %q, want professional detail", bc.SettlementReason)
	}
}

func TestAdjustBillingOnComplete_SubtitleStandard(t *testing.T) {
	task := &model.Task{
		Properties: model.Properties{OriginModelName: channelmediakit.ModelEraseVideoSubtitleStandard},
		PrivateData: model.TaskPrivateData{
			BillingContext: &model.TaskBillingContext{
				ModelPrice:      1,
				GroupRatio:      1,
				OriginModelName: channelmediakit.ModelEraseVideoSubtitleStandard,
				MediaKit: &model.TaskMediaKitParams{
					Tool: actionEraseSubtitle,
				},
			},
		},
		Data: []byte(`{
			"success": true,
			"task_id": "amk-task-upstream",
			"status": "success",
			"result": {
				"video_url": "https://example.com/output.mp4",
				"duration": 120
			}
		}`),
	}

	a := &TaskAdaptor{}
	got := a.AdjustBillingOnComplete(task, nil)
	want := int(math.Round(1 * 0.4 * 2 * common.QuotaPerUnit))
	if got != want {
		t.Fatalf("AdjustBillingOnComplete subtitle standard = %d, want %d", got, want)
	}
}

func TestParseTaskResult_NestedDataSuccess(t *testing.T) {
	a := &TaskAdaptor{}
	ti, err := a.ParseTaskResult([]byte(`{
		"success": true,
		"data": {
			"task_id": "amk-task-upstream",
			"status": "completed",
			"result": {
				"video_url": "https://example.com/output.mp4",
				"duration": 88.8
			}
		}
	}`))
	if err != nil {
		t.Fatalf("ParseTaskResult error: %v", err)
	}
	if ti.Status != model.TaskStatusSuccess {
		t.Fatalf("Status = %q, want %q", ti.Status, model.TaskStatusSuccess)
	}
	if ti.Url != "https://example.com/output.mp4" {
		t.Fatalf("Url = %q, want output URL", ti.Url)
	}
}

func TestEstimateBilling_EnhanceStandardUsesOnlyChargeMultipliers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"video_url":"https://example.com/video.mp4","resolution":"720p"}`)
	req := httptest.NewRequest("POST", channelmediakit.RouteEnhanceVideo, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	info := &relaycommon.RelayInfo{
		RequestURLPath:  channelmediakit.RouteEnhanceVideo,
		OriginModelName: channelmediakit.ModelEnhanceVideoStandard,
	}

	a := &TaskAdaptor{}
	ratios := a.EstimateBilling(c, info)
	if ratios["duration_minutes"] != 1 {
		t.Fatalf("duration_minutes = %v, want 1", ratios["duration_minutes"])
	}
	if ratios["billing_coefficient"] != 1 {
		t.Fatalf("billing_coefficient = %v, want 1", ratios["billing_coefficient"])
	}
	if _, ok := ratios["resolution_bucket"]; ok {
		t.Fatalf("resolution_bucket should be informational only and not used as a pre-charge multiplier: %v", ratios)
	}
	if _, ok := ratios["output_fps"]; ok {
		t.Fatalf("output_fps should be informational only and not used as a pre-charge multiplier: %v", ratios)
	}
}
