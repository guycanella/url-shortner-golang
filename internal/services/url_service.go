package services

import (
	"fmt"
	"guycanella-url-shortner/internal/cache"
	"guycanella-url-shortner/internal/config"
	"guycanella-url-shortner/internal/models"
	"guycanella-url-shortner/internal/pkg/utils"
	"guycanella-url-shortner/internal/repository"
	"net"
	"net/url"
	"strings"
	"time"
)

type URLService interface {
	CreateURL(req models.CreateURLRequest) (*models.CreateURLResponse, error)
	GetOriginalURL(shortCode string) (*models.URL, error)
	GetURLStats(shortCode string) (*models.URLStatsResponse, error)
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
	// VALIDATION 1: NORMALIZE URLs
	if !strings.HasPrefix(req.URL, "http://") ||
		!strings.HasPrefix(req.URL, "https://") {
		req.URL = "http://" + req.URL
	}

	// VALIDATION 2: URL PARSING
	parsedUrl, erro := url.Parse(req.URL)
	if erro != nil {
		return nil, fmt.Errorf("invalid URL: %w", erro)
	}

	// VALIDATION 3: BLOCK PRIVATE IPs
	ipAddrs, _ := net.LookupIP(parsedUrl.Hostname())
	for _, ip := range ipAddrs {
		if ip.IsLoopback() || ip.IsPrivate() {
			return nil, fmt.Errorf("forbidden URL: private or local address")
		}
	}

	// VALIDATION 4: DOMAINS BLACKLIST
	blackList := []string{"malware.com", "phishing.site", "spam.example"}
	hostname := parsedUrl.Hostname()
	for _, banned := range blackList {
		if strings.Contains(hostname, banned) {
			return nil, fmt.Errorf("forbidden URL: blacklisted domain")
		}
	}

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
	cacheKey := "short:" + shortCode

	val, err := cache.Client.Get(cache.Ctx, cacheKey).Result()
	if err == nil {
		return &models.URL{ShortCode: shortCode, OriginalUrl: val}, nil
	}

	url, err := service.repo.FindByShortCode(shortCode)
	if err != nil {
		return nil, err
	}

	if !url.IsValid() {
		return nil, fmt.Errorf("URL expired or inactive")
	}

	ttl := time.Until(url.ExpiresAt)
	if ttl > 0 {
		cache.Client.Set(cache.Ctx, cacheKey, url.OriginalUrl, ttl)
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

	url, err := service.repo.FindByID(uuid)
	if err != nil {
		return err
	}

	if err := service.repo.Delete(uuid); err != nil {
		return err
	}

	cacheKey := "short:" + url.ShortCode
	if err := cache.Client.Del(cache.Ctx, cacheKey).Err(); err != nil {
		fmt.Printf("⚠️ erro ao remover cache da URL %s: %v\n", url.ShortCode, err)
	}

	return nil
}

func (service *urlService) DeactivateExpired() (int64, error) {
	expired, err := service.repo.FindExpiredURLs()
	if err != nil {
		return 0, err
	}

	affected, err := service.repo.DeactivateExpiredURLs()
	if err != nil {
		return 0, err
	}

	for _, exp := range expired {
		cacheKey := "short:" + exp.ShortCode

		if err := cache.Client.Del(cache.Ctx, cacheKey).Err(); err != nil {
			fmt.Printf("⚠️ erro ao remover cache expirada %s: %v\n", exp.ShortCode, err)
		}
	}

	return affected, nil
}
