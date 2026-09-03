package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"dpmptsp/api/internal/domain"
)

type contactModel struct {
	ID         int64     `gorm:"column:id;primaryKey"`
	TicketCode string    `gorm:"column:ticket_code"`
	Nama       string    `gorm:"column:nama"`
	Telepon    string    `gorm:"column:telepon"`
	Email      string    `gorm:"column:email"`
	Kategori   string    `gorm:"column:kategori"`
	Subjek     string    `gorm:"column:subjek"`
	Pesan      string    `gorm:"column:pesan"`
	Status     string    `gorm:"column:status"`
	Catatan    string    `gorm:"column:catatan"`
	IPAddress  string    `gorm:"column:ip_address"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (contactModel) TableName() string { return "contact_messages" }

type ContactRepo struct{ db *gorm.DB }

func NewContactRepo(db *gorm.DB) *ContactRepo { return &ContactRepo{db: db} }

var _ domain.ContactRepository = (*ContactRepo)(nil)

func (r *ContactRepo) Create(ctx context.Context, m *domain.ContactMessage) error {
	row := contactModel{
		TicketCode: m.TicketCode, Nama: m.Nama, Telepon: m.Telepon, Email: m.Email,
		Kategori: m.Kategori, Subjek: m.Subjek, Pesan: m.Pesan,
		Status: m.Status, IPAddress: m.IPAddress,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return fmt.Errorf("create contact message: %w", err)
	}
	m.ID = row.ID
	m.CreatedAt = row.CreatedAt
	return nil
}

func (r *ContactRepo) ByTicket(ctx context.Context, code string) (*domain.ContactMessage, error) {
	var m contactModel
	err := r.db.WithContext(ctx).Where("ticket_code = ?", code).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find ticket: %w", err)
	}
	return &domain.ContactMessage{
		ID: m.ID, TicketCode: m.TicketCode, Nama: m.Nama, Telepon: m.Telepon,
		Email: m.Email, Kategori: m.Kategori, Subjek: m.Subjek, Pesan: m.Pesan,
		Status: m.Status, Catatan: m.Catatan, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}, nil
}

func (r *ContactRepo) CountRecentFromIP(ctx context.Context, ip string, since time.Time) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&contactModel{}).
		Where("ip_address = ? AND created_at >= ?", ip, since).Count(&n).Error
	if err != nil {
		return 0, fmt.Errorf("count recent submissions: %w", err)
	}
	return n, nil
}
