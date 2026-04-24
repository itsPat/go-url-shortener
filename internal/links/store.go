package links

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("link not found")
var ErrCodeTaken = errors.New("code already taken")

type Link struct {
	Code      string
	URL       string
	Hits      int64
	CreatedAt time.Time
}

type Store interface {
	Create(ctx context.Context, code, url string) (Link, error)
	Get(ctx context.Context, code string) (Link, error)
	GetAndIncrement(ctx context.Context, code string) (Link, error)
}
