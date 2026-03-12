package container

import "context"

// Repository 定義容器資料的持久化契約。
//
// 實作方應遵守以下原則：
// 1. 聚焦容器 metadata 的儲存與查詢，不應混入 runtime 操作（Start/Stop/Delete container runtime）邏輯。
// 2. 尊重 ctx 生命週期：當 ctx 取消或逾時，應盡快中止操作並回傳可判別的 context 錯誤。
// 3. 錯誤語義需可判別：至少能區分「參數不合法」、「資料不存在」、「儲存層故障」。
// 4. 所有查詢與修改都應維持「使用者隔離」語義：以 user 維度查詢時不得跨使用者回傳或修改資料。
// 5. 欄位語義需穩定：Command/Env 等結構化資料不得因實作差異產生不一致格式（例如 nil 與空集合語義漂移）。
type Repository interface {
	// Create 建立一筆容器資料。
	//
	// 邊界與約定：
	// 1. c 不可為 nil，必要欄位（UserID、Image、RuntimeID、Status）不可為零值/空字串。
	// 2. 實作可對可選欄位做一致化（例如 Command nil -> []、Env nil -> {}），但行為需固定且可預期。
	// 3. 成功時可回填儲存層生成欄位（例如 ID、CreatedAt、UpdatedAt）。
	// 4. 失敗時不得回傳部分成功狀態；呼叫端應能明確判斷本次建立未完成。
	Create(ctx context.Context, c *Container) error

	// GetByIDAndUser 依容器 ID 與使用者 ID 精確查詢單筆容器。
	//
	// 邊界與約定：
	// 1. id 與 userID 必須為有效主鍵（通常 > 0）；無效值應回傳參數錯誤。
	// 2. 查詢條件需同時包含 id 與 userID，不可只用 id 查詢避免越權存取。
	// 3. 查無資料時應回傳零值 Container 與可判別的 not found 錯誤。
	// 4. 查詢成功時，回傳資料應包含完整欄位，且 error 必須為 nil。
	GetByIDAndUser(ctx context.Context, id, userID uint) (Container, error)

	// Update 更新既有容器資料。
	//
	// 邊界與約定：
	// 1. c 不可為 nil，且需包含有效的 ID 與 UserID 以定位唯一資料。
	// 2. 更新條件必須同時包含 ID 與 UserID，避免跨使用者覆蓋資料。
	// 3. 若目標不存在，應回傳可判別的 not found 錯誤，而非靜默成功。
	// 4. 更新後的 Command/Env 等結構化欄位需維持與 Create 相同的一致化規則。
	Update(ctx context.Context, c *Container) error

	// Delete 刪除指定使用者名下的容器資料。
	//
	// 邊界與約定：
	// 1. id 與 userID 必須為有效值；刪除條件需同時包含兩者。
	// 2. 若目標不存在，應回傳可判別的 not found 錯誤。
	// 3. 刪除失敗時不得回傳成功，需保證回傳語義與實際狀態一致。
	Delete(ctx context.Context, id, userID uint) error

	// ListByUser 列出指定使用者的全部容器資料。
	//
	// 邊界與約定：
	// 1. userID 必須為有效值（通常 > 0）；無效值應回傳參數錯誤。
	// 2. 回傳結果需只包含該 userID 的資料，不可混入其他使用者記錄。
	// 3. 當無資料時應回傳空陣列（非 nil）或由實作明確定義的一致行為。
	// 4. 若有排序約定（例如建立時間或 ID 反序），實作需保持穩定且可預期。
	ListByUser(ctx context.Context, userID uint) ([]Container, error)
}
