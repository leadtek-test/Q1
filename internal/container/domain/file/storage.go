package file

import (
	"context"
	"io"
)

// ObjectStorage 定義物件儲存（S3/MinIO/本地模擬）上傳契約。
//
// 實作方應遵守以下原則：
// 1. 僅負責將內容流寫入指定 key，不應在此層直接處理業務授權或資料庫交易。
// 2. 必須依傳入的 size 寫入內容，避免多寫/少寫造成 metadata 與實體不一致。
// 3. 需處理不可信 key：避免路徑穿越、非法字元或覆蓋風險（依實作策略處理）。
// 4. 當 ctx 取消時，應盡快停止 I/O 並回傳對應錯誤。
type ObjectStorage interface {
	// Upload 上傳檔案內容到 object storage。
	//
	// 邊界與約定：
	// 1. key 應視為物件識別路徑，需由呼叫端保證可唯一定位；實作不可任意改寫語義。
	// 2. body 為一次性讀取串流，實作不可假設其可重複讀取。
	// 3. size 必須為非負值且與實際寫入長度一致；不一致時應回傳錯誤。
	// 4. contentType 可作為儲存 metadata 或 header，若無法使用也不得影響資料完整寫入。
	Upload(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
}
