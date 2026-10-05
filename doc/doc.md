# go-rest-client - Documentation

> Back to [README](../README.md)

## Prerequisites

- Go 1.25.1 or higher (only for building from source or `go install`)
- A terminal with 256-color support
- Network access to the target APIs

## Installation

### From Source

```bash
git clone https://github.com/pardnchiu/go-rest-client.git
cd go-rest-client
go build -o gorc ./cmd/tui
sudo cp gorc /usr/local/bin/gorc
```

### Using go install

```bash
go install github.com/pardnchiu/go-rest-client/cmd/tui@latest
sudo cp "$(go env GOPATH)/bin/tui" /usr/local/bin/gorc
```

`go install` produces a binary named `tui`; the command above copies it as `gorc` for convenience.

## Usage

### Basic

Create `api.http`:

```http
### Get User Info
GET https://api.github.com/users/pardnchiu
Accept: application/json

###

### Send POST Request
POST https://httpbin.org/post
Content-Type: application/json

{
  "name": "test",
  "value": 123
}

###
```

Launch:

```bash
gorc api.http
```

The left pane lists request names; moving the cursor previews the request in the right pane:

```
┌─ API ─────────────────────┐┌─ Info ─────────────────────────────────────────┐
│ Get User Info             ││ GET https://api.github.com/users/pardnchiu     │
│ Send POST Request         ││ Accept: application/json                       │
│                           ││                                                │
└───────────────────────────┘└────────────────────────────────────────────────┘
 GET https://api.github.com/users/pardnchiu
```

Press `Enter` to send; the right pane shows status, headers, duration, and the formatted body:

```
┌─ API ─────────────────────┐┌─ Info ─────────────────────────────────────────┐
│ Get User Info             ││ 200 OK                                         │
│ Send POST Request         ││ Headers:                                       │
│                           ││   Content-Type: application/json; charset=utf-8│
│                           ││   Duration: 245.123ms                          │
│                           ││                                                │
│                           ││ Body:                                          │
│                           ││   {                                            │
│                           ││     "login": "pardnchiu",                      │
│                           ││     "type": "User"                             │
│                           ││   }                                            │
└───────────────────────────┘└────────────────────────────────────────────────┘
 (200) GET https://api.github.com/users/pardnchiu | 245.123ms
```

### Advanced: Edit and Reload

Keep `gorc` running, then edit and save `api.http` in any editor. About 200ms after the write, the list is re-parsed and the hint bar shows `Edited at HH:MM:SS` → `Reloaded at HH:MM:SS`; on a parse failure it shows a red error and keeps the previous list.

### Advanced: SSE Streaming

```http
### SSE Stream
GET https://sse.dev/test
Accept: text/event-stream

###
```

When the response `Content-Type` contains `text/event-stream`, the client switches to streaming mode: the right pane appends lines and auto-scrolls, keeps only the last 100 lines, and updates duration and event count live; when the stream ends the hint bar shows the final status code and total duration.

## CLI Reference

### Commands

| Command | Syntax | Description |
|---------|--------|-------------|
| `gorc` | `gorc <file.http>` | Parse and watch the given `.http` file, then launch the TUI |

| Exit Code | Condition |
|-----------|-----------|
| `1` | Missing file argument, file not found, watch failure, parse failure, or TUI startup failure |

### Keyboard Controls

| Key | Function |
|-----|----------|
| `↑` `↓` | Move through the request list with preview |
| `Enter` | Send the selected request |
| `Tab` | Toggle focus between list and info pane |
| `←` | Focus the request list |
| `→` | Focus the info pane |
| `Ctrl+C` `Esc` | Exit |

### Supported HTTP Methods

`GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD`, `OPTIONS`

### .http Syntax

| Syntax | Description |
|--------|-------------|
| `###` | Request separator |
| `### Name` | Display name for the next request |
| `METHOD URL` | Request line; method must be uppercase |
| `Key: Value` | Header; key must start with a letter and contain only letters, digits, `-`, `_` |
| Blank line, `{`, `[` | Everything after is treated as body |
| `#` (not `###`) | Comment; also skipped inside body |
| `//`, `--` | Comments (outside body only) |

### Request Behavior

| Item | Value |
|------|-------|
| Timeout | 120 seconds |
| Status color | 2xx green, 3xx yellow, ≥400 red |
| JSON body | Pretty-printed when parsable |
| SSE display limit | Last 100 lines |
| Reload debounce | 200ms |

***

©️ 2026 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
