package database

import (
	"time"

	"dpmptsp/api/internal/domain"
)

// GORM models. These are the only place that knows the physical column names,
// which is what lets the domain types be shaped for the application rather than
// for the legacy schema.

type articleModel struct {
	ID           int64     `gorm:"column:id_post;primaryKey"`
	CategoryID   int64     `gorm:"column:id_category"`
	Slug         string    `gorm:"column:slug"`
	Title        string    `gorm:"column:title"`
	Content      string    `gorm:"column:content"`
	RefContent   string    `gorm:"column:ref_content"`
	SEOTitle     string    `gorm:"column:seotitle"`
	Tag          string    `gorm:"column:tag"`
	PublishedAt  time.Time `gorm:"column:published_at"`
	Editor       string    `gorm:"column:editor"`
	IsActive     bool      `gorm:"column:is_active"`
	IsHeadline   bool      `gorm:"column:is_headline"`
	Picture      string    `gorm:"column:picture"`
	Hits         int64     `gorm:"column:hits"`
	TitleEN      *string   `gorm:"column:title_en"`
	ContentEN    *string   `gorm:"column:content_en"`
	RefContentEN *string   `gorm:"column:ref_content_en"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`

	Category *categoryModel `gorm:"foreignKey:CategoryID;references:ID"`
}

func (articleModel) TableName() string { return "post" }

type categoryModel struct {
	ID       int64  `gorm:"column:id_category;primaryKey"`
	Title    string `gorm:"column:title"`
	IsActive bool   `gorm:"column:is_active"`
}

func (categoryModel) TableName() string { return "category_berita" }

func (m articleModel) toDomain() domain.Article {
	a := domain.Article{
		ID: m.ID, CategoryID: m.CategoryID, Slug: m.Slug, Title: m.Title,
		Content: m.Content, RefContent: m.RefContent, SEOTitle: m.SEOTitle,
		Tag: m.Tag, PublishedAt: m.PublishedAt, Editor: m.Editor,
		IsActive: m.IsActive, IsHeadline: m.IsHeadline, Picture: m.Picture,
		Hits: m.Hits, TitleEN: m.TitleEN, ContentEN: m.ContentEN,
		RefContentEN: m.RefContentEN,
	}
	if m.Category != nil {
		c := m.Category.toDomain()
		a.Category = &c
	}
	return a
}

func (m categoryModel) toDomain() domain.Category {
	return domain.Category{ID: m.ID, Title: m.Title, IsActive: m.IsActive}
}

func articleFromDomain(a *domain.Article) articleModel {
	return articleModel{
		ID: a.ID, CategoryID: a.CategoryID, Slug: a.Slug, Title: a.Title,
		Content: a.Content, RefContent: a.RefContent, SEOTitle: a.SEOTitle,
		Tag: a.Tag, PublishedAt: a.PublishedAt, Editor: a.Editor,
		IsActive: a.IsActive, IsHeadline: a.IsHeadline, Picture: a.Picture,
		Hits: a.Hits, TitleEN: a.TitleEN, ContentEN: a.ContentEN,
		RefContentEN: a.RefContentEN,
	}
}
