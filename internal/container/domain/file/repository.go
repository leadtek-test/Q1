package file

import "context"

// Repository 定義檔案中繼資料（metadata）的持久化契約。
//
// 實作方應遵守以下原則：
// 1. 只處理 metadata 的儲存與查詢，不應在此介面內混入實體檔案上傳、檔案內容轉換等行為。
// 2. 尊重 ctx 生命週期：當 ctx 取消或逾時，應盡快中止操作並回傳可辨識的 context 錯誤。
// 3. 錯誤語義可判別：至少要讓上層可區分「參數不合法」與「資料庫/儲存層故障」。
// 4. 不做隱式修正：不得默默修改呼叫端提供的關鍵欄位（如 UserID、ObjectKey、Size），除非契約已明確說明。
type Repository interface {
	// Create 建立一筆檔案 metadata。
	//
	// 邊界與約定：
	// 1. f 不可為 nil，且必要欄位（UserID、FileName、ObjectKey、ContentType、Size、WorkspacePath）不可為零值。
	// 2. 成功時可回填資料庫生成欄位（例如 ID、CreatedAt）。
	// 3. 失敗時不得回傳部分成功狀態；呼叫端應能以 error 判斷本次建立未完成。
	// 4. 若發生唯一鍵或約束衝突，應回傳可識別錯誤，避免上層只能依字串比對錯誤訊息。
	Create(ctx context.Context, f *File) error
}
