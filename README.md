# AI Security Gateway 

![Go](https://img.shields.io/badge/Go-1.27%2B-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green.svg)


A high-performance, plug-and-play **AI Security Gateway** written in **Go**. Designed to sit between developer tools (such as the **OpenCode CLI**) and upstream LLM providers (Anthropic Claude, OpenAI), it dynamically intercepts, inspects, and sanitizes outgoing requests in real-time to enforce data compliance, block PII leaks, and mitigate prompt injection attacks.

## Architecture & Core Features

- **Dual LLM Support:** Seamlessly handles request parsing and reverse-proxying for both Anthropic (`/v1/messages`) and OpenAI (`/v1/chat/completions`) endpoints.
- **Granular Regex Rule Engine:** Hot-reloads structured JSON rules from `config.json` on startup to pre-compile regular expressions for optimal routing throughput.
- **Intelligent Redaction:** Scans deep into multi-turn chat message bodies and system prompt instructions, replacing sensitive data (AWS keys, PEM credentials, credit cards, emails, DB strings) and jailbreak vectors with standardized `[REDACTED-SECURITY-POLICY]` markers.
- **Developer CLI Integration:** Fully compatible with custom AI agents and environments like **OpenCode**, acting as a secure intermediary layer.

## Project Structure

``` bash
ai-security-gateway/
├── model/           # Global structs and environment configurations
├── parser/          # Protocol-specific request deserializers and sanitizers
├── proxy/           # Reverse-proxy transport and payload rewriting logic
├── securitycheck/   # Regex matching and redaction policy execution
├── config.json      # Declarative security rules and patterns
├── config.schema.json # JSON Schema definition for configuration validation
└── main.go          # Server entrypoint, middleware, and route multiplexing
```

## Configuration (`config.json`)

```json
{
  "$schema": "./config.schema.json",
  "security_rules": [
    {
      "id": "RULE-001",
      "name": "AWS Access Key",
      "pattern": "\\bAKIA[0-9A-Z]{16}\\b",
      "enabled": true
    },
    {
      "id": "RULE-008",
      "name": "Jailbreak: Ignore Instructions",
      "pattern": "(?i)ignore\\s+(all\\s+)?previous\\s+(instructions|prompts|rules)",
      "enabled": true
    }
  ]
}
```

## Getting Started

**1. Prerequisites**

- Go `1.27` or higher installed on your machine.

**2. Environment Variables (Gateway Backend)**

Set your upstream provider target endpoints before launching the application:

```bash
export OPENAI_ENDPOINT="https://api.openai.com"
export CLAUDE_ENDPOINT="https://api.anthropic.com"
export PORT="8080" # Optional: defaults to 3000
```

**3. Run the Gateway**

``` bash
go run main.go
```

- Health Check: curl `http://localhost:8080/healthz`

## Building from Source

If you want to compile the binary locally for production or local testing:

1. **Clone the repository:**
   ```bash
   git clone [https://github.com/your-username/ai-security-gateway.git](https://github.com/your-username/ai-security-gateway.git)
   cd ai-security-gateway
   ```

2. **Download dependencies:**
   
    ``` bash
    go mod tidy
    ```

3. **Build the binary:**

    ``` bash
    go build -o ai-security-gateway main.go
    ```
4. **Run the compiled executable:**

``` bash
export OPENAI_ENDPOINT="https://api.openai.com"
export CLAUDE_ENDPOINT="https://api.anthropic.com"
export PORT="8080"
./ai-security-gateway
```

### ⚠️ Important Note on Configuration
Because the application loads security rules dynamically at startup, **`config.json`** (and its accompanying `config.schema.json`) **must be located in the exact directory from which you execute the binary or run `go run`**.

## Integrating with OpenCode CLI

This gateway was explicitly tested and verified with the OpenCode CLI to safeguard developer workflows.

1. Environment Variables for OpenCode
   Export your production API keys in your active terminal shell session:

    ``` bash
    export ANTHROPIC_API_KEY="sk-ant-..."
    ```

2. OpenCode Configuration (`opencode.json`)
   Place the following configuration file inside your project directory to route OpenCode traffic through your local security gateway:

    ```json
    {
      "$schema": "https://opencode.ai/config.json",
      "subagent_depth": 1,
      "providers": {
        "anthropic": {
          "settings": {
            "baseURL": "http://localhost:8080",
            "apiKey": "{env:ANTHROPIC_API_KEY}"
          },
          "models": {
            "claude-haiku-5-5": {
              "modelID": "claude-haiku-5-5",
              "name": "Claude Haiku 5.5 via Security Gateway"
            }
          }
        }
      },
      "model": "anthropic/claude-haiku-5-5",
      "agents": {
        "coder": {
          "description": "Create Python3 Scripts",
          "mode": "primary",
          "system": "Create Python3 Scripts",
          "permissions": [
            { "action": "edit", "resource": "*", "effect": "allow" }
          ]
        }
      }
    }
    ```

## Testing the Security Filter
When a developer queries the model through OpenCode containing a blacklisted pattern (e.g., an AWS key or a prompt injection string), the gateway intercepts the body, sanitizes the matching fields, increments the redaction counter, logs a structured warning (`slog`), and forwards the cleaned request downstream seamlessly.