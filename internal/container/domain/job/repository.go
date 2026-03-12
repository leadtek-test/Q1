package job

import "context"

// Repository 定義建立容器非同步任務狀態的持久化契約。
type Repository interface {
	// Create 新增一筆容器建立任務。
	// 成功時可回填儲存層生成欄位（例如 CreatedAt、UpdatedAt）。
	Create(ctx context.Context, job *CreateContainerJob) error

	// GetByJobIDAndUser 依 jobID 與 userID 查詢任務。
	// 查無資料時需回傳可判別的 not found 錯誤。
	GetByJobIDAndUser(ctx context.Context, jobID string, userID uint) (CreateContainerJob, error)

	// Update 更新既有任務狀態。
	// 需使用 jobID 與 userID 精準定位目標，避免跨使用者更新。
	Update(ctx context.Context, job *CreateContainerJob) error
}
