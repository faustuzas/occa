package db

import (
	"time"

	"gorm.io/gorm"

	pkgid "github.com/faustuzas/occa/src/pkg/id"
)

type BaseModel struct {
	ID        string    `gorm:"primary_key;size:36"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("ID", pkgid.NewID().String())
	return nil
}
