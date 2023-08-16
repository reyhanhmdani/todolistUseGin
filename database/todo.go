package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
	"todoGin/model/entity"
)

// adaptop pattern
type TodoRepository struct {
	DB       *gorm.DB
	S3Bucket *s3.Client
}

func NewTodoRepository(DB *gorm.DB, s3Bucket *s3.Client) *TodoRepository {
	return &TodoRepository{
		DB:       DB,
		S3Bucket: s3Bucket,
	}
}

func (t TodoRepository) GetAll() ([]entity.Todolist, error) {
	var todos []entity.Todolist

	result := t.DB.Preload("Attachments").Preload("User").Find(&todos)
	return todos, result.Error
}

func (t TodoRepository) GetAllUserByID(UserID int64) ([]entity.Todolist, error) {
	var todos []entity.Todolist

	// Ambil semua Todolist berdasarkan user_id
	result := t.DB.Preload("Attachments").Where("user_id = ?", UserID).Find(&todos)
	if result.Error != nil {
		return nil, result.Error
	}

	return todos, nil
}

func (t TodoRepository) GetByID(todoID, userID int64) (*entity.Todolist, error) {
	var todo entity.Todolist
	result := t.DB.Preload("Attachments").Where("id = ? AND user_id = ?", todoID, userID).First(&todo)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &todo, result.Error
}

func (t TodoRepository) Create(title string, userID int64) (*entity.Todolist, error) {
	todo := entity.Todolist{
		Title:  title,
		UserID: userID,
	}
	result := t.DB.Create(&todo)
	return &todo, result.Error
}

func (t TodoRepository) Update(todoID, userID int64, updates map[string]interface{}) (*entity.Todolist, error) {
	var todo entity.Todolist
	result := t.DB.Model(&todo).Where("id = ? AND user_id = ?", todoID, userID).Updates(updates)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &todo, result.Error
}

func (t TodoRepository) UpdatetoAtch(todo *entity.Todolist) error {
	err := t.DB.Save(todo).Error
	return err
}

func (t TodoRepository) Delete(todoID, userID int64) (int64, error) {
	todo := entity.Todolist{}

	// Fetch the Todolist by ID and user_id
	if err := t.DB.Where("id = ? AND user_id = ?", todoID, userID).First(&todo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If Todolist not found, return 0 RowsAffected
			return 0, nil
		}
		return 0, err
	}

	// Delete the fetched Todolist
	result := t.DB.Delete(&todo)
	return result.RowsAffected, result.Error
}

func (t TodoRepository) CreateUser(user *entity.User) error {
	if err := t.DB.Create(user).Error; err != nil {
		return err
	}
	return nil
}

func (t TodoRepository) GetUserByUsername(username string) (*entity.User, error) {
	var user entity.User
	if err := t.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil

}

/////////////////////////////////////////

func (t *TodoRepository) CreateAttachment(todoID int64, path string, order int64) (*entity.Attachment, error) {
	attachment := &entity.Attachment{
		TodoID:          todoID,
		Path:            path,
		AttachmentOrder: order,
	}
	if err := t.DB.Create(attachment).Error; err != nil {
		return nil, err
	}
	return attachment, nil
}

func (t *TodoRepository) UploadTodoFileS3Atch(file *multipart.FileHeader, todoID, userID int64) (*entity.Attachment, error) {
	//Mengambil Todolist berdasarkan ID dan user_id

	todolist := &entity.Todolist{}
	if err := t.DB.Where("id = ? AND user_id = ?", todoID, userID).First(todolist).Error; err != nil {
		return nil, err
	}

	src, err := file.Open()
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	defer src.Close()

	// bikin nama file yang uniq untuk menghindari konflik
	uniqueFilename := fmt.Sprintf("%s%s", uuid.NewString(), filepath.Ext(file.Filename))

	// Upload the file to S3
	bucketName := "bucketwithrey"
	objectKey := uniqueFilename
	_, err = t.S3Bucket.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   src,
		//ACL:    types.ObjectCannedACLPublicRead, // Optional: Mengatur ACL agar file yang diunggah dapat diakses oleh publik
	})
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	// Return the public URL of the uploaded file
	publicURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, objectKey)

	// Create an attachment record in the database
	var attachmentOrder int64 = 1 // Set the initial attachment_order
	// Get the count of existing attachments for the Todolist
	existingAttachmentCount := int64(0)
	t.DB.Model(&entity.Attachment{}).Where("todo_id = ?", todoID).Count(&existingAttachmentCount)
	attachmentOrder = existingAttachmentCount + 1 // Set attachment_order dynamically

	// Create an attachment record in the database
	attachment := &entity.Attachment{
		TodoID:          todoID,
		Path:            publicURL,
		AttachmentOrder: attachmentOrder, // atur order
		Timestamp:       time.Now(),
	}
	err = t.DB.Create(attachment).Error
	if err != nil {
		return nil, err
	}

	return attachment, nil
}
func (t *TodoRepository) UpdateTodoWithAttachments(todo *entity.Todolist) error {
	return t.DB.Transaction(func(tx *gorm.DB) error {
		// Pertama, hapus semua lampiran yang ada yang terkait dengan Todo
		if err := t.DB.Where("todo_id = ?", todo.ID).Delete(&entity.Attachment{}).Error; err != nil {
			return err
		}

		// Next, create new attachment records
		for i := range todo.Attachments {
			attachment := &todo.Attachments[i]
			if err := t.DB.Create(attachment).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (t *TodoRepository) UploadFileS3Buckets(file io.Reader, fileName string) (*string, error) {
	bucketName := "bucketwithrey"
	objectKey := fileName

	_, err := t.S3Bucket.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	// Return the public URL of the uploaded file
	publicURL := aws.String(fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, objectKey))

	return publicURL, nil
}

func (t *TodoRepository) UploadTodoFileLocalAtch(file *multipart.FileHeader, todoID, userID int64) (*entity.Attachment, error) {
	// Fetch the Todolist by ID and user_id
	todolist := &entity.Todolist{}
	if err := t.DB.Where("id = ? AND user_id = ?", todoID, userID).First(todolist).Error; err != nil {
		return nil, err
	}

	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// bikin nama file yang uniq untuk menghindari konflik
	uniqueFilename := fmt.Sprintf("%s%s", uuid.NewString(), filepath.Ext(file.Filename))

	// Upload the file to Local
	uploadDir := "uploads"

	// Buat direktori unggahan jika belum ada
	err = os.MkdirAll(uploadDir, 0755)
	if err != nil {
		return nil, err
	}

	// Create the destination file
	dest, err := os.Create(filepath.Join(uploadDir, uniqueFilename))
	if err != nil {
		return nil, err
	}
	defer dest.Close()

	// Copy file nya ke file tujuan
	_, err = io.Copy(dest, src)
	if err != nil {
		return nil, err
	}
	// Return the local file path
	localFilePath := filepath.Join(uploadDir, uniqueFilename)

	// Create an attachment record in the database
	var attachmentOrder int64 = 1 // Set the initial attachment_order
	// Get the count of existing attachments for the Todolist
	existingAttachmentCount := int64(0)
	t.DB.Model(&entity.Attachment{}).Where("todo_id = ?", todoID).Count(&existingAttachmentCount)
	attachmentOrder = existingAttachmentCount + 1 // Set attachment_order dynamically

	// Create an attachment record in the database
	attachment := &entity.Attachment{
		TodoID:          todoID,
		Path:            localFilePath,
		AttachmentOrder: attachmentOrder, // atur order
		Timestamp:       time.Now(),
	}
	err = t.DB.Create(attachment).Error
	if err != nil {
		return nil, err
	}

	return attachment, nil
}

func (t *TodoRepository) SearchTodolistByUser(userID int64, search string, page, perPage int) ([]entity.Todolist, int64, error) {
	var todos []entity.Todolist

	// Menghitung total data
	var total int64
	t.DB.Model(&entity.Todolist{}).Where("user_id = ? AND title LIKE ?", userID, "%"+search+"%").Count(&total)

	// Mengambil data dengan paginasi
	offset := (page - 1) * perPage
	err := t.DB.Where("user_id = ? AND title LIKE ?", userID, "%"+search+"%").
		Offset(offset).Limit(perPage).
		Preload("Attachments").Find(&todos).Error

	return todos, total, err
}

////////////////////////////////////////////////////////
