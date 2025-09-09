package repository

import (
	"errors"
	"fmt"
	"guycanella-url-shortner/internal/database"
	"guycanella-url-shortner/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type URLRepository interface {
	Create(url *models.URL) error
	FindByShortCode(shortCode string) (*models.URL, error)
	FindByID(id uuid.UUID) (*models.URL, error)
	Update(url *models.URL) error
	Delete(id uuid.UUID) error
	ExistsShortCode(shortCode string) (bool, error)
	IncrementClickCount(shortCode string) error
	FindExpiredURLs() ([]models.URL, error)
	DeactivateExpiredURLs() (int64, error)
	GetStats(shortCode string) (*models.URL, error)
}

type urlRepository struct {
	db *gorm.DB
}

func NewURLRepository() URLRepository {
	return &urlRepository{
		db: database.DB,
	}
}

func (repo *urlRepository) Create(url *models.URL) error {
	if err := repo.db.Create(url).Error; err != nil {
		return fmt.Errorf("failed to create URL: %w", err)
	}

	return nil
}

func (repo *urlRepository) FindByShortCode(shortCode string) (*models.URL, error) {
	var url models.URL

	err := repo.db.
		Where("short_code = ? AND is_active = ?", shortCode, true).
		First(&url).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("URL not found for short code: %s", shortCode)
		}

		return nil, fmt.Errorf("failed to find URL by short code: %w", err)
	}

	return &url, nil
}

func (repo *urlRepository) FindByID(id uuid.UUID) (*models.URL, error) {
	var url models.URL

	err := repo.db.
		Where("id = ?", id).
		First(&url).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("URL not found for ID: %s", id)
		}

		return nil, fmt.Errorf("failed to find URL by ID: %w", err)
	}

	return &url, nil
}

func (repo *urlRepository) Update(url *models.URL) error {
	if err := repo.db.Save(url).Error; err != nil {
		return fmt.Errorf("failed to update URL: %w", err)
	}

	return nil
}

func (repo *urlRepository) Delete(id uuid.UUID) error {
	result := repo.db.
		Model(&models.URL{}).
		Where("id = ?", id).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to delete URL: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("URL not found for ID: %s", id.String())
	}

	return nil
}

func (repo *urlRepository) ExistsShortCode(shortCode string) (bool, error) {
	var count int64

	err := repo.db.Model(&models.URL{}).
		Where("short_code = ?", shortCode).
		Count(&count).
		Error

	if err != nil {
		return false, fmt.Errorf("failed to check if short code exists: %w", err)
	}

	return count > 0, nil
}

func (repo *urlRepository) IncrementClickCount(shortCode string) error {
	result := repo.db.Model(&models.URL{}).
		Where("short_code = ? AND is_active = ?", shortCode, true).
		Update("click_count", gorm.Expr("click_count + ?", 1))

	if result.Error != nil {
		return fmt.Errorf("failed to increment click count: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("URL not found or inactive for short code: %s", shortCode)
	}

	return nil
}

func (repo *urlRepository) FindExpiredURLs() ([]models.URL, error) {
	var urls []models.URL

	err := repo.db.
		Where("expires_at <= ? AND is_active = ?", time.Now(), true).
		Find(&urls).
		Error

	if err != nil {
		return nil, fmt.Errorf("failed to find expired URLs: %w", err)
	}

	return urls, nil
}

func (repo *urlRepository) DeactivateExpiredURLs() (int64, error) {
	result := repo.db.Model(&models.URL{}).
		Where("expires_at < ? AND is_active = ?", time.Now(), true).
		Update("is_active", false)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to deactivate expired URLs: %w", result.Error)
	}

	return result.RowsAffected, nil
}

func (repo *urlRepository) GetStats(shortCode string) (*models.URL, error) {
	var url models.URL

	err := repo.db.Where("short_code = ?", shortCode).First(&url).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("URL not found for short code: %s", shortCode)
		}

		return nil, fmt.Errorf("failed to get URL stats: %w", err)
	}

	return &url, nil
}
