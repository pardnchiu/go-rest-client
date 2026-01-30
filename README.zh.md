> [!NOTE]
> 此 README 由 [Claude Code](https://github.com/pardnchiu/skill-readme-generate) 生成，英文版請參閱 [這裡](./README.md)。

# go-rest-client

[![pkg](https://pkg.go.dev/badge/github.com/pardnchiu/go-rest-client.svg)](https://pkg.go.dev/github.com/pardnchiu/go-rest-client)
[![license](https://img.shields.io/github/license/pardnchiu/go-rest-client)](LICENSE)

> 基於終端的 REST API 測試工具，相容 VSCode REST Client 擴充功能的 `.http` 檔案格式，透過直觀的 TUI 介面執行 HTTP 請求並即時顯示回應。

## 目錄

- [功能特點](#功能特點)
- [架構](#架構)
- [安裝](#安裝)
- [使用方法](#使用方法)
- [命令列參考](#命令列參考)
- [使用場景](#使用場景)
- [授權](#授權)
- [Author](#author)
- [Stars](#stars)

## 功能特點

- **TUI 介面**：使用 `tview` 建構的分割面板介面（API 清單 / 回應顯示）
- **VSCode REST Client 相容**：完整支援 `.http` 檔案格式
- **即時回應**：顯示狀態碼、標頭、回應主體與請求耗時
- **SSE 支援**：即時串流顯示 Server-Sent Events 資料
- **檔案監控**：自動偵測檔案變更並重新載入請求
- **JSON 格式化**：自動格式化 JSON 回應以提升可讀性
- **多重方法**：支援 GET、POST、PUT、DELETE、PATCH、HEAD、OPTIONS
- **鍵盤導航**：Tab 與方向鍵快速切換視圖

## 架構

```mermaid
graph TB
    A[main.go] --> B[ui.TUI]
    A --> C[parser]

    B --> D[TUI Components]
    D --> E[LeftView - API 清單]
    D --> F[RightView - 回應顯示]
    D --> G[HintView - 狀態提示]

    C --> H[ReadFile]
    C --> I[WatchFile]
    C --> J[ReloadFile]

    B --> K[sendRequest]
    K --> L[HTTP Client]
    L --> M{Content-Type}
    M -->|text/event-stream| N[handleSSEResponse]
    M -->|其他| O[handleResponse]

    I -->|檔案變更| J
```

## 安裝

### 從原始碼編譯

```bash
git clone https://github.com/pardnchiu/go-rest-client.git
cd go-rest-client
go build -o gorc ./cmd/tui
```

### 安裝至系統路徑

```bash
sudo cp gorc /usr/local/bin/gorc
```

### 使用 Go 安裝

```bash
go install github.com/pardnchiu/go-rest-client/cmd/tui@latest
sudo cp $(go env GOPATH)/bin/tui /usr/local/bin/gorc
```

## 使用方法

### 1. 建立請求檔案

建立 `api.http` 檔案並定義請求：

```http
### 取得使用者資訊
GET https://api.github.com/users/pardnchiu
Accept: application/json

###

### 發送 POST 請求
POST https://httpbin.org/post
Content-Type: application/json

{
  "name": "test",
  "value": 123
}

###

### SSE 串流
GET https://httpbin.org/stream/5
Accept: text/event-stream
```

### 2. 啟動程式

```bash
gorc api.http
```

### 3. 鍵盤操作

| 按鍵 | 功能 |
|------|------|
| `Tab` | 在 API 清單與回應視圖間切換 |
| `←` / `→` | 方向鍵快速切換視圖 |
| `↑` / `↓` | 上下選擇 API |
| `Enter` | 執行選取的請求 |
| `Ctrl+C` / `Esc` | 退出程式 |

## 命令列參考

### 語法

```bash
gorc <file.http>
```

### 參數

| 參數 | 說明 |
|------|------|
| `file.http` | `.http` 請求檔案路徑（必要） |

### 支援的 HTTP 方法

| 方法 | 說明 |
|------|------|
| `GET` | 取得資源 |
| `POST` | 建立資源 |
| `PUT` | 更新資源（完整） |
| `PATCH` | 更新資源（部分） |
| `DELETE` | 刪除資源 |
| `HEAD` | 取得標頭 |
| `OPTIONS` | 取得支援的方法 |

### .http 檔案格式

```http
### 請求名稱
METHOD URL
Header-Name: Header-Value

{request body}

###
```

- `###` 作為請求分隔符號
- `### 名稱` 定義請求名稱
- 標頭緊接在 METHOD URL 之後
- 空行後為請求主體
- 支援 `//` 與 `#` 註解

## 使用場景

### API 開發測試

在開發階段快速測試 RESTful API，無需切換至瀏覽器或使用 curl。

```http
### 建立使用者
POST https://api.example.com/users
Content-Type: application/json

{
  "name": "John",
  "email": "john@example.com"
}
```

### SSE 即時監控

監控 Server-Sent Events 串流資料，適用於即時通知或事件監聽。

```http
### 監聽事件
GET https://api.example.com/events
Accept: text/event-stream
```

### 多環境切換

為不同環境建立獨立的 `.http` 檔案：

```bash
gorc dev.http      # 開發環境
gorc staging.http  # 測試環境
gorc prod.http     # 正式環境
```

## 授權

MIT License

## Author

<img src="https://avatars.githubusercontent.com/u/25631760" align="left" width="96" height="96" style="margin-right: 0.5rem;">

<h4 style="padding-top: 0">邱敬幃 Pardn Chiu</h4>

<a href="mailto:dev@pardn.io" target="_blank">
<img src="https://pardn.io/image/email.svg" width="48" height="48">
</a> <a href="https://linkedin.com/in/pardnchiu" target="_blank">
<img src="https://pardn.io/image/linkedin.svg" width="48" height="48">
</a>

## Stars

[![Star](https://api.star-history.com/svg?repos=pardnchiu/go-rest-client&type=Date)](https://www.star-history.com/#pardnchiu/go-rest-client&Date)

***

©️ 2026 [邱敬幃 Pardn Chiu](https://linkedin.com/in/pardnchiu)
