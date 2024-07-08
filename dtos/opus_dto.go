package dtos

type AddCategoryRequest struct {
	Name     string `json:"name" form:"name"`
	Priority int    `json:"prioirty" form:"priority"`
}

type AddTaskRequest struct {
	Title      string `json:"title" form:"title"`
	Priority   int    `json:"prioirty" form:"priority"`
	CategoryID uint   `json:"category_id" form:"category_id"`
	ParentID   uint   `json:"parent_id" form:"parent_id"`
	ParentType string `json:"parent_type" form:"parent_type"`
}

type UpdateTaskRequest struct {
	Id        int    `json:"id" form:"id"`
	Updating  string `json:"updating" form:"updating"`
	Details   string `json:"details" form:"details"`
	Notes     string `json:"notes" form:"notes"`
	StartDate string `json:"start_date" form:"start_date"`
	EndDate   string `json:"end_date" form:"end_date"`
}

type UpdateOpusStateRequest struct {
	Id      int    `json:"id" form:"id"`
	Section string `json:"section" form:"section"`
	State   string `json:"state" form:"state"`
}
