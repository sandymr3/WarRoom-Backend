package models

import "time"

// Batch represents a cohort/group of participants who share a batch code.
// Admins create batches; participants register with the batch code.
type Batch struct {
	ID        string     `gorm:"primaryKey;type:varchar(191)" json:"id"`
	Code      string     `gorm:"column:code;uniqueIndex;not null;type:varchar(50)" json:"code"`
	Name      string     `gorm:"column:name" json:"name"`                      // e.g. "March 2026 Cohort A"
	Level     int        `gorm:"column:level;not null;default:1" json:"level"` // 1=Student, 2=Manager
	AdminID   string     `gorm:"column:admin_id;not null;type:varchar(191)" json:"adminId"`
	Active    bool       `gorm:"column:active;not null;default:true" json:"active"`
	StartsAt  *time.Time `gorm:"column:starts_at" json:"startsAt"`
	EndsAt    *time.Time `gorm:"column:ends_at" json:"endsAt"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updatedAt"`
}
