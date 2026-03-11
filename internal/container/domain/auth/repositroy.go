package auth

import "time"

// PasswordHasher 定義密碼處理元件與 auth 流程的交互方式。
// 實作時應確保此元件只負責密碼雜湊與比對，不承擔使用者查詢、登入狀態管理等業務邏輯。
// 另外需注意：
// 1. 不可回傳可逆的加密結果，必須使用適合密碼儲存的單向雜湊策略。
// 2. 相同明文密碼不應依賴固定輸出，應由演算法自行處理 salt 或等效機制。
// 3. Compare 的行為必須穩定，避免因格式錯誤或資料異常造成 panic。
type PasswordHasher interface {
	// Hash 將使用者輸入的原始密碼轉為可儲存的安全字串。
	// 實作時需保證回傳值可直接持久化保存，且不暴露原始密碼資訊。
	Hash(raw string) string

	// Compare 驗證原始密碼是否與已保存的雜湊值一致。
	// 實作時應將不合法的 encoded 視為比對失敗，而不是中斷流程，
	// 讓上層可以統一以登入失敗或驗證失敗處理。
	Compare(raw, encoded string) bool
}

// TokenManager 定義登入後身份憑證的建立與解析方式。
// 實作時應聚焦於 token 本身的生成、簽章驗證、有效期控制與 claims 還原，
// 不應在此介面中混入帳號密碼驗證、資料庫查詢或 HTTP 細節。
// 另外需注意：
// 1. Generate 與 Parse 必須使用同一套 claims 結構與簽章規則。
// 2. token 內容至少要能識別使用者身份，並可判斷是否已過期。
// 3. Parse 遇到格式錯誤、簽章錯誤、過期 token 時，應明確回傳錯誤供上層判斷。
type TokenManager interface {
	// Generate 根據已確認身份的使用者資訊建立 token 與過期時間。
	// 實作時需保證 userID、username 會被正確寫入 claims，
	// 並回傳實際生效的過期時間，讓上層可用於回應登入結果。
	Generate(userID uint, username string) (string, time.Time, error)

	// Parse 驗證 token 並還原 claims。
	// 實作時需同時完成 token 合法性檢查、過期判斷與 claims 解析，
	// 僅在 token 可被信任時回傳可供授權使用的 Claims。
	Parse(token string) (Claims, error)
}
