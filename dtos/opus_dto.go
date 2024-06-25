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
