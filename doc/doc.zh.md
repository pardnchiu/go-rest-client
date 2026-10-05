# go-rest-client - 技術文件

> 返回 [README](./README.zh.md)

## 前置需求

- Go 1.25.1 或更高版本（僅原始碼建置／`go install` 需要）
- 支援 256 色的終端機
- 可連線至目標 API 的網路環境

## 安裝

### 從原始碼建置

```bash
git clone https://github.com/pardnchiu/go-rest-client.git
cd go-rest-client
go build -o gorc ./cmd/tui
sudo cp gorc /usr/local/bin/gorc
```

### 使用 go install

```bash
go install github.com/pardnchiu/go-rest-client/cmd/tui@latest
sudo cp "$(go env GOPATH)/bin/tui" /usr/local/bin/gorc
```

`go install` 產出的執行檔名稱為 `tui`，上述指令將其複製為 `gorc` 以便使用。

## 使用方式

### 基礎

建立 `api.http`：

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

啟動：

```bash
gorc api.http
```

左欄列出請求名稱，移動游標時右欄即顯示請求內容：

```
┌─ API ─────────────────────┐┌─ Info ─────────────────────────────────────────┐
│ Get User Info             ││ GET https://api.github.com/users/pardnchiu     │
│ Send POST Request         ││ Accept: application/json                       │
│                           ││                                                │
└───────────────────────────┘└────────────────────────────────────────────────┘
 GET https://api.github.com/users/pardnchiu
```

按 `Enter` 送出後，右欄顯示狀態、Header、耗時與格式化後的 Body：

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

### 進階：編輯即重載

保持 `gorc` 執行，於任一編輯器修改並儲存 `api.http`。檔案寫入後約 200ms 內清單自動重新解析，底部提示列依序顯示 `Edited at HH:MM:SS` → `Reloaded at HH:MM:SS`；解析失敗時改顯示紅色錯誤訊息，原清單維持不變。

### 進階：SSE 串流

```http
### SSE Stream
GET https://sse.dev/test
Accept: text/event-stream

###
```

回應 `Content-Type` 含 `text/event-stream` 時切換為串流模式：右欄逐行附加資料並自動捲到底，僅保留最後 100 行，標頭列即時更新耗時與事件數；串流結束後提示列顯示最終狀態碼與總耗時。

## 命令列參考

### 指令

| 指令 | 語法 | 說明 |
|------|------|------|
| `gorc` | `gorc <file.http>` | 解析並監控指定 `.http` 檔，啟動 TUI |

| 結束碼 | 情境 |
|--------|------|
| `1` | 未提供檔案參數、檔案不存在、無法監控檔案、解析失敗或 TUI 啟動失敗 |

### 鍵盤操作

| 按鍵 | 功能 |
|------|------|
| `↑` `↓` | 於請求清單移動並預覽 |
| `Enter` | 送出選取的請求 |
| `Tab` | 在清單與資訊面板間切換焦點 |
| `←` | 焦點移至請求清單 |
| `→` | 焦點移至資訊面板 |
| `Ctrl+C` `Esc` | 離開 |

### 支援的 HTTP 方法

`GET`、`POST`、`PUT`、`PATCH`、`DELETE`、`HEAD`、`OPTIONS`

### .http 語法

| 語法 | 說明 |
|------|------|
| `###` | 請求分隔線 |
| `### 名稱` | 定義下一個請求的顯示名稱 |
| `METHOD URL` | 請求行，方法須為大寫 |
| `Key: Value` | Header；Key 須以英文字母開頭，僅含字母、數字、`-`、`_` |
| 空行、`{`、`[` | 之後的內容視為 Body |
| `#`、`//`、`--` | 註解（Body 外） |

### 請求行為

| 項目 | 值 |
|------|----|
| 逾時 | 120 秒 |
| 狀態碼顏色 | 2xx 綠、3xx 黃、≥400 紅 |
| JSON Body | 可解析時自動縮排 |
| SSE 顯示上限 | 最後 100 行 |
| 檔案重載防抖 | 200ms |

***

©️ 2026 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
