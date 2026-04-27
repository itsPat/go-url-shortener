package links

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{db: db}
}

func (s *Database) Create(ctx context.Context, code, url string) (Link, error) {
	link := Link{Code: code, URL: url}

	err := s.db.WithContext(ctx).Create(&link).Error

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return Link{}, ErrCodeTaken
	}

	if err != nil {
		return Link{}, err
	}

	return link, nil
}

func (s *Database) Get(ctx context.Context, code string) (Link, error) {
	var link Link
	err := s.db.WithContext(ctx).Where("code = ?", code).First(&link).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Link{}, ErrNotFound
	}

	if err != nil {
		return Link{}, err
	}

	return link, nil
}

func (s *Database) GetAndIncrement(ctx context.Context, code string) (Link, error) {
	result := s.db.WithContext(ctx).
		Model(&Link{}).
		Where("code = ?", code).
		Update("hits", gorm.Expr("hits + 1"))

	if result.Error != nil {
		return Link{}, result.Error
	}

	if result.RowsAffected == 0 {
		return Link{}, ErrNotFound
	}

	return s.Get(ctx, code)
}

var _ Store = (*Database)(nil)
