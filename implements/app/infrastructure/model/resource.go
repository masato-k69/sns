package model

type ResourceSearchIndex struct {
	ResourceID string `json:"resource_id"`
	Type       string `json:"type"`
	Keyword    string `json:"keyword"`
}

func (r ResourceSearchIndex) Index() string {
	return "resource_search_indexes"
}
