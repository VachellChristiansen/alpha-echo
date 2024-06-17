package dtos

type GateRequest struct {
	From     string `json:"from" form:"from"`
	Name     string `json:"name" form:"name"`
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
	Remember string `json:"remember" form:"remember"`
}

type GateSwitchRequest struct {
	To string `json:"to" form:"to"`
}
