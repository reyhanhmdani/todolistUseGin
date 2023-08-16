package request

import "todoGin/model/entity"

type SearchResponse struct {
	Status int               `json:"status"`
	Data   []entity.Todolist `json:"data"`
	Total  int64             `json:"total"`
}
