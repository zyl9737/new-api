package controller

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
)

type volcTaskPromptPayload struct {
	Prompt  string `json:"prompt"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

type rawTaskPromptPayload struct {
	Prompt string `json:"prompt"`
}

func extractTaskPromptInput(c *gin.Context, relayInfo *relaycommon.RelayInfo) string {
	if prompt := extractTaskRequestPrompt(c); prompt != "" {
		return prompt
	}
	if relayInfo != nil && relayInfo.ChannelType == constant.ChannelTypeVolcAdapter {
		return extractVolcTaskPrompt(c)
	}
	return extractRawTaskPrompt(c)
}

func extractTaskRequestPrompt(c *gin.Context) string {
	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(req.Prompt)
}

func extractRawTaskPrompt(c *gin.Context) string {
	var payload rawTaskPromptPayload
	if err := common.UnmarshalBodyReusable(c, &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.Prompt)
}

func extractVolcTaskPrompt(c *gin.Context) string {
	var payload volcTaskPromptPayload
	if err := common.UnmarshalBodyReusable(c, &payload); err != nil {
		return ""
	}
	if prompt := strings.TrimSpace(payload.Prompt); prompt != "" {
		return prompt
	}

	parts := make([]string, 0, len(payload.Content))
	for _, item := range payload.Content {
		if item.Type != "" && item.Type != "text" {
			continue
		}
		if text := strings.TrimSpace(item.Text); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n\n")
}
