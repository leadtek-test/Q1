package file

// Workspace 定義本地工作區（workspace）檔案保存契約。
//
// 實作方應遵守以下原則：
// 1. 只負責將位元組資料保存到工作區，路徑規則需穩定且可追蹤。
// 2. 需確保 userID 與 fileName 產生的最終路徑安全，避免目錄穿越或覆寫非預期檔案。
// 3. 檔案寫入失敗時不可回傳成功路徑，避免上層誤以為本地副本存在。
// 4. 回傳路徑需可被後續流程使用（例如容器掛載、追蹤、清理）。
type Workspace interface {
	// EnsureUserDir 確保使用者 workspace 目錄存在，並回傳可用路徑。
	//
	// 邊界與約定：
	// 1. userID 必須是有效識別值（通常 > 0），無效值應回傳參數錯誤。
	// 2. 成功回傳的路徑需可安全使用於後續掛載或檔案保存流程。
	EnsureUserDir(userID uint) (string, error)

	// Save 將檔案內容保存到指定使用者工作區，並回傳最終保存路徑。
	//
	// 邊界與約定：
	// 1. userID 必須是有效識別值（通常 > 0），無效值應回傳參數錯誤。
	// 2. fileName 不可為空，且應視為不可信輸入處理（至少需做 basename 或等效防護）。
	// 3. data 可為空內容，但實作應明確定義空檔案是否允許並保持一致行為。
	// 4. 成功時回傳可讀取的絕對或相對路徑（由實作約定），失敗時必須回傳非 nil error。
	Save(userID uint, fileName string, data []byte) (string, error)
}
