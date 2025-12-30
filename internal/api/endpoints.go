// Package api provides API client types, endpoints, and error codes.
package api

// API Endpoints
const (
	// Auth endpoints
	EndpointLogin         = "/v1/auth/login"
	EndpointLoginTelegram = "/v1/auth/login/telegram"
	EndpointLogout        = "/v1/auth/logout"

	// User endpoints
	EndpointUserInfo     = "/v1/public/user/info"
	EndpointUserLang     = "/v1/public/user/lang"
	EndpointUserPassword = "/v1/public/user/password"
	EndpointUserNotify   = "/v1/public/user/notify"
	EndpointUserDevices  = "/v1/public/user/devices"

	// Subscription endpoints
	EndpointUserSubscribe     = "/v1/public/user/subscribe"
	EndpointUserSubscribeList = "/v1/public/user/subscribe/list"
	EndpointUserTrafficLog    = "/v1/public/user/traffic_log"

	// Ticket endpoints
	EndpointTicket       = "/v1/public/ticket/"
	EndpointTicketList   = "/v1/public/ticket/list"
	EndpointTicketDetail = "/v1/public/ticket/detail"

	// Binding endpoints
	EndpointBindTelegram = "/v1/public/user/bind_telegram"
	EndpointBindEmail    = "/v1/public/user/bind_email"
	EndpointBindMobile   = "/v1/public/user/bind_mobile"
	EndpointBindOAuth    = "/v1/public/user/bind_oauth"

	// Subscribe catalog endpoints
	EndpointSubscribeList = "/v1/public/subscribe/list"

	// Affiliate endpoints
	EndpointAffiliateCount = "/v1/public/user/affiliate/count"

	// Payment endpoints
	EndpointPaymentMethods = "/v1/public/payment/methods"

	// Order endpoints
	EndpointOrderPre      = "/v1/public/order/pre"
	EndpointOrderPurchase = "/v1/public/order/purchase"
	EndpointOrderDetail   = "/v1/public/order/detail"
	EndpointOrderCheckout = "/v1/public/portal/order/checkout"
)
