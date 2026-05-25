package openai

import (
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

func TestRecordFinishReasonFromStreamItemsChat(t *testing.T) {
	info := &relaycommon.RelayInfo{}
	streamItems := []string{
		`{"choices":[{"delta":{"content":"hello"},"finish_reason":null,"index":0}]}`,
		`{"choices":[{"delta":{},"finish_reason":"length","index":0}]}`,
		`{"choices":[],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`,
	}

	recordFinishReasonFromStreamItems(info, streamItems)
	if info.FinishReason != "length" {
		t.Fatalf("expected finish reason length, got %q", info.FinishReason)
	}
}

func TestRecordFinishReasonFromStreamItemsCompletions(t *testing.T) {
	info := &relaycommon.RelayInfo{}
	streamItems := []string{
		`{"choices":[{"text":"hello","finish_reason":""}]}`,
		`{"choices":[{"text":"","finish_reason":"stop"}]}`,
	}

	recordFinishReasonFromStreamItems(info, streamItems)
	if info.FinishReason != "stop" {
		t.Fatalf("expected finish reason stop, got %q", info.FinishReason)
	}
}

func TestRecordFinishReasonFromStreamItemsKeepExisting(t *testing.T) {
	info := &relaycommon.RelayInfo{FinishReason: "content_filter"}
	streamItems := []string{
		`{"choices":[{"delta":{},"finish_reason":"length","index":0}]}`,
	}

	recordFinishReasonFromStreamItems(info, streamItems)
	if info.FinishReason != "content_filter" {
		t.Fatalf("expected existing finish reason to be kept, got %q", info.FinishReason)
	}
}
