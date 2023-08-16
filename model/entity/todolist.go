package entity

type Todolist struct {
	ID          int64        `gorm:"primaryKey" json:"id"`
	Title       string       `gorm:"type:varchar(300)" json:"title"`
	Status      bool         `gorm:"default:false" json:"status"`
	UserID      int64        `json:"-"`
	Attachments []Attachment `gorm:"foreignKey:todo_id" json:"attachments"`
}

//func (t Todolist) Read(p []byte) (n int, err error) {
//	//TODO implement me
//	panic("implement me")
//}
//type Todolist struct {
//	ID     int64  `gorm:"primaryKey" json:"id"`
//	Title  string `gorm:"type:varchar(300)" json:"title"`
//	Status bool   `gorm:"default:false" json:"status"`
//}
