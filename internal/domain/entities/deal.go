package entities

import "time"

type DealAssignee struct {
	ID        string `json:"id,omitempty" gorm:"-"`
	Name      string `json:"name" gorm:"-"`
	AvatarURL string `json:"avatarUrl" gorm:"-"`
}

type Deal struct {
	ID             string     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string     `json:"-" gorm:"type:uuid;not null;column:organization_id"`
	AssigneeID     string     `json:"assigneeId,omitempty" gorm:"type:uuid;column:assignee_id"`
	Title          string     `json:"title" gorm:"type:text;not null"`
	Company        string     `json:"company" gorm:"type:text;not null"`
	Value          int64      `json:"value" gorm:"type:bigint;not null"`
	Priority       string     `json:"priority" gorm:"type:text;not null"`
	Stage          string     `json:"stage" gorm:"type:text;not null;column:stage_key"`
	LostReason     string     `json:"lostReason,omitempty" gorm:"type:text;not null;default:''"`
	Assignee       DealAssignee `json:"assignee" gorm:"-"`
	CreatedAt      time.Time  `json:"createdAt,omitempty" gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt      time.Time  `json:"updatedAt,omitempty" gorm:"type:timestamptz;not null;default:now()"`
	DeletedAt      *time.Time `json:"-" gorm:"type:timestamptz"`
}

func (Deal) TableName() string { return "deals" }
