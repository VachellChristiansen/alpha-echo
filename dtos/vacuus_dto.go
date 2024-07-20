package dtos

type UpdateVacuusAnimationRequest struct {
	Name     string `json:"name" form:"name"`
	Category string `json:"category" form:"category"`
}
