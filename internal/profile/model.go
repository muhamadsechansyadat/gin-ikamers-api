package profile

import "time"

type Profile struct {
	ID        int64      `gorm:"primaryKey;column:id"`
	UUID      string     `gorm:"column:uuid;type:uuid;default:gen_random_uuid()"`
	UserID    int64      `gorm:"column:user_id"`
	FullName  string     `gorm:"column:full_name"`
	Gender    *string    `gorm:"column:gender"`
	BirthDate *time.Time `gorm:"column:birth_date"`
	AvatarURL *string    `gorm:"column:avatar_url"`
	Bio       *string    `gorm:"column:bio"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (Profile) TableName() string {
	return "profiles"
}
