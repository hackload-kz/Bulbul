package handlers

import (
	"bulbul/internal/cache"
	"bulbul/internal/service"
)

type Handlers struct {
	services     *service.Services
	valkeyClient *cache.ValkeyClient
}

func NewHandlers(services *service.Services, valkeyClient *cache.ValkeyClient) *Handlers {
	return &Handlers{
		services:     services,
		valkeyClient: valkeyClient,
	}
}
