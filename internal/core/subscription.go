package core

// SubscriptionService is a placeholder for subscription-related operations.
type SubscriptionService struct{}

// NewSubscriptionService creates a new subscription service.
// The `_ any` parameter is for backwards compatibility.
func NewSubscriptionService(_ any) *SubscriptionService {
	return &SubscriptionService{}
}

// GetStatusText returns subscription status (placeholder).
func (s *SubscriptionService) GetStatusText(tgID int64) string {
	return "Status feature coming soon"
}
