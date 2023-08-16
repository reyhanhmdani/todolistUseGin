package request

type LoginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
	UserID  int    `json:"user_id"`
	//Todolist []entity.Todolist `json:"todolist"`
}
