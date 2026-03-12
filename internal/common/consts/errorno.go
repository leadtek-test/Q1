package consts

const (
	ErrnoSuccess      = 0
	ErrnoUnknownError = 1

	// param error 1xxx
	ErrnoBindRequestError     = 1000
	ErrnoRequestValidateError = 1001

	// mysql error 2xxx

	// auth error 3xxx
	ErrnoAuthMissingAuthorizationHeader = 3000
	ErrnoAuthInvalidAuthorizationHeader = 3001
	ErrnoAuthTokenManagerUnavailable    = 3002
	ErrnoAuthInvalidToken               = 3003
	ErrnoAuthJWTSecretNotConfigured     = 3004
	ErrnoAuthJWTInvalidExpireConfig     = 3005
	ErrnoAuthJWTSignFailed              = 3006
	ErrnoAuthJWTMalformedToken          = 3007
	ErrnoAuthJWTSignatureInvalid        = 3008
	ErrnoAuthJWTExpiredToken            = 3009
	ErrnoAuthJWTInvalidClaims           = 3010
	ErrnoAuthJWTParseFailed             = 3011

	// user error 4xxx
	ErrnoUserUsernameRequired = 4000
	ErrnoUserPasswordRequired = 4001
	ErrnoUserPasswordTooShort = 4002
	ErrnoUserAlreadyExists    = 4003
	ErrnoUserNotFound         = 4004
	ErrnoUserPasswordNotMatch = 4005

	// file error 6xxx
	ErrnoFileRequired          = 6000
	ErrnoFileNameRequired      = 6001
	ErrnoFileOpenFailed        = 6002
	ErrnoFileReadFailed        = 6003
	ErrnoFileSizeExceeded      = 6004
	ErrnoFileUploadFailed      = 6005
	ErrnoFileWorkspaceSaveFail = 6006

	// container error 7xxx
	ErrnoContainerImageRequired        = 7000
	ErrnoContainerWorkspacePrepareFail = 7001
	ErrnoContainerRuntimeCreateFail    = 7002
	ErrnoContainerNotFound             = 7003
	ErrnoContainerInvalidStatusAction  = 7004
	ErrnoContainerRuntimeStartFail     = 7005
	ErrnoContainerRuntimeStopFail      = 7006
	ErrnoContainerRuntimeDeleteFail    = 7007
	ErrnoContainerCreateJobQueueFull   = 7008
	ErrnoContainerCreateJobNotFound    = 7009
	ErrnoContainerCreateJobUnavailable = 7010
	ErrnoContainerActionWaitTimeout    = 7011

	// database error 5xxx
	ErrnoDatabaseError = 5000
)

var ErrMsg = map[int]string{
	ErrnoSuccess:      "success",
	ErrnoUnknownError: "unknown error",

	ErrnoBindRequestError:     "bind request error",
	ErrnoRequestValidateError: "validate request error",

	ErrnoAuthMissingAuthorizationHeader: "missing Authorization header",
	ErrnoAuthInvalidAuthorizationHeader: "invalid Authorization header format",
	ErrnoAuthTokenManagerUnavailable:    "auth token manager unavailable",
	ErrnoAuthInvalidToken:               "invalid token",
	ErrnoAuthJWTSecretNotConfigured:     "jwt secret is not configured",
	ErrnoAuthJWTInvalidExpireConfig:     "jwt expire config is invalid",
	ErrnoAuthJWTSignFailed:              "jwt token sign failed",
	ErrnoAuthJWTMalformedToken:          "jwt token is malformed",
	ErrnoAuthJWTSignatureInvalid:        "jwt token signature is invalid",
	ErrnoAuthJWTExpiredToken:            "jwt token is expired",
	ErrnoAuthJWTInvalidClaims:           "jwt token claims is invalid",
	ErrnoAuthJWTParseFailed:             "jwt token parse failed",

	ErrnoUserUsernameRequired: "username is required",
	ErrnoUserPasswordRequired: "password is required",
	ErrnoUserPasswordTooShort: "password must be at least 6 characters",
	ErrnoUserAlreadyExists:    "username already exists",
	ErrnoUserNotFound:         "user not found",
	ErrnoUserPasswordNotMatch: "password not match",

	ErrnoFileRequired:          "file is required",
	ErrnoFileNameRequired:      "file name is required",
	ErrnoFileOpenFailed:        "failed to open file",
	ErrnoFileReadFailed:        "failed to read file",
	ErrnoFileSizeExceeded:      "file size exceeds 20MB limit",
	ErrnoFileUploadFailed:      "failed to upload file",
	ErrnoFileWorkspaceSaveFail: "failed to save file in workspace",

	ErrnoContainerImageRequired:        "container image is required",
	ErrnoContainerWorkspacePrepareFail: "failed to prepare container workspace",
	ErrnoContainerRuntimeCreateFail:    "failed to create container runtime",
	ErrnoContainerNotFound:             "container not found",
	ErrnoContainerInvalidStatusAction:  "invalid container status action",
	ErrnoContainerRuntimeStartFail:     "failed to start container runtime",
	ErrnoContainerRuntimeStopFail:      "failed to stop container runtime",
	ErrnoContainerRuntimeDeleteFail:    "failed to delete container runtime",
	ErrnoContainerCreateJobQueueFull:   "container create job queue is full",
	ErrnoContainerCreateJobNotFound:    "container create job not found",
	ErrnoContainerCreateJobUnavailable: "container create job service unavailable",
	ErrnoContainerActionWaitTimeout:    "等待超時（資源已被佔用）請稍後重試",

	ErrnoDatabaseError: "database error",
}
