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

type AddTaskGoalRequest struct {
	TaskID        uint   `json:"task_id" form:"task_id"`
	GoalText      string `json:"goal_text" form:"goal_text"`
	StartDateGoal string `json:"start_date_goal" form:"start_date_goal"`
	EndDateGoal   string `json:"end_date_goal" form:"end_date_goal"`
}

type UpdateTaskRequest struct {
	ID        uint   `json:"id" form:"id"`
	Updating  string `json:"updating" form:"updating"`
	Details   string `json:"details" form:"details"`
	Notes     string `json:"notes" form:"notes"`
	StartDate string `json:"start_date" form:"start_date"`
	EndDate   string `json:"end_date" form:"end_date"`
}

type UpdateOpusStateRequest struct {
	ID      uint   `json:"id" form:"id"`
	Section string `json:"section" form:"section"`
	State   string `json:"state" form:"state"`
}
