package controller

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
)

func TestBuildTaskQueryParamsReadsTokenName(t *testing.T) {
	t.Setenv("GIN_MODE", "test")
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("GET", "/api/task?task_id=task_1&token_name=tk-1&channel_id=58&start_timestamp=10&end_timestamp=20", nil)

	params := buildTaskQueryParams(ctx)
	if params.TaskID != "task_1" {
		t.Fatalf("task_id = %q, want task_1", params.TaskID)
	}
	if params.TokenName != "tk-1" {
		t.Fatalf("token_name = %q, want tk-1", params.TokenName)
	}
	if params.ChannelID != "58" {
		t.Fatalf("channel_id = %q, want 58", params.ChannelID)
	}
	if params.StartTimestamp != 10 || params.EndTimestamp != 20 {
		t.Fatalf("unexpected timestamps: %+v", params)
	}
}

func TestExtractTaskPromptInputUsesStoredTaskRequest(t *testing.T) {
	t.Setenv("GIN_MODE", "test")
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("task_request", relaycommon.TaskSubmitReq{Prompt: "stored prompt"})

	got := extractTaskPromptInput(ctx, &relaycommon.RelayInfo{})
	if got != "stored prompt" {
		t.Fatalf("prompt = %q, want stored prompt", got)
	}
}

func TestExtractTaskPromptInputHandlesVolcContentText(t *testing.T) {
	t.Setenv("GIN_MODE", "test")
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"doubao-seedance-2-0","content":[{"type":"text","text":"第一段提示词"},{"type":"image_url","image_url":{"url":"https://example.com/image.png"}},{"type":"text","text":"第二段提示词"}]}`)
	req := httptest.NewRequest("POST", "/v1/video/generations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req

	got := extractTaskPromptInput(ctx, &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{ChannelType: constant.ChannelTypeVolcAdapter},
	})
	if got != "第一段提示词\n\n第二段提示词" {
		t.Fatalf("prompt = %q", got)
	}
}

func TestTasksToDtoStripsPromptForNonRoot(t *testing.T) {
	tasks := []*model.Task{
		{
			TaskID:    "task_1",
			TokenName: "tk-1",
			Properties: model.Properties{
				Input:             "secret prompt",
				OriginModelName:   "seedance",
				UpstreamModelName: "seedance-upstream",
			},
		},
	}

	items := tasksToDto(tasks, false, false)
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].TokenName != "tk-1" {
		t.Fatalf("token_name = %q, want tk-1", items[0].TokenName)
	}

	properties, ok := items[0].Properties.(model.Properties)
	if !ok {
		t.Fatalf("properties type = %T, want model.Properties", items[0].Properties)
	}
	if properties.Input != "" {
		t.Fatalf("prompt should be stripped, got %q", properties.Input)
	}
	if properties.OriginModelName != "seedance" {
		t.Fatalf("origin_model_name = %q, want seedance", properties.OriginModelName)
	}
}
