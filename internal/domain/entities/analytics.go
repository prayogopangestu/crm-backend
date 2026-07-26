package entities

import "time"

type DashboardStats struct {
	TotalLeads       int64  `json:"totalLeads" gorm:"-"`
	LeadsTrend       string `json:"leadsTrend" gorm:"-"`
	DealWonCount     int64  `json:"dealWonCount" gorm:"-"`
	WonTrend         string `json:"wonTrend" gorm:"-"`
	TotalRevenue     string `json:"totalRevenue" gorm:"-"`
	RevenueTrend     string `json:"revenueTrend" gorm:"-"`
	UrgentTasksCount int64  `json:"urgentTasksCount" gorm:"-"`
}

type ConversionPoint struct {
	Name       string  `json:"name" gorm:"-"`
	Conversion float64 `json:"Konversi" gorm:"-"`
}

type Activity struct {
	ID             string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string    `json:"-" gorm:"type:uuid;not null;column:organization_id"`
	ActorID        string    `json:"-" gorm:"type:uuid;column:actor_id"`
	ActorName      string    `json:"-" gorm:"type:text;not null"`
	User           string    `json:"user" gorm:"-"`
	Action         string    `json:"action" gorm:"type:text;not null"`
	Target         string    `json:"target" gorm:"type:text;not null"`
	Time           string    `json:"time" gorm:"-"`
	IsHighlight    bool       `json:"isHighlight" gorm:"type:boolean;not null;default:false"`
	DeletedAt      *time.Time `json:"-" gorm:"type:timestamptz"`
	CreatedAt      time.Time  `json:"createdAt,omitempty" gorm:"type:timestamptz;not null;default:now()"`
}

func (Activity) TableName() string { return "activities" }

type LeaderboardEntry struct {
	Rank       int    `json:"rank" gorm:"-"`
	Name       string `json:"name" gorm:"-"`
	Role       string `json:"role" gorm:"-"`
	Amount     int64  `json:"amount" gorm:"-"`
	Trend      string `json:"trend" gorm:"-"`
	IsPositive bool   `json:"isPositive" gorm:"-"`
	AvatarURL  string `json:"avatarUrl" gorm:"-"`
}

type LostReason struct {
	Name       string `json:"name" gorm:"-"`
	Value      int64  `json:"value" gorm:"-"`
	Percentage int64  `json:"percentage" gorm:"-"`
	Color      string `json:"color" gorm:"-"`
}

type PerformanceGoal struct {
	ID             string    `json:"-" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string    `json:"-" gorm:"type:uuid;not null;column:organization_id"`
	Month          string    `json:"month" gorm:"-"`
	MonthDate      time.Time `json:"-" gorm:"type:date;not null;column:month"`
	Goal           int64     `json:"goal" gorm:"type:bigint;not null"`
	Actual         int64     `json:"actual" gorm:"-"`
	Status         string    `json:"status" gorm:"-"`
	Percentage     int64     `json:"percentage" gorm:"-"`
	CreatedAt      time.Time `json:"-" gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt      time.Time  `json:"-" gorm:"type:timestamptz;not null;default:now()"`
	DeletedAt      *time.Time `json:"-" gorm:"type:timestamptz"`
}

func (PerformanceGoal) TableName() string { return "performance_goals" }
