package dtos

type GateRequest struct {
	From     string `json:"from" form:"from" validate:"required"`
	Name     string `json:"name" form:"name" validate:"required_with_all=Email Password,alpha"`
	Email    string `json:"email" form:"email" validate:"required,email"`
	Password string `json:"password" form:"password" validate:"required"`
	Remember string `json:"remember" form:"remember"`
}

type GateSwitchRequest struct {
	To string `json:"to" form:"to"`
}
