package model

import "regexp"

type SecurityRules struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Pattern       string `json:"pattern"`
	Action        string `json:"action"`
	Enabled       bool   `json:"enabled"`
	CompiledRegex *regexp.Regexp
}

type Config struct {
	SecurityRules []SecurityRules `json:"security_rules"`
}

var GlobalConfig Config

var ClaudeEndpoint string
var OpenAiEndpoint string
