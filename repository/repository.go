package repository

import (
	"io"
	"mime/multipart"
	"todoGin/model/entity"
)

type TodoRepository interface {
	GetAll() ([]entity.Todolist, error)
	GetAllUserByID(UserID int64) ([]entity.Todolist, error)
	GetByID(todoID, userID int64) (*entity.Todolist, error)
	Create(title string, userID int64) (*entity.Todolist, error)
	Update(todoID, userID int64, updates map[string]interface{}) (*entity.Todolist, error)
	UpdatetoAtch(todo *entity.Todolist) error
	Delete(todoID, userID int64) (int64, error)
	CreateUser(user *entity.User) error
	GetUserByUsername(username string) (*entity.User, error)
	//UploadTodoFileS3(file *multipart.FileHeader, url string) error
	//UploadTodoFileLocal(file *multipart.FileHeader, url string) error
	/////////////////////
	CreateAttachment(todoID int64, path string, order int64) (*entity.Attachment, error)
	UploadTodoFileS3Atch(file *multipart.FileHeader, todoID, userID int64) (*entity.Attachment, error)
	UpdateTodoWithAttachments(todo *entity.Todolist) error
	UploadFileS3Buckets(file io.Reader, fileName string) (*string, error)
	UploadTodoFileLocalAtch(file *multipart.FileHeader, todoID, userID int64) (*entity.Attachment, error)
	SearchTodolistByUser(userID int64, search string, page, perPage int) ([]entity.Todolist, int64, error)
}
