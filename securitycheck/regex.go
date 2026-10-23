package securitycheck

import (
	"ai-security-gateway/model"
	"log/slog"
)

func RegexCheck(content string, roleNme string, redactedCount int) (string, int) {
	for _, rule := range model.GlobalConfig.SecurityRules {
		if rule.Enabled && rule.CompiledRegex != nil {
			if rule.CompiledRegex.MatchString(content) {
				content = rule.CompiledRegex.ReplaceAllString(content, "[REDACTED-SECURITY-POLICY]")
				redactedCount++
				slog.Warn("[SECURITY REDACTION]", "rule", rule.Name, "role", roleNme)
			}
		}
	}
	return content, redactedCount
}
