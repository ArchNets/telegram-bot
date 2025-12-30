// Package api provides API client types and error codes.
package api

// Response codes from the ArchNet backend.
const (
	// General
	Success         = 200
	CodeError       = 500
	InvalidParams   = 400
	TooManyRequests = 401

	// Authentication & Token
	ErrorTokenEmpty   = 40002
	ErrorTokenInvalid = 40003
	ErrorTokenExpire  = 40004
	InvalidAccess     = 40005
	InvalidCiphertext = 40006
	SecretIsEmpty     = 40007

	// User Errors
	UserExist               = 20001
	UserNotExist            = 20002
	UserPasswordError       = 20003
	UserDisabled            = 20004
	InsufficientBalance     = 20005
	StopRegister            = 20006
	TelegramNotBound        = 20007
	UserNotBindOauth        = 20008
	InviteCodeError         = 20009
	UserCommissionNotEnough = 20010

	// Database Errors
	DatabaseQueryError   = 10001
	DatabaseUpdateError  = 10002
	DatabaseInsertError  = 10003
	DatabaseDeletedError = 10004

	// Subscription Errors
	SubscribeExpired                = 60001
	SubscribeNotAvailable           = 60002
	UserSubscribeExist              = 60003
	SubscribeIsUsedError            = 60004
	SingleSubscribeModeExceedsLimit = 60005
	SubscribeQuotaLimit             = 60006

	// Order & Payment Errors
	OrderNotExist         = 61001
	PaymentMethodNotFound = 61002
	OrderStatusError      = 61003
	InsufficientOfPeriod  = 61004
	ExistAvailableTraffic = 61005

	// Order Status Codes
	OrderStatusPending  = 1 // Order created but not paid (Pending)
	OrderStatusPaid     = 2 // Order paid and ready for processing
	OrderStatusClose    = 3 // Order closed/cancelled
	OrderStatusFailed   = 4 // Order processing failed
	OrderStatusFinished = 5 // Order successfully completed (Finished)

	// Coupon Errors
	CouponNotExist          = 50001
	CouponAlreadyUsed       = 50002
	CouponNotApplicable     = 50003
	CouponInsufficientUsage = 50004
	CouponExpired           = 50005

	// Node Errors
	NodeExist         = 30001
	NodeNotExist      = 30002
	NodeGroupExist    = 30003
	NodeGroupNotExist = 30004
	NodeGroupNotEmpty = 30005

	// Verification & SMS Errors
	VerifyCodeError            = 70001
	SendSmsError               = 90002
	SmsNotEnabled              = 90003
	EmailNotEnabled            = 90004
	TodaySendCountExceedsLimit = 90015

	// Device & Binding Errors
	TelephoneAreaCodeIsEmpty           = 90007
	PasswordIsEmpty                    = 90008
	AreaCodeIsEmpty                    = 90009
	PasswordOrVerificationCodeRequired = 90010
	EmailExist                         = 90011
	TelephoneExist                     = 90012
	DeviceExist                        = 90013
	TelephoneError                     = 90014
	DeviceNotExist                     = 90017
	UseridNotMatch                     = 90018
)

// IsSuccess returns true if the code indicates success.
func IsSuccess(code int) bool {
	return code == Success
}

// IsAuthError returns true if the code is an authentication error.
func IsAuthError(code int) bool {
	return code >= 40002 && code <= 40007
}

// IsUserError returns true if the code is a user-related error.
func IsUserError(code int) bool {
	return code >= 20001 && code <= 20010
}
