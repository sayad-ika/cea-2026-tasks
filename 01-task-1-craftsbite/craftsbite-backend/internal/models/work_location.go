package models

import (
	"time"

	"github.com/google/uuid"
)

// WorkLocationStatus represents a user's explicit work location setting for a date
type WorkLocationStatus struct {
	ID        uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID    `gorm:"type:uuid;not null;uniqueIndex:unique_work_location_user_date;index:idx_work_location_status_user_date" json:"user_id" validate:"required"`
	Date      string       `gorm:"type:date;not null;uniqueIndex:unique_work_location_user_date;index:idx_work_location_status_user_date;index:idx_work_location_status_date" json:"date" validate:"required"`
	Location  WorkLocation `gorm:"type:varchar(20);not null" json:"location" validate:"required"`
	UpdatedBy *uuid.UUID   `gorm:"type:uuid" json:"updated_by,omitempty"`
	Reason    *string      `gorm:"type:varchar(255)" json:"reason,omitempty"`
	CreatedAt time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time    `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User    User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Updater *User `gorm:"foreignKey:UpdatedBy;constraint:OnDelete:SET NULL" json:"updater,omitempty"`
}

// TableName specifies the table name for GORM
func (WorkLocationStatus) TableName() string {
	return "work_location_statuses"
}

// GlobalWorkLocationPolicy represents a global location policy for a date range
type GlobalWorkLocationPolicy struct {
	ID         uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	StartDate  string       `gorm:"type:date;not null;index:idx_global_work_location_policy_active_range" json:"start_date" validate:"required"`
	EndDate    string       `gorm:"type:date;not null;index:idx_global_work_location_policy_active_range" json:"end_date" validate:"required"`
	Location   WorkLocation `gorm:"type:varchar(20);not null" json:"location" validate:"required"`
	IsActive   bool         `gorm:"not null;default:true;index:idx_global_work_location_policy_active_range" json:"is_active"`
	Reason     *string      `gorm:"type:varchar(255)" json:"reason,omitempty"`
	DeclaredBy uuid.UUID    `gorm:"type:uuid;not null" json:"declared_by" validate:"required"`
	CreatedAt  time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time    `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Declarer User `gorm:"foreignKey:DeclaredBy;constraint:OnDelete:CASCADE" json:"declarer,omitempty"`
}

// TableName specifies the table name for GORM
func (GlobalWorkLocationPolicy) TableName() string {
	return "global_work_location_policies"
}

// WorkLocationHistory represents the audit trail for work location changes
type WorkLocationHistory struct {
	ID               uuid.UUID                 `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID                 `gorm:"type:uuid;not null;index:idx_work_location_history_user_date" json:"user_id" validate:"required"`
	Date             string                    `gorm:"type:date;not null;index:idx_work_location_history_user_date" json:"date" validate:"required"`
	PreviousLocation *WorkLocation             `gorm:"type:varchar(20)" json:"previous_location,omitempty"`
	NewLocation      WorkLocation              `gorm:"type:varchar(20);not null" json:"new_location" validate:"required"`
	ChangedByUserID  *uuid.UUID                `gorm:"type:uuid" json:"changed_by_user_id,omitempty"`
	Reason           *string                   `gorm:"type:varchar(255)" json:"reason,omitempty"`
	Action           WorkLocationHistoryAction `gorm:"type:varchar(30);not null" json:"action" validate:"required"`
	CreatedAt        time.Time                 `gorm:"autoCreateTime;index:idx_work_location_history_created_at" json:"created_at"`

	// Relationships
	User      User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	ChangedBy *User `gorm:"foreignKey:ChangedByUserID;constraint:OnDelete:SET NULL" json:"changed_by,omitempty"`
}

// TableName specifies the table name for GORM
func (WorkLocationHistory) TableName() string {
	return "work_location_history"
}
