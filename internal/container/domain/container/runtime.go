package container

import "context"

// Runtime 定義容器執行時（Docker 或其他容器引擎）的抽象能力。
//
// 實作方應遵守以下原則：
// 1. 專注在容器生命週期操作（建立、啟動、停止、刪除），不混入 HTTP/DB 邏輯。
// 2. 尊重 ctx：若上層取消或逾時，應盡快停止遠端呼叫並回傳可識別的 context 錯誤。
// 3. 錯誤應可判斷：至少能區分「輸入不合法」、「runtime 不可用」、「目標不存在/狀態不合法」。
// 4. runtimeID 必須穩定可追蹤，且足以讓後續 Start/Stop/Delete 精準定位同一資源。
// 5. 不可 silently ignore 失敗：外部系統操作未成功時，必須回傳錯誤讓上層補償。
type Runtime interface {
	// Create 根據 user 與規格建立容器，並回傳 runtime 端的容器識別碼。
	//
	// 邊界與約定：
	// 1. userID 必須有效（通常 > 0）；無效值應立即回傳參數錯誤。
	// 2. spec.Image 必須有效；若缺失或格式不合法應回傳輸入錯誤。
	// 3. workspacePath 應為可掛載且可存取路徑；若不可用應回傳錯誤而非降級忽略。
	// 4. 成功時回傳的 runtimeID 不可為空字串，且需可被後續操作使用。
	Create(ctx context.Context, userID uint, spec CreateSpec, workspacePath string) (string, error)

	// Start 啟動指定 runtimeID 的容器。
	//
	// 邊界與約定：
	// 1. runtimeID 不可為空；空值應回傳輸入錯誤。
	// 2. 若目標容器不存在或無法啟動，需回傳可追蹤錯誤。
	// 3. 若容器已在執行中，實作可選擇視為成功或特定錯誤，但行為需一致。
	Start(ctx context.Context, runtimeID string) error

	// Stop 停止指定 runtimeID 的容器。
	//
	// 邊界與約定：
	// 1. runtimeID 不可為空；空值應回傳輸入錯誤。
	// 2. 若目標不存在或無法停止，需回傳錯誤讓上層決定重試或補償。
	// 3. 若容器已停止，實作可選擇冪等成功或特定錯誤，但需保持一致語義。
	Stop(ctx context.Context, runtimeID string) error

	// Delete 刪除指定 runtimeID 的容器（必要時可強制刪除）。
	//
	// 邊界與約定：
	// 1. runtimeID 不可為空；空值應回傳輸入錯誤。
	// 2. 刪除失敗（例如被占用、權限不足）必須回傳錯誤，不得視為成功。
	// 3. 若採冪等策略處理「目標不存在」，需在實作文件中明確說明。
	Delete(ctx context.Context, runtimeID string) error
}
