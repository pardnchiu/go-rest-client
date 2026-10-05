# go-rest-client - Architecture

> Back to [README](../README.md)

## Overview

```mermaid
graph TB
    CLI[cmd/tui entry] -->|file path| P[internal/parser]
    CLI -->|create| T[internal/ui TUI]
    CLI -->|create| W[fsnotify Watcher]
    F[.http File] --> P
    W -->|Write / Create events| P
    P -->|Requests| T
    T -->|Enter| H[net/http Client]
    H -->|regular response| R[handleResponse]
    H -->|text/event-stream| S[handleSSEResponse]
    R --> V[RightView info pane]
    S --> V
```

## Module: cmd/tui

Validates CLI arguments, wires the watcher and TUI, performs the initial parse, and starts the event loop.

```mermaid
graph TB
    subgraph cmd/tui
        A[Check os.Args] --> B[os.Stat file exists]
        B --> C[fsnotify.NewWatcher]
        C --> D[ui.NewTUI]
        D --> E[watcher.Add file]
        E --> G[parser.ReadFile initial parse]
        G --> I[go parser.WatchFile]
        I --> J[tui.App.Run]
    end
    A -->|missing arg| X[exit 1]
    B -->|not found| X
    E -->|failed| X
    G -->|failed| X
```

## Module: internal/parser

Parses the `.http` file line by line into `[]*ui.Request` and watches the file to trigger reloads.

```mermaid
graph TB
    subgraph parser
        RF[ReadFile] --> SC[bufio.Scanner per line]
        SC --> RX[Regex match request line]
        SC --> CH[checkHeader validates header name]
        SC --> BD[Body accumulation]
        RF -->|Mu.Lock| ST[Write TUI.Requests]
        RF --> UL[TUI.UpdateLeftView]
        WF[WatchFile] -->|200ms debounce| RL[ReloadFile]
        RL --> RF
        RL -->|restore cursor| SI[LeftView.SetCurrentItem]
    end
    FS[fsnotify Events / Errors] --> WF
    RL -->|QueueUpdateDraw| HV[HintView hint bar]
```

## Module: internal/ui

Builds the split-pane interface with tview and handles key bindings, request dispatch, and response rendering.

```mermaid
classDiagram
    class TUI {
        +App *tview.Application
        +Pages *tview.Pages
        +LeftView *tview.List
        +RightView *tview.TextView
        +HintView *tview.TextView
        +Watcher *fsnotify.Watcher
        +Filepath string
        +Requests []*Request
        +Mu sync.RWMutex
        +UpdateLeftView()
        -showRequestDetail(req)
        -sendRequest(index)
        -handleResponse(resp, start, method, url)
        -handleSSEResponse(resp, start, method, url)
    }
    class Request {
        +Name string
        +Method string
        +URL string
        +Headers map[string]string
        +Body string
    }
    TUI "1" --> "*" Request
```

```mermaid
graph LR
    subgraph Layout
        L[LeftView request list 1/3] --- R[RightView info pane 2/3]
        H[HintView hint bar 1 row]
    end
    K[InputCapture] -->|← / Tab| L
    K -->|→ / Tab| R
    K -->|Ctrl+C / Esc| Q[App.Stop]
```

## Data Flow

```mermaid
sequenceDiagram
    participant U as User
    participant T as TUI
    participant G as goroutine
    participant API as Target API
    U->>T: Enter
    T->>T: Mu.RLock reads Request, focus RightView
    T->>G: Send in background (120s timeout)
    G->>API: http.Client.Do
    API-->>G: Response
    alt Content-Type contains text/event-stream
        loop each line
            G->>T: QueueUpdateDraw append data (keep last 100 lines)
        end
        G->>T: Hint bar shows status code and total duration
    else regular response
        G->>G: io.ReadAll + JSON indent
        G->>T: QueueUpdateDraw status, headers, duration, body
    end
```

```mermaid
sequenceDiagram
    participant E as Editor
    participant W as fsnotify
    participant P as WatchFile
    participant R as ReloadFile
    participant T as TUI
    E->>W: Save
    W->>P: Write / Create event
    P->>P: Reset 200ms timer
    P->>R: Timer fires
    R->>T: Hint Edited at
    R->>R: ReadFile re-parse
    alt parse succeeded
        R->>T: Refresh list, restore cursor, hint Reloaded at
    else parse failed
        R->>T: Red error hint, keep previous list
    end
```

## State Machine

State transitions while `ReadFile` parses each line:

```mermaid
stateDiagram-v2
    [*] --> NoRequest
    NoRequest --> NoRequest: ### Name / comment / blank
    NoRequest --> Headers: METHOD URL
    Headers --> Headers: valid Key: Value
    Headers --> Body: blank / { / [ / other text
    Body --> Headers: METHOD URL (commit request and body)
    Body --> Body: any line
    Headers --> NoRequest: ### (commit request)
    Body --> NoRequest: ### (commit request and body)
    NoRequest --> [*]: EOF
    Headers --> [*]: EOF (commit request)
    Body --> [*]: EOF (commit request and body)
```

***

©️ 2026 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
