package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles HTTP requests to the ArchNet backend API.
type Client struct {
	http    *http.Client
	baseURL string
}

// NewClient creates a new API client.
func NewClient(baseURL string, timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		http:    &http.Client{Timeout: timeout},
		baseURL: baseURL,
	}
}

// Response is the standard API response wrapper.
type Response struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// Error represents an API error.
type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("api error %d: %s", e.Code, e.Message)
}

// --- Request Helpers ---

// doRequest performs an HTTP request and decodes the response.
func (c *Client) doRequest(ctx context.Context, method, path string, body any, auth string) (*Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var apiResp Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !IsSuccess(apiResp.Code) {
		return &apiResp, &Error{Code: apiResp.Code, Message: apiResp.Message}
	}

	return &apiResp, nil
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path, auth string) (*Response, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil, auth)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, body any, auth string) (*Response, error) {
	return c.doRequest(ctx, http.MethodPost, path, body, auth)
}

// Put performs a PUT request.
func (c *Client) Put(ctx context.Context, path string, body any, auth string) (*Response, error) {
	return c.doRequest(ctx, http.MethodPut, path, body, auth)
}

// --- User Types ---

// UserInfo represents user data from the API.
type UserInfo struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Lang      string `json:"lang"`
	ReferCode string `json:"refer_code"`
	Balance   int64  `json:"balance"`
}

// --- User API Methods ---

// GetUserInfo fetches user info using their JWT token.
func (c *Client) GetUserInfo(ctx context.Context, token string) (*UserInfo, error) {
	resp, err := c.Get(ctx, EndpointUserInfo, token)
	if err != nil {
		return nil, err
	}

	var user UserInfo
	if err := json.Unmarshal(resp.Data, &user); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}
	return &user, nil
}

// UpdateUserLanguage updates the user's language preference.
func (c *Client) UpdateUserLanguage(ctx context.Context, token, lang string) error {
	_, err := c.Put(ctx, EndpointUserLang, map[string]string{"lang": lang}, token)
	return err
}

// --- Subscription Types ---

// SubscribeDiscount represents quantity-based pricing tiers.
type SubscribeDiscount struct {
	Quantity int64 `json:"quantity"` // Number of periods (e.g., 1, 3, 6, 12 months)
	Discount int64 `json:"discount"` // Discount percentage (0-100)
}

// SubscribePlan represents an available subscription plan (from /v1/public/subscribe/list).
type SubscribePlan struct {
	ID          int64               `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	UnitPrice   int64               `json:"unit_price"`
	UnitTime    string              `json:"unit_time"` // "month", "year", etc.
	Traffic     int64               `json:"traffic"`   // bytes
	DeviceLimit int64               `json:"device_limit"`
	SpeedLimit  int64               `json:"speed_limit"`
	Discount    []SubscribeDiscount `json:"discount"` // Quantity-based discounts
}

// UserSubscription represents a user's active subscription (from /v1/public/user/subscribe).
type UserSubscription struct {
	ID            int64  `json:"id"`
	SubscribeID   int64  `json:"subscribe_id"`
	SubscribeName string `json:"subscribe_name"`
	Status        int    `json:"status"`
	ExpiredAt     int64  `json:"expired_at"`
	Traffic       int64  `json:"traffic"`  // total bytes
	Upload        int64  `json:"upload"`   // used bytes
	Download      int64  `json:"download"` // used bytes
}

// AffiliateCount represents affiliate/referral statistics (from /v1/public/user/affiliate/count).
type AffiliateCount struct {
	Registers       int64 `json:"registers"`
	TotalCommission int64 `json:"total_commission"`
}

// --- Subscription API Methods ---

// GetSubscribePlans fetches available subscription plans.
func (c *Client) GetSubscribePlans(ctx context.Context, token, lang string) ([]SubscribePlan, error) {
	path := EndpointSubscribeList + "?language=" + lang
	resp, err := c.Get(ctx, path, token)
	if err != nil {
		return nil, err
	}

	var result struct {
		List []SubscribePlan `json:"list"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal plans: %w", err)
	}
	return result.List, nil
}

// GetUserSubscriptions fetches user's active subscriptions.
func (c *Client) GetUserSubscriptions(ctx context.Context, token string) ([]UserSubscription, error) {
	resp, err := c.Get(ctx, EndpointUserSubscribe, token)
	if err != nil {
		return nil, err
	}

	var result struct {
		List []UserSubscription `json:"list"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal subscriptions: %w", err)
	}
	return result.List, nil
}

// GetAffiliateCount fetches user's affiliate statistics.
func (c *Client) GetAffiliateCount(ctx context.Context, token string) (*AffiliateCount, error) {
	resp, err := c.Get(ctx, EndpointAffiliateCount, token)
	if err != nil {
		return nil, err
	}

	var result AffiliateCount
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal affiliate: %w", err)
	}
	return &result, nil
}

// --- Payment Types ---

// PaymentMethod represents a payment option.
type PaymentMethod struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Platform    string `json:"platform"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	FeeMode     int    `json:"fee_mode"`
	FeePercent  int64  `json:"fee_percent"`
	FeeAmount   int64  `json:"fee_amount"`
}

// GetPaymentMethods fetches available payment methods.
func (c *Client) GetPaymentMethods(ctx context.Context, token string) ([]PaymentMethod, error) {
	resp, err := c.Get(ctx, EndpointPaymentMethods, token)
	if err != nil {
		return nil, err
	}

	var result struct {
		List []PaymentMethod `json:"list"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal payment methods: %w", err)
	}
	return result.List, nil
}

// --- Order Types ---

// PreOrderRequest is used to preview order pricing.
type PreOrderRequest struct {
	SubscribeID int64  `json:"subscribe_id"`
	Quantity    int64  `json:"quantity"`
	Payment     int64  `json:"payment,omitempty"`
	Coupon      string `json:"coupon,omitempty"`
}

// PreOrderResponse contains order price preview.
type PreOrderResponse struct {
	Price          int64  `json:"price"`
	Amount         int64  `json:"amount"`
	Discount       int64  `json:"discount"`
	GiftAmount     int64  `json:"gift_amount"`
	Coupon         string `json:"coupon"`
	CouponDiscount int64  `json:"coupon_discount"`
	FeeAmount      int64  `json:"fee_amount"`
}

// PreOrder previews order pricing before purchase.
func (c *Client) PreOrder(ctx context.Context, token string, req PreOrderRequest) (*PreOrderResponse, error) {
	resp, err := c.Post(ctx, EndpointOrderPre, req, token)
	if err != nil {
		return nil, err
	}

	var result PreOrderResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal pre-order: %w", err)
	}
	return &result, nil
}

// PurchaseOrderRequest creates a new order.
type PurchaseOrderRequest struct {
	SubscribeID int64  `json:"subscribe_id"`
	Quantity    int64  `json:"quantity"`
	Payment     int64  `json:"payment"`
	Coupon      string `json:"coupon,omitempty"`
	CustomName  string `json:"custom_name,omitempty"`
}

// PurchaseOrderResponse contains the created order number.
type PurchaseOrderResponse struct {
	OrderNo string `json:"order_no"`
}

// PurchaseOrder creates a new subscription order.
func (c *Client) PurchaseOrder(ctx context.Context, token string, req PurchaseOrderRequest) (*PurchaseOrderResponse, error) {
	resp, err := c.Post(ctx, EndpointOrderPurchase, req, token)
	if err != nil {
		return nil, err
	}

	var result PurchaseOrderResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal purchase order: %w", err)
	}
	return &result, nil
}

// OrderDetail contains order information.
type OrderDetail struct {
	OrderNo    string `json:"order_no"`
	Status     int    `json:"status"`
	Amount     int64  `json:"amount"`
	PaymentURL string `json:"payment_url"`
}

// GetOrderDetail fetches order details by order number.
func (c *Client) GetOrderDetail(ctx context.Context, token, orderNo string) (*OrderDetail, error) {
	path := EndpointOrderDetail + "?order_no=" + orderNo
	resp, err := c.Get(ctx, path, token)
	if err != nil {
		return nil, err
	}

	var result OrderDetail
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal order detail: %w", err)
	}
	return &result, nil
}

// CheckoutOrderRequest contains parameters for processing order payment.
type CheckoutOrderRequest struct {
	OrderNo   string `json:"orderNo"`
	ReturnURL string `json:"returnUrl,omitempty"`
}

// CheckoutOrderResponse contains checkout result.
type CheckoutOrderResponse struct {
	Type        string `json:"type"`         // "url", "qr", "stripe", or empty for balance
	CheckoutURL string `json:"checkout_url"` // Payment URL for external payments
}

// CheckoutOrder processes order payment (deducts balance for balance payments).
func (c *Client) CheckoutOrder(ctx context.Context, token string, req CheckoutOrderRequest) (*CheckoutOrderResponse, error) {
	resp, err := c.Post(ctx, EndpointOrderCheckout, req, token)
	if err != nil {
		return nil, err
	}

	var result CheckoutOrderResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal checkout order: %w", err)
	}
	return &result, nil
}
