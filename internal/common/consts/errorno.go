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
}
