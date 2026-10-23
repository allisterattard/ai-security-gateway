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

func HandleOpenAiRequest(w http.ResponseWriter, r *http.Request) {

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var payload map[string]any
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		proxy.Send(w, r, 0, payload, bodyBytes, model.OpenAiEndpoint)
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
				}
			}
		}
	}

	proxy.Send(w, r, redactedCount, payload, bodyBytes, model.OpenAiEndpoint)
}
