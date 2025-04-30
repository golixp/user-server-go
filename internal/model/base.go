package model

import (
	"time"

	"user-server-go/pkg/ids"

	"gorm.io/gorm"
)

type Model struct {
	ID        uint64         `gorm:"column:id;primaryKey" json:"id,string"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// 自动创建主键
func (m *Model) BeforeCreate(tx *gorm.DB) error {
	if m.ID == 0 {
		m.ID = ids.GenerateID()
	}
	return nil
}
