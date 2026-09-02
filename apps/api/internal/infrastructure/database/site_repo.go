package database

import (
	"errors"
	"fmt"
	"sort"

	"context"

	"gorm.io/gorm"

	"dpmptsp/api/internal/domain"
)

type siteSettingsModel struct {
	ID   int64  `gorm:"column:id;primaryKey"`
	Nama string `gorm:"column:nama"`
	Logo string `gorm:"column:logo"`
}

func (siteSettingsModel) TableName() string { return "website_pengaturan" }

type menuModel struct {
	ID           int64  `gorm:"column:id;primaryKey"`
	ParentID     *int64 `gorm:"column:parent_id"`
	Nama         string `gorm:"column:nama"`
	URL          string `gorm:"column:url"`
	Urutan       int    `gorm:"column:urutan"`
	IsActive     bool   `gorm:"column:is_active"`
	Tipe         string `gorm:"column:tipe"`
	TombolKontak bool   `gorm:"column:tombol_kontak"`
}

func (menuModel) TableName() string { return "header_menu" }

type SiteRepo struct{ db *gorm.DB }

func NewSiteRepo(db *gorm.DB) *SiteRepo { return &SiteRepo{db: db} }

var _ domain.SiteRepository = (*SiteRepo)(nil)

// Settings falls back rather than failing. A missing settings row should not
// take the whole site down — every page renders the header.
func (r *SiteRepo) Settings(ctx context.Context) (domain.SiteSettings, error) {
	var m siteSettingsModel
	err := r.db.WithContext(ctx).Where("id = ?", 1).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.SiteSettings{Name: "DPM-PTSP Surabaya"}, nil
	}
	if err != nil {
		return domain.SiteSettings{}, fmt.Errorf("site settings: %w", err)
	}
	return domain.SiteSettings{Name: m.Nama, Logo: m.Logo}, nil
}

// MenuTree returns the navigation already nested and sorted.
func (r *SiteRepo) MenuTree(ctx context.Context) ([]domain.MenuNode, []domain.MenuNode, error) {
	var rows []menuModel
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&rows).Error; err != nil {
		return nil, nil, fmt.Errorf("header menu: %w", err)
	}

	byID := make(map[int64]*domain.MenuNode, len(rows))
	parentOf := make(map[int64]*int64, len(rows))
	order := make(map[int64]int, len(rows))

	for _, m := range rows {
		byID[m.ID] = &domain.MenuNode{
			ID: m.ID, Name: m.Nama, URL: m.URL, Order: m.Urutan,
			Type: m.Tipe, ContactButton: m.TombolKontak,
		}
		// The legacy data uses both NULL and 0 to mean "no parent".
		if m.ParentID != nil && *m.ParentID != 0 {
			parentOf[m.ID] = m.ParentID
		}
		order[m.ID] = m.Urutan
	}

	var roots []*domain.MenuNode
	for id, node := range byID {
		if pid, ok := parentOf[id]; ok {
			if parent, exists := byID[*pid]; exists {
				parent.Children = append(parent.Children, *node)
				continue
			}
			// Parent is missing or inactive: treat as a root rather than
			// dropping the item, so a bad row cannot hide navigation.
		}
		roots = append(roots, node)
	}

	sortNodes(roots, order)

	var nav, contact []domain.MenuNode
	for _, n := range roots {
		sortChildren(n)
		if n.ContactButton {
			contact = append(contact, *n)
		} else {
			nav = append(nav, *n)
		}
	}
	if nav == nil {
		nav = []domain.MenuNode{}
	}
	if contact == nil {
		contact = []domain.MenuNode{}
	}
	return nav, contact, nil
}

func sortNodes(nodes []*domain.MenuNode, _ map[int64]int) {
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].Order != nodes[j].Order {
			return nodes[i].Order < nodes[j].Order
		}
		return nodes[i].ID < nodes[j].ID
	})
}

func sortChildren(n *domain.MenuNode) {
	sort.SliceStable(n.Children, func(i, j int) bool {
		if n.Children[i].Order != n.Children[j].Order {
			return n.Children[i].Order < n.Children[j].Order
		}
		return n.Children[i].ID < n.Children[j].ID
	})
	for i := range n.Children {
		sortChildren(&n.Children[i])
	}
}
