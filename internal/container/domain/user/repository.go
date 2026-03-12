package user

import "context"

// Repository 定義 User 的持久化契約。
//
// 實作方應遵守以下原則：
// 1. 尊重 ctx 的生命週期：當 ctx 已取消或逾時，應盡快停止操作並回傳可被識別的 context 錯誤。
// 2. 回傳可判別的業務錯誤：至少需要讓上層可區分「資料不存在」與「唯一鍵衝突」兩類情境。
// 3. 查詢語義保持穩定：回傳的 User 應是完整且一致的資料快照（含 ID、Username、PasswordMD5、CreatedAt、UpdatedAt）。
// 4. 不做隱式模糊行為：不應對輸入做猜測式修正（例如模糊查詢、忽略大小寫），除非實作已明確約定並文件化。
type Repository interface {
	// Create 新增一筆使用者資料。
	//
	// 邊界與約定：
	// 1. u 不可為 nil，Username 與 PasswordMD5 不可為空字串；否則應回傳參數錯誤。
	// 2. 若 Username 已存在，應回傳可判別的衝突錯誤，避免上層只能依字串比對錯誤訊息。
	// 3. 成功後可回填資料庫產生欄位到 u（例如 ID、CreatedAt、UpdatedAt）。
	// 4. 失敗時不得產生「半成功」狀態：要嘛完整建立，要嘛不建立。
	Create(ctx context.Context, u *User) error

	// GetByUsername 依 Username 精確查詢單一使用者。
	//
	// 邊界與約定：
	// 1. username 不可為空字串；空值應回傳參數錯誤。
	// 2. 查詢應為精確匹配，不應進行模糊比對或自動正規化。
	// 3. 查無資料時，應回傳零值 User 與 consts.ErrnoUserNotFound 錯誤
	// 4. 找到資料時，回傳值應包含完整欄位，且 error 必須為 nil。
	GetByUsername(ctx context.Context, username string) (User, error)

	// GetByID 依主鍵 ID 精確查詢單一使用者。
	//
	// 邊界與約定：
	// 1. id 必須為有效主鍵（通常 > 0）；無效 id 應回傳參數錯誤。
	// 2. 查無資料時，應回傳零值 User 與 consts.ErrnoUserNotFound 錯誤。
	// 3. 查詢結果應是與儲存層一致的最新已提交資料，不應回傳部分欄位。
	// 4. 若 ctx 取消或逾時，應優先回傳對應的 context 錯誤。
	GetByID(ctx context.Context, id uint) (User, error)
}
