package request

import (
	pipelineusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/pipeline"
)

// PipelineCreate represents the JSON body for the create pipeline stage endpoint.
type PipelineCreate struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// ToInput converts the request DTO into the pipeline usecase Input.
func (r PipelineCreate) ToInput() pipelineusecase.Input {
	return pipelineusecase.Input{Name: r.Name, Color: r.Color}
}

// PipelineReorder represents the JSON body for the reorder pipeline stages endpoint.
type PipelineReorder struct {
	StagesOrder []string `json:"stagesOrder"`
}
