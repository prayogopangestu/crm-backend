package entities

import "time"

type Task struct {
	ID             string     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string     `json:"-" gorm:"type:uuid;not null;column:organization_id"`
	AssigneeID     string     `json:"assigneeId,omitempty" gorm:"type:uuid;column:assignee_id"`
	Title          string     `json:"title" gorm:"type:text;not null"`
	Company        string     `json:"company" gorm:"type:text;not null"`
	Time           string     `json:"time" gorm:"-"`
	Date           string     `json:"date" gorm:"-"`
	DueDate        time.Time  `json:"-" gorm:"type:date;not null"`
	DueTime        string     `json:"-" gorm:"type:time;not null"`
	Type           string     `json:"type" gorm:"type:text;not null"`
	Status         string     `json:"status" gorm:"-"`
	Completed      bool       `json:"completed" gorm:"type:boolean;not null;default:false"`
	CompletedAt    *time.Time `json:"-" gorm:"type:timestamptz"`
	Notes          string     `json:"notes" gorm:"type:text;not null;default:''"`
	Priority       string     `json:"priority" gorm:"type:text;not null"`
	Assignee       string     `json:"assignee" gorm:"-"`
	CreatedAt      time.Time  `json:"createdAt,omitempty" gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt      time.Time  `json:"updatedAt,omitempty" gorm:"type:timestamptz;not null;default:now()"`
	DeletedAt      *time.Time `json:"-" gorm:"type:timestamptz"`
}

func (Task) TableName() string { return "tasks" }
