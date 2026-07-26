package request

import (
	dealusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/deal"
)

// DealCreate represents the JSON body for create/update deal endpoints.
type DealCreate struct {
	Title      string `json:"title"`
	Company    string `json:"company"`
	Value      int64  `json:"value"`
	Priority   string `json:"priority"`
	Stage      string `json:"stage"`
	AssigneeID string `json:"assigneeId"`
	LostReason string `json:"lostReason"`
}

// ToInput converts the request DTO into the deal usecase Input.
func (r DealCreate) ToInput() dealusecase.Input {
	return dealusecase.Input{
		Title: r.Title, Company: r.Company, Value: r.Value,
		Priority: r.Priority, Stage: r.Stage, AssigneeID: r.AssigneeID, LostReason: r.LostReason,
	}
}

// DealStageUpdate represents the JSON body for the deal stage update endpoint.
type DealStageUpdate struct {
	Stage      string `json:"stage"`
	LostReason string `json:"lostReason"`
}

// ToInput converts the request DTO into the deal usecase StageInput.
func (r DealStageUpdate) ToInput() dealusecase.StageInput {
	return dealusecase.StageInput{Stage: r.Stage, LostReason: r.LostReason}
}
