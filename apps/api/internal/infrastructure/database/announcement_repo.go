package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"dpmptsp/api/internal/domain"
)

type announcementModel struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Judul     string    `gorm:"column:judul"`
	Deskripsi string    `gorm:"column:deskripsi"`
	Foto      string    `gorm:"column:foto"`
	LinkURL   string    `gorm:"column:link_url"`
	IsActive  bool      `gorm:"column:is_active"`
	Urutan    int       `gorm:"column:urutan"`
	Tipe      string    `gorm:"column:tipe"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (announcementModel) TableName() string { return "pemberitahuan" }

func (m announcementModel) toDomain() domain.Announcement {
	return domain.Announcement{
		ID: m.ID, Judul: m.Judul, Deskripsi: m.Deskripsi, Foto: m.Foto,
		LinkURL: m.LinkURL, IsActive: m.IsActive, Urutan: m.Urutan, Tipe: m.Tipe,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}
}

type AnnouncementRepo struct{ db *gorm.DB }

func NewAnnouncementRepo(db *gorm.DB) *AnnouncementRepo { return &AnnouncementRepo{db: db} }

var _ domain.AnnouncementRepository = (*AnnouncementRepo)(nil)

func (r *AnnouncementRepo) List(ctx context.Context, f domain.AnnouncementFilter) ([]domain.Announcement, error) {
	q := r.db.WithContext(ctx).Model(&announcementModel{})
	if f.Tipe != "" {
		q = q.Where("tipe = ?", f.Tipe)
	}
	if f.ActiveOnly {
		q = q.Where("is_active = ?", true)
	}

	var rows []announcementModel
	if err := q.Order("urutan ASC").Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list announcements: %w", err)
	}
	out := make([]domain.Announcement, 0, len(rows))
	for _, m := range rows {
		out = append(out, m.toDomain())
	}
	return out, nil
}

func (r *AnnouncementRepo) ByID(ctx context.Context, id int64) (*domain.Announcement, error) {
	var m announcementModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find announcement: %w", err)
	}
	a := m.toDomain()
	return &a, nil
}

func (r *AnnouncementRepo) Create(ctx context.Context, a *domain.Announcement) error {
	now := time.Now()
	m := announcementModel{
		Judul: a.Judul, Deskripsi: a.Deskripsi, Foto: a.Foto, LinkURL: a.LinkURL,
		IsActive: a.IsActive, Urutan: a.Urutan, Tipe: a.Tipe,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return fmt.Errorf("create announcement: %w", err)
	}
	a.ID = m.ID
	return nil
}

func (r *AnnouncementRepo) Update(ctx context.Context, a *domain.Announcement) error {
	res := r.db.WithContext(ctx).Model(&announcementModel{}).
		Where("id = ?", a.ID).
		Select("judul", "deskripsi", "foto", "link_url", "is_active", "urutan", "tipe", "updated_at").
		Updates(announcementModel{
			Judul: a.Judul, Deskripsi: a.Deskripsi, Foto: a.Foto, LinkURL: a.LinkURL,
			IsActive: a.IsActive, Urutan: a.Urutan, Tipe: a.Tipe, UpdatedAt: time.Now(),
		})
	if res.Error != nil {
		return fmt.Errorf("update announcement: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AnnouncementRepo) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&announcementModel{})
	if res.Error != nil {
		return fmt.Errorf("delete announcement: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
