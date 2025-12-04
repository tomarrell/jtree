# jtree

Display Jaeger traces as a hierarchical tree in your terminal.

## Installation

```bash
go install github.com/tomarrell/jtree@latest
```

## Usage

```bash
# Basic usage - human readable output
jtree <trace-id>

# Pass a full Jaeger URL directly
jtree http://localhost:16686/trace/abc123def456

# Verbose JSON output with all tags
jtree -json <trace-id>

# Filter to slow spans (>= 100ms)
jtree -min-duration 100ms <trace-id>

# Show only error spans
jtree -error <trace-id>

# Filter by service
jtree -service orchestrator <trace-id>

# Limit tree depth
jtree -depth 3 <trace-id>

# Show relative timestamps
jtree -relative <trace-id>

# Combine filters
jtree -min-duration 1s -depth 2 -relative <trace-id>
```

## Output

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

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-url` | `http://localhost:16686` | Jaeger URL |
| `-json` | `false` | Output verbose JSON with all tags |
| `-min-duration` | `0` | Only show spans with duration >= value (e.g. 100ms, 1s) |
| `-error` | `false` | Only show error spans and their ancestors |
| `-service` | | Only show spans from this service |
| `-depth` | `0` | Limit tree depth (0 = unlimited) |
| `-relative` | `false` | Show timestamps relative to trace start |

## License

MIT
