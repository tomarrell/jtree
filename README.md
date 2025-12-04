# 🌲 jtree

Display Jaeger traces as a hierarchical tree in your terminal. The compact output is ideal for piping to LLM CLI agents, enabling AI-assisted trace analysis without overwhelming context.

## 🚀 Usage

```bash
# Basic usage
jtree <trace-id>

# Pass a full Jaeger URL directly
jtree http://localhost:16686/trace/abc123def456

# Filter to slow error spans
jtree -min-duration 100ms -error <trace-id>

# Verbose JSON output with all tags
jtree -json <trace-id>
```

## 🤖 LLM Integration

Pipe trace data directly to LLM CLI agents for AI-assisted debugging:

```bash
jtree abc123 | claude "What's causing the latency in this trace?"
jtree -error abc123 | claude "Explain these errors and suggest fixes"
```

## 📋 Output

Default human-readable format:
```
call-abc123 [orchestrator] 16:43:33.529 55.47s
  stt.websocket.connect [orchestrator] 16:43:33.539 810.76ms
  conversation.turn.bot [orchestrator] 16:43:41.178 3.11s
    tts.turn [orchestrator] 16:43:41.179 3.11s
      tts.reader [orchestrator] 16:43:41.179 3.11s
```

With `-relative`:
```
call-abc123 [orchestrator] +0us 55.47s
  stt.websocket.connect [orchestrator] +9.94ms 810.76ms
  conversation.turn.bot [orchestrator] +7.65s 3.11s
```

JSON format (`-json`):
```
call-abc123 {"duration":"55.47s","service":"orchestrator","span_id":"f1f173a9f8639951","tags":{...}}
```

## 📦 Installation

### Homebrew (macOS/Linux)

```bash
brew install tomarrell/tap/jtree
```

### Go

```bash
go install github.com/tomarrell/jtree@latest
```

### Binary

Download the latest binary from the [releases page](https://github.com/tomarrell/jtree/releases).

## 🎛️ Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-url` | `http://localhost:16686` | Jaeger URL |
| `-json` | `false` | Output verbose JSON with all tags |
| `-min-duration` | `0` | Only show spans with duration >= value (e.g. 100ms, 1s) |
| `-error` | `false` | Only show error spans and their ancestors |
| `-service` | | Only show spans from this service |
| `-depth` | `0` | Limit tree depth (0 = unlimited) |
| `-relative` | `false` | Show timestamps relative to trace start |
| `-version` | `false` | Print version and exit |

## 📄 License

MIT
