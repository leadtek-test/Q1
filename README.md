# Q1 Container Service (Interview Project)

這是一個以 Go 實作的容器管理服務，提供使用者註冊/登入、檔案上傳、非同步建立容器，以及容器列表/啟停/刪除等能力。

## 技術棧
- Language: Go `1.26.1`
- Web: `gin`
- ORM/DB: `gorm` + `PostgreSQL`
- Container Runtime: Docker Engine API (`moby/client`)
- Config: `viper` (`global.yaml` + env override)
- Auth: JWT
- Logging: `logrus`
- API Spec: OpenAPI 3 (`api/openapi/openapi.yaml`)

## 系統架構
本專案採用偏 Clean Architecture 的分層方式，核心依賴方向如下：

```text
HTTP/Router (ports)
        |
        v
Application (app: command/query)
        |
        v
Domain (entities + repository/runtime interfaces)
        ^
        |
Adapters (postgres/docker/local fs/jwt/queue)

Service Composition (service/application.go) 負責組裝依賴
```

關鍵點：
- `domain` 不依賴 gin/gorm/docker 等框架。
- `app` 透過 `domain` 介面工作，不直接綁定基礎設施實作。
- `adapters` 實作 domain 介面，將外部技術細節包起來。
- `ports` 僅做 request/response mapping + auth/context 萃取。

## Folder Tree 與資料夾說明
```text
.
├── api/
│   └── openapi/
│       └── openapi.yaml            # API 規格文件
├── internal/
│   ├── common/                     # 共用基礎元件（跨模組）
│   │   ├── client/                 # API request schema（client side DTO）
│   │   ├── config/                 # viper 載入與設定
│   │   ├── consts/                 # errno / 錯誤訊息定義
│   │   ├── decorator/              # command/query 裝飾器（logging 等）
│   │   ├── handler/                # 通用 handler（errors、redis、singleton）
│   │   ├── logging/                # logging 初始化與欄位
│   │   ├── middleware/             # HTTP request logging middleware
│   │   ├── server/                 # HTTP server 建立與啟停封裝
│   │   └── util/                   # 小型工具函式
│   └── container/                  # 主要業務模組
│       ├── main.go                 # 程式進入點（signal + graceful shutdown）
│       ├── service/                # 依賴注入/組裝 Application
│       ├── ports/                  # HTTP handlers/router/middleware
│       │   ├── middleware/         # JWT 驗證 middleware
│       │   └── contextx/           # gin context key 定義
│       ├── app/                    # Use case 層（Command/Query）
│       │   ├── command/            # 寫操作（create/start/stop/delete/upload...)
│       │   ├── query/              # 讀操作（list containers/job status）
│       │   └── dto/                # API response DTO
│       ├── domain/                 # 核心領域模型 + 介面（repository/runtime/token）
│       ├── adapters/               # domain 介面實作（postgres/docker/local/jwt/channel）
│       └── infrastructure/         # DB persistence model 與查詢 builder
├── docker-compose.yml              # 本地相依服務（Postgres/MinIO）
├── go.work                         # 多模組 workspace
└── workspace_test.go               # workspace 驗證測試
```

## 核心設計細節
### 1. Command/Query 分離
- `app/command` 專注寫入流程與副作用。
- `app/query` 專注讀取投影。
- 共通橫切邏輯由 `internal/common/decorator` 包裝，減少 handler 重複碼。

### 2. 非同步建立容器 Job
- `POST /containers` 只建立 job 並回傳 `job_id`。
- 背景 dispatcher (`adapters/create_container_dispatcher_channel.go`) 以 channel 排程 job。
- job 狀態機：`accepted -> creating -> succeeded/failed`。

### 3. 同一 Container 的互斥控制
- 以 `(userID, containerID)` 作為鎖 key。
- `UpdateContainerStatus` / `DeleteContainer` 共享同一把鎖，避免啟動/刪除競爭。
- 具備 timeout 策略：
  - 等待鎖超過 3 秒：回 `ErrnoContainerActionWaitTimeout`。
  - 持有鎖超過 10 秒未釋放：自動釋放，避免死鎖。

### 4. Graceful Shutdown
- `main.go` 監聽 `SIGINT/SIGTERM`。
- 先 `http.Server.Shutdown()`：停止收新請求，等待 in-flight API 完成。
- 再 cancel job listener context。
- listener 會先 drain queue，並等待正在 create container 的 goroutine 完成後才退出。

### 5. Persistence 與 Runtime 邊界
- metadata（user/file/container/job）存 PostgreSQL（gorm auto-migrate）。
- container life-cycle 透過 Docker API（create/start/stop/delete）。
- 檔案上傳目前是 local object storage + local workspace。
- `docker-compose` 雖提供 MinIO，現行程式尚未切換到 MinIO adapter。

### 6. API 回應模型
- 除 `/healthz` 外，統一回傳：
  - `{ "errno": number, "message": string, "data": any|null }`
- 業務成功與失敗由 `errno` 區分結果，status code 由 errorno 決定（OpenAPI 已同步描述）。

## 本地執行
### 1. 啟動相依服務
```bash
docker compose up -d
```

### 2. 啟動 API Server
```bash
go run ./internal/container
```

預設位址：`http://localhost:8080`  
OpenAPI：`api/openapi/openapi.yaml`

### 3. 跑測試
```bash
(cd internal/common && go test ./...)
(cd internal/container && go test ./...)
```
