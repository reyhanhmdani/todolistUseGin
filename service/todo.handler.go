package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"path/filepath"
	"strconv"
	"todoGin/cfg"
	"todoGin/model/entity"
	"todoGin/model/request"
	"todoGin/model/respErr"
	"todoGin/repository"
)

type Handler struct {
	TodoRepository repository.TodoRepository
}

func NewTodoService(todoRepo repository.TodoRepository) *Handler {
	return &Handler{
		TodoRepository: todoRepo,
	}
}

func (h *Handler) Register(ctx *gin.Context) {
	var user entity.User

	// binding request body ke struct user
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.Error{
			Error: "invalid request Body",
		})
		return
	}

	// cek apakah pengguna sudah ada di database
	existingUser, err := h.TodoRepository.GetUserByUsername(user.Username)
	if existingUser != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.Error{
			Error: "User already exist",
		})
		return
	}

	// hash password pengguna sebelum disimpan ke database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, respErr.Error{
			Error: "Failed hash Password",
		})
		return
	}

	// simpan pengguna ke database
	newUser := &entity.User{
		Username: user.Username,
		Password: string(hashedPassword),
	}
	err = h.TodoRepository.CreateUser(newUser)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, respErr.Error{
			Error: "Failed Create User",
		})
		return
	}

	// mengembalikan pesan berhasil sebagai response
	ctx.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
}

func (h *Handler) Login(ctx *gin.Context) {
	var user entity.User

	// binding request body ke struct user
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.Error{
			Error: "invalid request Body",
		})
		return
	}

	// cek apakah pengguna ada di database
	storedUser, err := h.TodoRepository.GetUserByUsername(user.Username)
	if err != nil || storedUser == nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, respErr.Error{
			Error: "invalid Username or Password",
		})
		return
	}

	// bandingkan password yang dimasukkan dengan hash password di database
	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, respErr.Error{
			Error: "invalid Username or Password",
		})
		return
	}

	userID := storedUser.Id

	// membuat token
	token, err := cfg.CreateToken(user.Username, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, respErr.Error{
			Error: "Failed to generate Token",
		})
		return
	}

	_, err = h.TodoRepository.GetAllUserByID(storedUser.Id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, respErr.Error{
			Error: "Failed to get Todolist",
		})
		return
	}

	// Membuat response
	response := request.LoginResponse{
		Message: fmt.Sprintf("Hello %s! You are now logged in.", user.Username),
		Token:   token,
		UserID:  int(storedUser.Id),
		//Todolist: todolist,
	}

	// Menampilkan pesan hello user dengan username yang berhasil login
	// mengembalikan token sebagai response
	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) Access(ctx *gin.Context) {
	// ambil username dari konteks
	username, ok := ctx.Get("username")
	userID, _ := ctx.Get("user_id")
	if !ok {
		// jika tidak ada username di dalam konteks, berarti pengguna belum terautentikasi
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, respErr.Error{
			Error: "Unauthorized",
		})
		return
	}

	// kirim pesan hello ke pengguna
	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Hello %s!", username),
		"user_id": userID,
	})
}

func (h *Handler) TodolistHandlerGetAll(ctx *gin.Context) {
	// Get the user ID from the token
	userID, _ := ctx.Get("user_id")
	if userID == nil {
		logrus.Error("User not authenticated")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, respErr.ErrorResponse{
			Message: "User not authenticated",
			Status:  http.StatusUnauthorized,
		})
		return
	}

	// Cast the userID to int64
	userIDInt64, ok := userID.(int64)
	if !ok {
		logrus.Error("Invalid user_id")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Invalid user_id",
			Status:  http.StatusBadRequest,
		})
		return
	}

	todos, err := h.TodoRepository.GetAllUserByID(userIDInt64)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, &respErr.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		})
		return
	}

	logrus.Info(http.StatusOK, " Success Get All Data")
	logrus.Info(userID)
	//ctx.AbortWithStatusJSON(http.StatusOK, todos)
	ctx.AbortWithStatusJSON(http.StatusOK, request.TodoResponseToGetAll{
		Message: "Success Get All",
		UserId:  userIDInt64,
		Data:    len(todos),
		Todos:   todos,
	})

}
func (h *Handler) TodolistHandlerCreate(ctx *gin.Context) {
	todolist := new(request.TodolistCreateRequest)
	err := ctx.ShouldBindJSON(todolist)
	if err != nil {
		logrus.Error(err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Invalid input",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Get the user ID from the token
	userID, _ := ctx.Get("user_id")
	if userID == nil {
		logrus.Error("User not authenticated")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, respErr.ErrorResponse{
			Message: "User not authenticated",
			Status:  http.StatusUnauthorized,
		})
		return
	}

	// Cast the userID to int64
	userIDInt64, ok := userID.(int64)
	if !ok {
		logrus.Error("Invalid user_id")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Invalid user_id",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Set the user ID in the TodolistCreateRequest
	todolist.UserID = userIDInt64

	// Rest of your existing code...

	newTodo, errCreate := h.TodoRepository.Create(todolist.Title, todolist.UserID)
	if errCreate != nil {
		logrus.Error(errCreate)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: "Internal Server Error",
			Status:  http.StatusInternalServerError,
		})
		return
	}

	// Log the newly created Todo for debugging purposes
	logrus.Info("Newly Created Todo:", newTodo)

	logrus.Info(http.StatusOK, " Success Create Todo", todolist)
	ctx.JSON(http.StatusOK, request.TodoResponse{
		Status:  http.StatusOK,
		Message: "New Todo Created",
		Data:    *newTodo,
	})
}
func (h *Handler) TodolistHandlerGetByID(ctx *gin.Context) {

	// Get the user ID from the token
	userID, _ := ctx.Get("user_id")
	if userID == nil {
		logrus.Error("User not authenticated")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, respErr.ErrorResponse{
			Message: "User not authenticated",
			Status:  http.StatusUnauthorized,
		})
		return
	}

	// Cast the userID to int64
	userIDInt64, ok := userID.(int64)
	if !ok {
		logrus.Error("Invalid user_id")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Invalid user_id",
			Status:  http.StatusBadRequest,
		})
		return
	}

	userId := ctx.Param("id")
	todoID, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		logrus.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Bad request",
			Status:  http.StatusBadRequest,
		})
		return
	}
	todo, err := h.TodoRepository.GetByID(todoID, userIDInt64)
	if err != nil {
		logrus.Errorf("failed when get todo by id: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: "Internal Server Error",
			Status:  http.StatusInternalServerError,
		})
		return
	}
	if todo == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, respErr.ErrorResponse{
			Message: "Not Found",
			Status:  http.StatusNotFound,
		})
		return
	}
	logrus.Info(http.StatusOK, " Success Get By ID")
	ctx.JSON(http.StatusOK, request.TodoResponse{
		Status:  http.StatusOK,
		Message: "Success Get Id",
		Data:    *todo,
	})
}

func (h *Handler) TodolistHandlerUpdate(ctx *gin.Context) {
	// Get the user ID from the token
	userID, _ := ctx.Get("user_id")
	if userID == nil {
		logrus.Error("User not authenticated")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, respErr.ErrorResponse{
			Message: "User not authenticated",
			Status:  http.StatusUnauthorized,
		})
		return
	}

	// Cast the userID to int64
	userIDInt64, ok := userID.(int64)
	if !ok {
		logrus.Error("Invalid user_id")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Invalid user_id",
			Status:  http.StatusBadRequest,
		})
		return
	}

	userId := ctx.Param("id")
	todoID, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		logrus.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "parse ID error",
			Status:  http.StatusBadRequest,
		})
		return
	}
	reqBody := new(request.TodolistUpdateRequest)
	if err := ctx.ShouldBindJSON(reqBody); err != nil {
		logrus.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Bad request",
			Status:  http.StatusBadRequest,
		})
		return
	}
	ErrId, err := h.TodoRepository.GetByID(todoID, userIDInt64)
	if err != nil {
		logrus.Errorf("failed when get todo by id: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: "Internal Server Error",
			Status:  http.StatusInternalServerError,
		})
		return
	}
	if ErrId == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, respErr.ErrorResponse{
			Message: "ID not Found",
			Status:  http.StatusNotFound,
		})
		return
	}
	rowsAffected, err := h.TodoRepository.Update(todoID, userIDInt64, reqBody.ReqTodo())
	if err != nil {
		logrus.Errorf("failed when updating todo: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: "Internal Server Error",
			Status:  http.StatusInternalServerError,
		})
		return
	}
	if rowsAffected == nil {
		ctx.AbortWithStatusJSON(http.StatusOK, request.TodoIDResponse{
			Message: "Not Change",
			Data:    reqBody,
		})
		return
	}

	logrus.Info(http.StatusOK, " Success Update Todo")
	ctx.JSON(http.StatusOK, request.TodoUpdateResponse{
		Status:  http.StatusOK,
		Message: "Success Update Todo",
		Todos:   reqBody,
	})

}
func (h *Handler) TodolistHandlerDelete(ctx *gin.Context) {
	// Get the user ID from the token
	userID, _ := ctx.Get("user_id")
	if userID == nil {
		logrus.Error("User not authenticated")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, respErr.ErrorResponse{
			Message: "User not authenticated",
			Status:  http.StatusUnauthorized,
		})
		return
	}

	// Cast the userID to int64
	userIDInt64, ok := userID.(int64)
	if !ok {
		logrus.Error("Invalid user_id")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Invalid user_id",
			Status:  http.StatusBadRequest,
		})
		return
	}

	userId := ctx.Param("id")
	todoID, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		logrus.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Parse ID Error",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Delete the Todolist with the specified todoID and userID
	isDeleted, err := h.TodoRepository.Delete(todoID, userIDInt64)
	if err != nil {
		logrus.Errorf("failed when deleting todo: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: "Internal Server Error",
			Status:  http.StatusInternalServerError,
		})
		return
	}

	if isDeleted == 0 {
		ctx.AbortWithStatusJSON(http.StatusNotFound, respErr.ErrorResponse{
			Message: "Not Found",
			Status:  http.StatusNotFound,
		})
		return
	}

	logrus.Info(http.StatusOK, " Success DELETE")
	ctx.JSON(http.StatusOK, request.TodoDeleteResponse{
		Status:  http.StatusOK,
		Message: "Success Delete",
	})
}

func (h *Handler) UploadTodoFileS3AtchHandler(ctx *gin.Context) {
	// Get the user ID from the token
	userID, _ := ctx.Get("user_id")
	if userID == nil {
		logrus.Error("User not authenticated")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, respErr.ErrorResponse{
			Message: "User not authenticated",
			Status:  http.StatusUnauthorized,
		})
		return
	}

	// Cast the userID to int64
	userIDInt64, ok := userID.(int64)
	if !ok {
		logrus.Error("Invalid user_id")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Invalid user_id",
			Status:  http.StatusBadRequest,
		})
		return
	}

	todoIDStr := ctx.Param("id")
	todoID, err := strconv.ParseInt(todoIDStr, 10, 64)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Check if the Todo with the given ID exists
	todo, err := h.TodoRepository.GetByID(todoID, userIDInt64)
	if err != nil {
		ctx.JSON(http.StatusNotFound, respErr.ErrorResponse{
			Message: "Todo not found",
			Status:  http.StatusNotFound,
		})
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "No File Upload",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Check file type yang boleh cuman jpg jpeg png webp
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}
	ext := filepath.Ext(file.Filename)
	if !allowedExtensions[ext] {
		ctx.JSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "error File not allowed type",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Use the TodoRepository to upload the file to S3
	attachment, err := h.TodoRepository.UploadTodoFileS3Atch(file, todoID, userIDInt64)
	if err != nil {
		// Periksa apakah error merupakan "Todolist not found" atau bukan
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Jika error disebabkan oleh record not found, kirim respons 404
			ctx.JSON(http.StatusNotFound, respErr.ErrorResponse{
				Message: "Todolist not found",
				Status:  http.StatusNotFound,
			})
		} else {
			// Jika error bukan karena record not found, kirim respons 500
			ctx.JSON(http.StatusInternalServerError, respErr.ErrorResponse{
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			})
			logrus.Error(err)
		}
		return
	}

	// Update the Todo's Attachments field with the new attachment
	todo.Attachments = append(todo.Attachments, *attachment)

	// Create an attachment record in the database
	err = h.TodoRepository.UpdateTodoWithAttachments(todo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: "Failed to update Todo with attachments",
			Status:  http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, request.SuccessMessage{
		Message: "File uploaded and attachment created successfully",
		Data:    attachment,
		Status:  http.StatusOK,
	})
}

func (h *Handler) UploadFileS3BucketsHandler(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "No File Upload",
		})
		return
	}

	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to open file",
		})
		return
	}
	defer src.Close()

	// Use the TodoRepository to upload the file to S3
	publicURL, err := h.TodoRepository.UploadFileS3Buckets(src, file.Filename)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upload file to S3",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "File uploaded to S3 successfully",
		"url":     *publicURL,
	})

}

func (h *Handler) UploadTodoLocalAtchHandler(ctx *gin.Context) {

	// Get the userID from the token
	userID, _ := ctx.Get("user_id")
	if userID == nil {
		logrus.Error("User not authenticated")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, respErr.ErrorResponse{
			Message: "User not authenticated",
			Status:  http.StatusUnauthorized,
		})
		return
	}

	// Cast the userID to int64
	userIDInt64, ok := userID.(int64)
	if !ok {
		logrus.Error("Invalid user_id")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Invalid user_id",
			Status:  http.StatusBadRequest,
		})
		return
	}

	todoIDStr := ctx.Param("id")
	todoID, err := strconv.ParseInt(todoIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Invalid Todo ID",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Check if the Todo with given ID exists
	todo, err := h.TodoRepository.GetByID(todoID, userIDInt64)
	if err != nil {
		ctx.JSON(http.StatusNotFound, respErr.ErrorResponse{
			Message: "Todo not found",
			Status:  http.StatusBadRequest,
		})
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "No FIle Upload",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Check file
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}
	ext := filepath.Ext(file.Filename)
	if !allowedExtensions[ext] {
		ctx.JSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "error not allowed type",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// use
	attachment, err := h.TodoRepository.UploadTodoFileLocalAtch(file, todoID, userIDInt64)
	if err != nil {
		// Periksa apakah error merupakan "Todolist not found" atau bukan
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Jika error disebabkan oleh record not found, kirim respons 404
			ctx.JSON(http.StatusNotFound, respErr.ErrorResponse{
				Message: "Todolist not found",
				Status:  http.StatusNotFound,
			})
		} else {
			// Jika error bukan karena record not found, kirim respons 500
			ctx.JSON(http.StatusInternalServerError, respErr.ErrorResponse{
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			})
			logrus.Error(err)
		}
		return
	}

	// Update the Todo's Attachments field with the new attachment
	todo.Attachments = append(todo.Attachments, *attachment)

	// Save the updated Todo to the database
	err = h.TodoRepository.UpdatetoAtch(todo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Todo"})
		return
	}

	ctx.JSON(http.StatusOK, request.SuccessMessage{
		Status:  http.StatusOK,
		Message: "FIle Uploaded and attachment created successfully",
		Data:    attachment,
	})
}

func (h *Handler) TodolistsSearchHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	if userID == nil {
		logrus.Error("User not authenticated")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, respErr.ErrorResponse{
			Message: "User not authenticated",
			Status:  http.StatusUnauthorized,
		})
		return
	}

	// Cast userID ke int64
	userIDInt64, ok := userID.(int64)
	if !ok {
		logrus.Error("Invalid user_id")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Invalid user_id",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Dapatkan parameter search dari query string
	search := ctx.Query("search")

	// Dapatkan parameter page dan per_page dari query string
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "10"))

	todolists, total, err := h.TodoRepository.SearchTodolistByUser(userIDInt64, search, page, perPage)
	if err != nil {
		logrus.Errorf("failed when searching todos: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: "Internal Server Error",
			Status:  http.StatusInternalServerError,
		})
		return
	}

	// Membuat respons dengan data hasil pencarian
	response := request.SearchResponse{
		Status: http.StatusOK,
		Data:   todolists,
		Total:  total,
	}

	ctx.JSON(http.StatusOK, response)
}

///////////////////////////////////////////////////////////////////
