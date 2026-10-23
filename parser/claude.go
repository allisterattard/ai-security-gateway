package parser

import (
	"ai-security-gateway/model"
	"ai-security-gateway/proxy"
	"ai-security-gateway/securitycheck"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func HandleClaudeRequest(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = "/v1/messages"

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var payload map[string]any
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		proxy.Send(w, r, 0, payload, bodyBytes, model.ClaudeEndpoint)
		return
	}

	redactedCount := 0

	if messages, ok := payload["messages"].([]any); ok {
		for _, msg := range messages {
			if msgMap, ok := msg.(map[string]any); ok {
				roleStr := "unknown"
				if role, ok := msgMap["role"].(string); ok {
					roleStr = role
				}
				if content, ok := msgMap["content"].(string); ok {
					content, redactedCount = securitycheck.RegexCheck(content, roleStr, redactedCount)
					msgMap["content"] = content
				} else if contentBlocks, ok := msgMap["content"].([]any); ok {
					for _, block := range contentBlocks {
						if blockMap, ok := block.(map[string]any); ok {
							if text, ok := blockMap["text"].(string); ok {
								text, redactedCount = securitycheck.RegexCheck(text, roleStr, redactedCount)
								blockMap["text"] = text
							}
						}
					}
				}
			}
		}
	}
	if sys, exists := payload["system"]; exists {
		if sysStr, ok := sys.(string); ok {
			sysStr, redactedCount = securitycheck.RegexCheck(sysStr, "system", redactedCount)
			payload["system"] = sysStr
		} else if sysBlocks, ok := sys.([]any); ok {
			for _, block := range sysBlocks {
				if blockMap, ok := block.(map[string]any); ok {
					if text, ok := blockMap["text"].(string); ok {
						text, redactedCount = securitycheck.RegexCheck(text, "system", redactedCount)
						blockMap["text"] = text
					}
				}
			}
		}
	}

	proxy.Send(w, r, redactedCount, payload, bodyBytes, model.ClaudeEndpoint)
}
