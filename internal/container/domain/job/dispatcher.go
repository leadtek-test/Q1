package job

import "context"

// CreateContainerDispatcher 負責建立容器任務的派送與背景監聽。
// 同一個實作必須提供完整流程，避免派送與執行使用不同底層技術。
type CreateContainerDispatcher interface {
	// DispatchCreateContainer 應儲存任務初始狀態並排入待處理佇列，成功後回傳 jobID。
	DispatchCreateContainer(ctx context.Context, task CreateContainerTask) (string, error)
	// Listen 應持續監聽任務來源並觸發處理流程，直到 context 結束。
	Listen(ctx context.Context)
}
