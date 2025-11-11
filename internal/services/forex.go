package services

import (
	"errors"
	"strings"
	"time"

	"github.com/shakfu/buyer/internal/models"
	"gorm.io/gorm"
)

// ForexService handles business logic for forex rates
type ForexService struct {
	db *gorm.DB
}

// NewForexService creates a new forex service
func NewForexService(db *gorm.DB) *ForexService {
	return &ForexService{db: db}
}

// Create creates a new forex rate
func (s *ForexService) Create(fromCurrency, toCurrency string, rate float64, effectiveDate time.Time) (*models.Forex, error) {
	fromCurrency = strings.ToUpper(strings.TrimSpace(fromCurrency))
	toCurrency = strings.ToUpper(strings.TrimSpace(toCurrency))

	if len(fromCurrency) != 3 {
		return nil, &ValidationError{Field: "from_currency", Message: "currency must be a 3-letter ISO 4217 code"}
	}
	if len(toCurrency) != 3 {
		return nil, &ValidationError{Field: "to_currency", Message: "currency must be a 3-letter ISO 4217 code"}
	}
	if rate <= 0 {
		return nil, &ValidationError{Field: "rate", Message: "rate must be positive"}
	}
	if effectiveDate.IsZero() {
		effectiveDate = time.Now()
	}

	forex := &models.Forex{
		FromCurrency:  fromCurrency,
		ToCurrency:    toCurrency,
		Rate:          rate,
		EffectiveDate: effectiveDate,
	}
	if err := s.db.Create(forex).Error; err != nil {
		return nil, err
	}

	return forex, nil
}

// GetLatestRate retrieves the latest exchange rate for a currency pair
func (s *ForexService) GetLatestRate(fromCurrency, toCurrency string) (*models.Forex, error) {
	fromCurrency = strings.ToUpper(strings.TrimSpace(fromCurrency))
	toCurrency = strings.ToUpper(strings.TrimSpace(toCurrency))

	// If same currency, return rate of 1.0
	if fromCurrency == toCurrency {
		return &models.Forex{
			FromCurrency:  fromCurrency,
			ToCurrency:    toCurrency,
			Rate:          1.0,
			EffectiveDate: time.Now(),
		}, nil
	}

	var forex models.Forex
	err := s.db.Where("from_currency = ? AND to_currency = ?", fromCurrency, toCurrency).
		Order("effective_date DESC").
		First(&forex).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &NotFoundError{Entity: "Forex rate", ID: fromCurrency + "/" + toCurrency}
	}
	if err != nil {
		return nil, err
	}

	return &forex, nil
}

// Convert converts an amount from one currency to another using the latest rate
func (s *ForexService) Convert(amount float64, fromCurrency, toCurrency string) (float64, float64, error) {
	forex, err := s.GetLatestRate(fromCurrency, toCurrency)
	if err != nil {
		return 0, 0, err
	}

	convertedAmount := amount * forex.Rate
	return convertedAmount, forex.Rate, nil
}

// List retrieves all forex rates with optional pagination
func (s *ForexService) List(limit, offset int) ([]models.Forex, error) {
	var rates []models.Forex
	query := s.db.Order("effective_date DESC, from_currency ASC, to_currency ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&rates).Error
	return rates, err
}

// Delete deletes a forex rate by ID
func (s *ForexService) Delete(id uint) error {
	result := s.db.Delete(&models.Forex{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return &NotFoundError{Entity: "Forex", ID: id}
	}
	return nil
}

// Count returns the total number of forex rates
func (s *ForexService) Count() (int64, error) {
	var count int64
	err := s.db.Model(&models.Forex{}).Count(&count).Error
	return count, err
}
