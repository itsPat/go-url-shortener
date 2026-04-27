package links

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("link not found")
var ErrCodeTaken = errors.New("code already taken")

type Link struct {
	Code      string `gorm:"primaryKey;size:16"`
	URL       string `gorm:"not null"`
	Hits      int64  `gorm:"not null;default:0"`
	CreatedAt time.Time
}

type Store interface {
	Create(ctx context.Context, code, url string) (Link, error)
	Get(ctx context.Context, code string) (Link, error)
	GetAndIncrement(ctx context.Context, code string) (Link, error)
}
