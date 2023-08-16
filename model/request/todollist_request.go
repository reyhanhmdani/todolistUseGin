package request

type TodolistCreateRequest struct {
	Title  string `json:"title" binding:"required,min=2"`
	UserID int64  `json:"user_id"`
}

type TodolistUpdateRequest struct {
	Title  string `json:"title"`
	Status bool   `json:"status"`
}

func (r *TodolistUpdateRequest) ReqTodo() map[string]interface{} {
	updates := make(map[string]interface{})
	if r.Title != "" {
		updates["title"] = r.Title
	}
	updates["status"] = r.Status

	//if r.Status != "" {
	//	updates["status"] = r.Status
	//
	//}

	return updates
}

//type TodolistStatusRequest struct {
//	Status bool `gorm:"default:false" json:"status"`
//}
