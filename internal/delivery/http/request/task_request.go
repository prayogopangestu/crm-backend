package request

import (
	taskusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/task"
)

// TaskCreate represents the JSON body for create/update task endpoints.
type TaskCreate struct {
	Title      string `json:"title"`
	Company    string `json:"company"`
	Time       string `json:"time"`
	Date       string `json:"date"`
	Type       string `json:"type"`
	Priority   string `json:"priority"`
	Assignee   string `json:"assignee"`
	AssigneeID string `json:"assigneeId"`
	Notes      string `json:"notes"`
	Completed  bool   `json:"completed"`
}

// ToInput converts the request DTO into the task usecase Input.
func (r TaskCreate) ToInput() taskusecase.Input {
	return taskusecase.Input{
		Title: r.Title, Company: r.Company, Time: r.Time, Date: r.Date,
		Type: r.Type, Priority: r.Priority, Assignee: r.Assignee, AssigneeID: r.AssigneeID,
		Notes: r.Notes, Completed: r.Completed,
	}
}

// TaskToggle represents the JSON body for the task toggle endpoint.
type TaskToggle struct {
	Completed bool `json:"completed"`
}
