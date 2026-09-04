package database

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"dpmptsp/api/internal/domain"
)

// SliderRepo manages beranda_slider. The model and row mapping live in
// page_repo.go, which reads the same table for the homepage.
type SliderRepo struct{ db *gorm.DB }

func NewSliderRepo(db *gorm.DB) *SliderRepo { return &SliderRepo{db: db} }

var _ domain.SliderRepository = (*SliderRepo)(nil)

func (r *SliderRepo) List(ctx context.Context) ([]domain.Slider, error) {
	var rows []berandaSliderModel
	if err := r.db.WithContext(ctx).Order("urutan ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list sliders: %w", err)
	}
	out := make([]domain.Slider, 0, len(rows))
	for _, m := range rows {
		out = append(out, toSlider(m.Slider))
	}
	return out, nil
}

func (r *SliderRepo) ByID(ctx context.Context, id int64) (*domain.Slider, error) {
	var m berandaSliderModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find slider: %w", err)
	}
	s := toSlider(m.Slider)
	return &s, nil
}

func (r *SliderRepo) Create(ctx context.Context, s *domain.Slider) error {
	row := berandaSliderModel{Slider: sliderModel{
		Gambar: s.Gambar, Judul: s.Judul, Deskripsi: s.Deskripsi, Urutan: s.Urutan,
	}}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return fmt.Errorf("create slider: %w", err)
	}
	s.ID = row.Slider.ID
	return nil
}

func (r *SliderRepo) Update(ctx context.Context, s *domain.Slider) error {
	res := r.db.WithContext(ctx).Model(&berandaSliderModel{}).Where("id = ?", s.ID).
		Updates(map[string]any{
			"gambar": s.Gambar, "judul": s.Judul,
			"deskripsi": s.Deskripsi, "urutan": s.Urutan,
		})
	if res.Error != nil {
		return fmt.Errorf("update slider: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SliderRepo) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&berandaSliderModel{})
	if res.Error != nil {
		return fmt.Errorf("delete slider: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
