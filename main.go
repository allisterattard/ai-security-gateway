package main

import (
	"ai-security-gateway/model"
	"ai-security-gateway/parser"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"regexp"
)

func init() {

	model.OpenAiEndpoint = os.Getenv("OPENAI_ENDPOINT")
	model.ClaudeEndpoint = os.Getenv("CLAUDE_ENDPOINT")

	if model.OpenAiEndpoint == "" && model.ClaudeEndpoint == "" {
		slog.Error("OPENAI_ENDPOINT or CLAUDE_ENDPOINT environment variable is not set")
		os.Exit(1)
	}

	jsonFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()
	byteValiue, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(byteValiue, &model.GlobalConfig)
	if err != nil {
		log.Fatal(err)
	}

	for i := range model.GlobalConfig.SecurityRules {
		rule := &model.GlobalConfig.SecurityRules[i]
		if rule.Enabled {
			re, err := regexp.Compile(rule.Pattern)
			if err != nil {
				log.Fatal(err)
			}
			rule.CompiledRegex = re
		}
	}
	slog.Info("Loaded security rules into memory successfully!", "count", len(model.GlobalConfig.SecurityRules))
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/chat/completions", parser.HandleOpenAiRequest)
	mux.HandleFunc("/v1/messages", parser.HandleClaudeRequest)
	mux.HandleFunc("/messages", parser.HandleClaudeRequest)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"ai-security-gateway"}`))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // sensible default fallback
	}
	// Ensure it has a colon prefix
	if port[0] != ':' {
		port = ":" + port
	}
	loggedMux := RequestLogger(mux)

	slog.Info("Starting proxy server", "port", port)
	log.Fatal(http.ListenAndServe(port, loggedMux))
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("incoming request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
		// Pass the request to the actual router/handler
		next.ServeHTTP(w, r)
	})
}
