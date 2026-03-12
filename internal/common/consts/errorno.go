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

	ErrnoDatabaseError: "database error",
}
