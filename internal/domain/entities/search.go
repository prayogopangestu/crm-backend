package entities

type SearchResult struct {
	Contacts []Contact `json:"contacts"`
	Tasks    []Task    `json:"tasks"`
	Deals    []Deal    `json:"deals"`
}
