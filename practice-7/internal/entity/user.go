package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	Username        string    `gorm:"uniqueIndex"`
	Email           string    `gorm:"uniqueIndex"`
	Password        string
	Role            string    `gorm:"default:user"`
	Verified        bool      `gorm:"default:false"`
	VerifyCode      string
	VerifyExpiresAt time.Time
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}