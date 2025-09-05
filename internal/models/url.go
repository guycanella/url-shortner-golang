package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type URL struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ShortCode   string    `json:"shortCode" gorm:"uniqueIndex;not null"`
	OriginalUrl string    `json:"originalUrl" gorm:"not null"`
	ClickCount  int64     `json:"clickCount" gorm:"default:0"`
	IsActive    bool      `json:"isActive" gorm:"default:true"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

func (url *URL) BeforeCreate(tx *gorm.DB) error {
	if url.ID == uuid.Nil {
		url.ID = uuid.New()
	}

	return nil
}

func (url *URL) IsExpired() bool {
	return time.Now().After(url.ExpiresAt)
}

func (url *URL) IsValid() bool {
	return url.IsActive && !url.IsExpired()
}

func (url *URL) Click() {
	url.ClickCount++
}

type CreateURLRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type CreateURLResponse struct {
	ID        uuid.UUID `json:"id"`
	ShortCode string    `json:"shortCode"`
	ShortURL  string    `json:"shortUrl"`
	LongURL   string    `json:"longUrl"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

type URLStatsResponse struct {
	ID         uuid.UUID `json:"id"`
	ShortCode  string    `json:"shortCode"`
	ShortURL   string    `json:"shortUrl"`
	LongURL    string    `json:"longUrl"`
	ClickCount int64     `json:"clickCount"`
	IsActive   bool      `json:"isActive"`
	IsExpired  bool      `json:"isExpired"`
	CreatedAt  time.Time `json:"createdAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
}
