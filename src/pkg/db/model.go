package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uuid.UUID `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("ID", uuid.New().String())
	return nil
}
