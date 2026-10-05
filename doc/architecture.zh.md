# go-rest-client - 架構

> 返回 [README](./README.zh.md)

## 概覽

```mermaid
graph TB
    CLI[cmd/tui 進入點] -->|檔案路徑| P[internal/parser]
    CLI -->|建立| T[internal/ui TUI]
    CLI -->|建立| W[fsnotify Watcher]
    F[.http 檔案] --> P
    W -->|Write / Create 事件| P
    P -->|Requests| T
    T -->|Enter| H[net/http Client]
    H -->|一般回應| R[handleResponse]
    H -->|text/event-stream| S[handleSSEResponse]
    R --> V[RightView 資訊面板]
    S --> V
```

## Module: cmd/tui

驗證命令列參數、組裝 Watcher 與 TUI，完成首次解析後啟動事件迴圈。

```mermaid
graph TB
    subgraph cmd/tui
        A[檢查 os.Args] --> B[os.Stat 檔案存在]
        B --> C[fsnotify.NewWatcher]
        C --> D[ui.NewTUI]
        D --> E[watcher.Add 檔案]
        E --> G[parser.ReadFile 首次解析]
        G --> I[go parser.WatchFile]
        I --> J[tui.App.Run]
    end
    A -->|參數不足| X[exit 1]
    B -->|不存在| X
    E -->|失敗| X
    G -->|失敗| X
```

## Module: internal/parser

將 `.http` 檔逐行解析為 `[]*ui.Request`，並監控檔案變動觸發重載。

```mermaid
graph TB
    subgraph parser
        RF[ReadFile] --> SC[bufio.Scanner 逐行]
        SC --> RX[正規式比對請求行]
        SC --> CH[checkHeader 驗證 Header 名稱]
        SC --> BD[Body 累積]
        RF -->|Mu.Lock| ST[寫入 TUI.Requests]
        RF --> UL[TUI.UpdateLeftView]
        WF[WatchFile] -->|200ms 防抖| RL[ReloadFile]
        RL --> RF
        RL -->|保留游標| SI[LeftView.SetCurrentItem]
    end
    FS[fsnotify Events / Errors] --> WF
    RL -->|QueueUpdateDraw| HV[HintView 提示列]
```

## Module: internal/ui

以 tview 建構雙欄介面，負責按鍵綁定、請求送出與回應渲染。

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
    subgraph 版面
        L[LeftView 請求清單 1/3] --- R[RightView 資訊面板 2/3]
        H[HintView 提示列 1 行]
    end
    K[InputCapture] -->|← / Tab| L
    K -->|→ / Tab| R
    K -->|Ctrl+C / Esc| Q[App.Stop]
```

## 資料流

```mermaid
sequenceDiagram
    participant U as 使用者
    participant T as TUI
    participant G as goroutine
    participant API as 目標 API
    U->>T: Enter
    T->>T: Mu.RLock 取 Request，焦點移至 RightView
    T->>G: 背景送出（120s 逾時）
    G->>API: http.Client.Do
    API-->>G: Response
    alt Content-Type 含 text/event-stream
        loop 每一行
            G->>T: QueueUpdateDraw 附加資料（保留最後 100 行）
        end
        G->>T: 提示列顯示狀態碼與總耗時
    else 一般回應
        G->>G: io.ReadAll + JSON 縮排
        G->>T: QueueUpdateDraw 狀態、Header、耗時、Body
    end
```

```mermaid
sequenceDiagram
    participant E as 編輯器
    participant W as fsnotify
    participant P as WatchFile
    participant R as ReloadFile
    participant T as TUI
    E->>W: 存檔
    W->>P: Write / Create 事件
    P->>P: 重設 200ms timer
    P->>R: timer 觸發
    R->>T: 提示 Edited at
    R->>R: ReadFile 重新解析
    alt 解析成功
        R->>T: 更新清單、還原游標、提示 Reloaded at
    else 解析失敗
        R->>T: 紅色錯誤提示，保留原清單
    end
```

## 狀態機

`ReadFile` 逐行解析時的狀態轉換：

```mermaid
stateDiagram-v2
    [*] --> 無請求
    無請求 --> 無請求: ### 名稱 / 註解 / 空行
    無請求 --> Header區: METHOD URL
    Header區 --> Header區: 合法 Key: Value
    Header區 --> Body區: 空行 / { / [ / 其他文字
    Body區 --> Header區: METHOD URL（收錄請求與 Body）
    Body區 --> Body區: 任意行
    Header區 --> 無請求: ###（收錄請求）
    Body區 --> 無請求: ###（收錄請求與 Body）
    無請求 --> [*]: EOF
    Header區 --> [*]: EOF（收錄請求）
    Body區 --> [*]: EOF（收錄請求與 Body）
```

***

©️ 2026 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
