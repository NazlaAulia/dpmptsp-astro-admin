package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"dpmptsp/api/internal/domain"
)

type userModel struct {
	ID           int64      `gorm:"column:id;primaryKey"`
	Username     string     `gorm:"column:username"`
	Email        string     `gorm:"column:email"`
	PasswordHash string     `gorm:"column:password_hash"`
	Role         string     `gorm:"column:role"`
	IsActive     bool       `gorm:"column:is_active"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
}

func (userModel) TableName() string { return "users" }

type UserRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

var _ domain.UserRepository = (*UserRepo)(nil)

func (r *UserRepo) ByUsername(ctx context.Context, username string) (*domain.User, error) {
	var m userModel
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	return &domain.User{
		ID: m.ID, Username: m.Username, Email: m.Email, Role: m.Role,
		IsActive: m.IsActive, LastLoginAt: m.LastLoginAt, PasswordHash: m.PasswordHash,
	}, nil
}

func (r *UserRepo) TouchLogin(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&userModel{}).
		Where("id = ?", id).UpdateColumn("last_login_at", now).Error
}

func (r *UserRepo) UpdatePasswordHash(ctx context.Context, id int64, hash string) error {
	return r.db.WithContext(ctx).Model(&userModel{}).
		Where("id = ?", id).UpdateColumn("password_hash", hash).Error
}
