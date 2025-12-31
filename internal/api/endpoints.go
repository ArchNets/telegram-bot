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
	EndpointUserSubscribe = "/v1/public/user/subscribe"

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

)
