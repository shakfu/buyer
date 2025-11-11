package services

import (
	"testing"
	"time"
)

func TestForexService_Create(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	svc := NewForexService(cfg.DB)

	tests := []struct {
		name    string
		from    string
		to      string
		rate    float64
		wantErr bool
	}{
		{
			name:    "valid rate",
			from:    "EUR",
			to:      "USD",
			rate:    1.18,
			wantErr: false,
		},
		{
			name:    "invalid from currency",
			from:    "EU",
			to:      "USD",
			rate:    1.18,
			wantErr: true,
		},
		{
			name:    "invalid to currency",
			from:    "EUR",
			to:      "US",
			rate:    1.18,
			wantErr: true,
		},
		{
			name:    "negative rate",
			from:    "GBP",
			to:      "USD",
			rate:    -1.5,
			wantErr: true,
		},
		{
			name:    "zero rate",
			from:    "GBP",
			to:      "USD",
			rate:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			forex, err := svc.Create(tt.from, tt.to, tt.rate, time.Now())

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if forex.FromCurrency != tt.from {
					t.Errorf("Expected from %s, got %s", tt.from, forex.FromCurrency)
				}
				if forex.ToCurrency != tt.to {
					t.Errorf("Expected to %s, got %s", tt.to, forex.ToCurrency)
				}
				if forex.Rate != tt.rate {
					t.Errorf("Expected rate %f, got %f", tt.rate, forex.Rate)
				}
			}
		})
	}
}

func TestForexService_GetLatestRate(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	svc := NewForexService(cfg.DB)

	// Create some rates
	svc.Create("EUR", "USD", 1.18, time.Now().Add(-24*time.Hour))
	svc.Create("EUR", "USD", 1.20, time.Now())

	tests := []struct {
		name     string
		from     string
		to       string
		wantRate float64
		wantErr  bool
	}{
		{
			name:     "existing rate",
			from:     "EUR",
			to:       "USD",
			wantRate: 1.20, // Should get the latest rate
			wantErr:  false,
		},
		{
			name:     "same currency",
			from:     "USD",
			to:       "USD",
			wantRate: 1.0,
			wantErr:  false,
		},
		{
			name:    "non-existent rate",
			from:    "JPY",
			to:      "USD",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			forex, err := svc.GetLatestRate(tt.from, tt.to)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if forex.Rate != tt.wantRate {
					t.Errorf("Expected rate %f, got %f", tt.wantRate, forex.Rate)
				}
			}
		})
	}
}

func TestForexService_Convert(t *testing.T) {
	cfg := setupTestDB(t)
	defer cfg.Close()

	svc := NewForexService(cfg.DB)

	// Create a rate
	svc.Create("EUR", "USD", 1.20, time.Now())

	tests := []struct {
		name           string
		amount         float64
		from           string
		to             string
		wantConverted  float64
		wantErr        bool
	}{
		{
			name:          "valid conversion",
			amount:        100,
			from:          "EUR",
			to:            "USD",
			wantConverted: 120,
			wantErr:       false,
		},
		{
			name:          "same currency",
			amount:        100,
			from:          "USD",
			to:            "USD",
			wantConverted: 100,
			wantErr:       false,
		},
		{
			name:    "no rate available",
			amount:  100,
			from:    "JPY",
			to:      "USD",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converted, rate, err := svc.Convert(tt.amount, tt.from, tt.to)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if converted != tt.wantConverted {
					t.Errorf("Expected converted amount %f, got %f", tt.wantConverted, converted)
				}
				if rate <= 0 {
					t.Errorf("Expected positive rate, got %f", rate)
				}
			}
		})
	}
}
