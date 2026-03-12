package ports

import "github.com/gin-gonic/gin"

// ServerInterface 定義 container 模組的 HTTP 入口行為。
// 實作者需要在各方法中完成請求解析、必要驗證、調用對應業務邏輯，
// 並依處理結果回傳一致的 HTTP 響應。
type ServerInterface interface {
	// Register 建立新用戶帳號。
	// 需負責驗證註冊資料、檢查帳號是否可建立，並在成功後回傳註冊結果。
	Register(c *gin.Context)

	// Login 驗證用戶身份並建立登入狀態。
	// 需負責檢查登入憑證是否合法，並在成功後回傳可供後續受保護操作使用的身份資訊（token）。
	Login(c *gin.Context)

	// Upload 接收檔案並綁定到指定用戶。
	// 需負責確認操作者身份、驗證上傳內容與目標用戶是否合法，
	// 並在成功後回傳檔案已被保存或登記的結果。
	Upload(c *gin.Context)

	// CreateContainer 為當前用戶提交建立容器任務。
	// 需負責驗證建立參數、確認資源歸屬與建立條件，並回傳可追蹤任務狀態的 job id。
	CreateContainer(c *gin.Context)

	// GetCreateContainerJob 查詢建立容器任務狀態。
	// 需負責確認任務屬於當前用戶，並回傳任務目前狀態資訊。
	GetCreateContainerJob(c *gin.Context)

	// ListContainers 查詢當前用戶所擁有的所有容器。
	// 需負責識別用戶身份，並回傳該用戶可查看的容器清單與必要狀態資訊。
	ListContainers(c *gin.Context)

	// UpdateContainerStatus 更新指定容器的運行狀態。
	// 需負責確認容器屬於當前用戶、目標狀態是否合法，
	// 並執行開啟或關閉等狀態切換流程後回傳最新狀態。
	UpdateContainerStatus(c *gin.Context)

	// DeleteContainer 刪除當前用戶所擁有的指定容器。
	// 需負責確認容器存在且可由該用戶刪除，完成刪除流程後回傳刪除結果。
	DeleteContainer(c *gin.Context)
}
