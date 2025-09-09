package services

import (
	"fmt"
	"guycanella-url-shortner/internal/config"
	"guycanella-url-shortner/internal/models"
	"guycanella-url-shortner/internal/pkg/utils"
	"guycanella-url-shortner/internal/repository"
	"time"
)

type URLService interface {
	CreateURL(req models.CreateURLRequest) (*models.CreateURLResponse, error)
	GetOriginalURL(shortCode string) (*models.URL, error)
	GetURLStats(shortCode string) (*models.CreateURLResponse, error)
	DeleteURL(id string) error
	DeactivateExpired() (int64, error)
}

type urlService struct {
	repo repository.URLRepository
	cfg  *config.Config
}

func NewURLService(repo repository.URLRepository, cfg *config.Config) *urlService {
	return &urlService{
		repo: repo,
		cfg:  cfg,
	}
}

func (service *urlService) CreateURL(req models.CreateURLRequest) (*models.CreateURLResponse, error) {
	var shortCode string
	var err error

	for {
		shortCode, err = utils.GenerateShortCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate short code: %w", err)
		}

		exists, err := service.repo.ExistsShortCode(shortCode)
		if err != nil {
			return nil, err
		}

		if !exists {
			break
		}
	}

	expiration := time.Now().Add(time.Duration(service.cfg.App.DefaultExpirationMin) * time.Minute)

	urlModel := &models.URL{
		ShortCode:   shortCode,
		OriginalUrl: req.URL,
		ExpiresAt:   expiration,
		IsActive:    true,
	}

	if err := service.repo.Create(urlModel); err != nil {
		return nil, err
	}

	return &models.CreateURLResponse{
		ID:        urlModel.ID,
		ShortCode: urlModel.ShortCode,
		ShortURL:  fmt.Sprintf("%s/%s", service.cfg.App.BaseURL, urlModel.ShortCode),
		LongURL:   urlModel.OriginalUrl,
		CreatedAt: urlModel.CreatedAt,
		ExpiresAt: urlModel.ExpiresAt,
	}, nil
}

func (service *urlService) GetOriginalURL(shortCode string) (*models.URL, error) {
	url, err := service.repo.FindByShortCode(shortCode)
	if err != nil {
		return nil, err
	}

	if !url.IsValid() {
		return nil, fmt.Errorf("URL expired or inactive")
	}

	if err := service.repo.IncrementClickCount(shortCode); err != nil {
		return nil, err
	}

	return url, nil
}

func (service *urlService) GetURLStats(shortCode string) (*models.URLStatsResponse, error) {
	url, err := service.repo.GetStats(shortCode)
	if err != nil {
		return nil, err
	}

	return &models.URLStatsResponse{
		ID:         url.ID,
		ShortCode:  url.ShortCode,
		ShortURL:   fmt.Sprintf("%s/%s", service.cfg.App.BaseURL, url.ShortCode),
		LongURL:    url.OriginalUrl,
		ClickCount: url.ClickCount,
		IsActive:   url.IsActive,
		IsExpired:  url.IsExpired(),
		CreatedAt:  url.CreatedAt,
		ExpiresAt:  url.ExpiresAt,
	}, nil
}

func (service *urlService) DeleteURL(id string) error {
	uuid, err := utils.ParseUUID(id)

	if err != nil {
		return err
	}

	return service.repo.Delete(uuid)
}

func (service *urlService) DeactivateExpired() (int64, error) {
	affected, err := service.repo.DeactivateExpiredURLs()
	if err != nil {
		return 0, err
	}

	return affected, nil
}
