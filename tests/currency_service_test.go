package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"flypro-assessment-ayo-rt/internal/config"
	"flypro-assessment-ayo-rt/internal/services"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestCurrencyService_ConvertCurrency(t *testing.T) {
	tests := []struct {
		name          string
		amount        float64
		from          string
		to            string
		mockResponse  map[string]interface{}
		expectedRate  float64
		expectedError bool
	}{
		{
			name:   "same currency",
			amount: 100.0,
			from:   "USD",
			to:     "USD",
		},
		{
			name:   "USD to EUR conversion",
			amount: 100.0,
			from:   "USD",
			to:     "EUR",
			mockResponse: map[string]interface{}{
				"base": "USD",
				"date": "2024-01-01",
				"rates": map[string]float64{
					"EUR": 0.85,
				},
			},
			expectedRate: 0.85,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockServer *httptest.Server
			if tt.mockResponse != nil {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}))
				defer mockServer.Close()
			}

			cfg := &config.Config{
				Currency: config.CurrencyConfig{
					APIURL: mockServer.URL,
				},
			}

			redisClient := redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			})

			service := services.NewCurrencyService(redisClient, cfg)
			result, err := service.ConvertCurrency(context.Background(), tt.amount, tt.from, tt.to)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				if tt.from == tt.to {
					assert.NoError(t, err)
					assert.Equal(t, tt.amount, result)
				} else {
					assert.NoError(t, err)
					assert.Greater(t, result, 0.0)
				}
			}
		})
	}
}

func TestCurrencyService_GetExchangeRate(t *testing.T) {
	tests := []struct {
		name          string
		from          string
		to            string
		expectedError bool
	}{
		{
			name: "same currency returns 1.0",
			from: "USD",
			to:   "USD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Currency: config.CurrencyConfig{
					APIURL: "https://api.exchangerate-api.com/v4/latest",
				},
			}

			redisClient := redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			})

			service := services.NewCurrencyService(redisClient, cfg)
			rate, err := service.GetExchangeRate(context.Background(), tt.from, tt.to)

			if tt.from == tt.to {
				assert.NoError(t, err)
				assert.Equal(t, 1.0, rate)
			} else if !tt.expectedError {
				assert.NoError(t, err)
				assert.Greater(t, rate, 0.0)
			}
		})
	}
}
