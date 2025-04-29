package model

import (
	"time"
)

type User struct {
	Model `gorm:"embedded"` // embed id and time

	Username string    `gorm:"column:username;type:varchar(100);not null" json:"username"`
	Nickname string    `gorm:"column:nickname;type:varchar(100);not null" json:"nickname"`
	Password string    `gorm:"column:password;type:varchar(100);not null" json:"password"`
	LoginAt  time.Time `gorm:"column:login_at;type:datetime" json:"loginAt"`
	LoginIP  string    `gorm:"column:login_ip;type:varchar(100)" json:"loginIP"`
}

// TableName table name
func (m *User) TableName() string {
	return "user"
}
