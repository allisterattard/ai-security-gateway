package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func Send(w http.ResponseWriter, r *http.Request, redactedCount int, payload map[string]any, bodyBytes []byte, targetUrl string) {
	target, err := url.Parse(targetUrl)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	var finalBody []byte
	proxy := httputil.NewSingleHostReverseProxy(target)
	if redactedCount > 0 {
		finalBody, err = json.Marshal(payload)
		if err != nil {
			http.Error(w, "Failed to marshal payload", http.StatusInternalServerError)
			return
		}
		slog.Warn("[SECURITY REDACTION] Redacted", "count", redactedCount)
	} else {
		finalBody = bodyBytes
	}
	r.Body = io.NopCloser(bytes.NewReader(finalBody))
	r.ContentLength = int64(len(finalBody))

	r.Host = target.Host

	proxy.ServeHTTP(w, r)
}
