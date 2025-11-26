package core

import (
	"context"

	"github.com/archnets/telegram-bot/service"
)

type SubscriptionService struct {
	api *service.APIClient
}

func NewSubscriptionService(api *service.APIClient) *SubscriptionService {
	return &SubscriptionService{api: api}
}

func (s *SubscriptionService) GetStatusText(ctx context.Context, tgID int64) (string, error) {
	return s.api.GetStatus(ctx, tgID)
}

func (s *SubscriptionService) GetConfigsText(ctx context.Context, tgID int64) (string, error) {
	return s.api.GetConfigs(ctx, tgID)
}
