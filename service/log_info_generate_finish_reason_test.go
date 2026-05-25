package service

import (
	"net/http/httptest"
	"testing"
	"time"

	relaycommon "github.com/QuantumNous/new-api/relay/common"

	"github.com/gin-gonic/gin"
)

func TestGenerateTextOtherInfoIncludesFinishReason(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	now := time.Now()
	info := &relaycommon.RelayInfo{
		StartTime:         now,
		FirstResponseTime: now.Add(10 * time.Millisecond),
		FinishReason:      "max_tokens",
		ChannelMeta:       &relaycommon.ChannelMeta{},
	}

	other := GenerateTextOtherInfo(ctx, info, 1, 1, 1, 0, 0, 0, 1)
	got, ok := other["finish_reason"].(string)
	if !ok {
		t.Fatalf("expected finish_reason field in other info")
	}
	if got != "max_tokens" {
		t.Fatalf("expected finish_reason max_tokens, got %q", got)
	}
}
