go-rest-client
- 解析 .http 檔案
- 使用 tview 來切割左右區塊
- 左邊區塊為全部解析完的 API
- 右邊區塊為顯示 API 的執行結果
- 啟動時監聽 .http 檔案並時時更新左邊區塊的內容
- 編譯完的檔案直接放到 /usr/sbin => gorc (go-rest-client)
- gorc {file_path}
