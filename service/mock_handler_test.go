package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-playground/assert/v2"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"todoGin/mocks"
	"todoGin/model/entity"
	"todoGin/model/request"
	"todoGin/model/respErr"
)

//
// golangci-lint run --timeout=5m --fix ./...

func TestTodolist(t *testing.T) {
	t.Run("TestGetAll", TestGetAll)
	t.Run("TestCreate", TestCreate)
	t.Run("TestUpdate", TestUpdate)
	t.Run("TestGetByID", TestGetByID)
	t.Run("TestDelete", TestDelete)
}

// one for All
func TestGetAll(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		mockTodo := []entity.Todolist{
			{ID: 1, Title: "Task 1", Status: false},
			{ID: 2, Title: "Task 2", Status: false},
		}

		// success
		repo := mocks.NewTodoRepository(t)
		repo.On("GetAll").Return(mockTodo, nil)

		handler := NewTodoService(repo)

		req, err := http.NewRequest("GET", "/manage-todos", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router := gin.Default()
		router.GET("/manage-todos", handler.TodolistHandlerGetAll)
		router.ServeHTTP(rr, req)

		var resp request.TodoResponseToGetAll
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Success Get All", resp.Message)
		assert.Equal(t, len(mockTodo), resp.Data)
		assert.Equal(t, mockTodo, resp.Todos)

	})

	// Internal Server Error
	t.Run("Internal Server Error", func(t *testing.T) {
		repo := mocks.NewTodoRepository(t)
		repo.On("GetAll").Return(nil, errors.New("some error"))

		handler := NewTodoService(repo)

		req, err := http.NewRequest("GET", "/manage-todos", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router := gin.Default()
		router.GET("/manage-todos", handler.TodolistHandlerGetAll)
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var resp respErr.ErrorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "some error", resp.Message)
		assert.Equal(t, http.StatusInternalServerError, resp.Status)
	})

	t.Run("Empty", func(t *testing.T) {
		repo := mocks.NewTodoRepository(t)
		repo.On("GetAll").Return([]entity.Todolist{}, nil)

		handler := NewTodoService(repo)

		req, err := http.NewRequest("GET", "/manage-todos", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router := gin.Default()
		router.GET("/manage-todos", handler.TodolistHandlerGetAll)
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp request.TodoResponseToGetAll
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Success Get All", resp.Message)
		assert.Equal(t, 0, resp.Data)
		assert.Equal(t, []entity.Todolist{}, resp.Todos)
		//assert.IsEqual(t, resp.Todos)
		//assert.
	})
}

// Create
func TestCreate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		todoRepo := mocks.NewTodoRepository(t)

		// Set up mock behavior
		newTodo := &entity.Todolist{
			//ID:     1,
			Title:  "Makan",
			Status: false,
		}

		todoRepo.On("Create", "Makan").Return(newTodo, nil)

		// Initialize todo service with mock repository
		handler := NewTodoService(todoRepo)

		// Call the create endpoint
		endpoint := "/manage-todo"
		r := gin.New()
		r.POST(endpoint, handler.TodolistHandlerCreate)

		// Create an HTTP request to create a new Todo
		reqBody := bytes.NewBufferString(`{"title": "Makan"}`)
		req, err := http.NewRequest(http.MethodPost, endpoint, reqBody)
		require.NoError(t, err)

		// Send the request and read the response body
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		respBody, err := io.ReadAll(w.Body)
		if err != nil {
			log.Printf("Response body: %s\n", respBody)
		}
		//_ = fmt.Sprintf("%s", respBody)
		require.NoError(t, err)

		// Unmarshal the response body into a Todo object
		var result request.TodoResponse

		if err := json.Unmarshal(respBody, &result); err != nil {
			log.Printf("Failed to unmarshal JSON response body: %v", err)
		}

		// Assert that the response has a 200 ok status code and returns the new Todo
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, *newTodo, result.Data)
	})

	// invalid
	t.Run("Invalid", func(t *testing.T) {

		todorepo := mocks.NewTodoRepository(t)
		handler := NewTodoService(todorepo)

		expectedErrors := errors.New("Invalid input")

		//todorepo.On("Create", "").Return(nil, expectedErrors)

		endpoint := "/manage-todo"

		// Create invalid input
		body := bytes.NewBufferString(`{"title": ""}`)
		req, err := http.NewRequest(http.MethodPost, endpoint, body)
		require.NoError(t, err)

		// Set up Gin context
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.POST(endpoint, func(context *gin.Context) {
			handler.TodolistHandlerCreate(context)
		})

		// Perform request
		c.Request = req
		r.ServeHTTP(w, req)

		respBody, err := io.ReadAll(w.Body)
		require.NoError(t, err)

		var errResp respErr.ErrorResponse
		err = json.Unmarshal(respBody, &errResp)
		require.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, expectedErrors.Error(), errResp.Message)

		// Check mock call
		todorepo.AssertNotCalled(t, "Called", mock.Anything)

	})

	// internal Server Error
	t.Run("Internal Server Error", func(t *testing.T) {

		todoRepo := mocks.NewTodoRepository(t)

		handler := NewTodoService(todoRepo)

		expectedError := errors.New("Internal Server Error")
		endpoint := "/manage-todo"

		todoRepo.On("Create", "Test Todo").Return(nil, expectedError)

		// Create valid input
		body := bytes.NewBufferString(`{"title": "Test Todo"}`)
		req, err := http.NewRequest("POST", endpoint, body)
		require.NoError(t, err)

		// Set up Gin context
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.POST(endpoint, handler.TodolistHandlerCreate)

		// Perform request
		c.Request = req
		r.ServeHTTP(w, req)

		respBody, err := io.ReadAll(w.Body)
		require.NoError(t, err)

		var errResp respErr.ErrorResponse
		err = json.Unmarshal(respBody, &errResp)
		require.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, expectedError.Error(), errResp.Message)

		// Check mock call
		todoRepo.AssertCalled(t, "Create", "Test Todo")
	})

}

// Update
func TestUpdate(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		// membuat object mock
		mockRepo := mocks.NewTodoRepository(t)

		// membuat object handler dan menambahkan dependensi mock
		handler := NewTodoService(mockRepo)

		// create request body
		reqBody := request.TodolistUpdateRequest{
			Title: "New Title",
		}
		requestBodyBytes, _ := json.Marshal(reqBody)

		// create expected response
		expectedTodo := entity.Todolist{
			ID:     1,
			Title:  "New Title",
			Status: false,
		}
		mockRepo.On("GetByID", int64(1)).Return(&entity.Todolist{}, nil)
		mockRepo.On("Update", int64(1), mock.Anything).Return(&expectedTodo, nil)

		// create test request
		req, _ := http.NewRequest(http.MethodPut, "/manage-todo/todo/1", bytes.NewBuffer(requestBodyBytes))
		rr := httptest.NewRecorder()

		// perform test request
		r := gin.Default()
		r.PUT("/manage-todo/todo/:id", handler.TodolistHandlerUpdate)
		r.ServeHTTP(rr, req)

		// check response
		assert.Equal(t, http.StatusOK, rr.Code)
		var resp request.TodoUpdateResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Status)
		assert.Equal(t, "Success Update Todo", resp.Message)

	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo := mocks.NewTodoRepository(t)

		// membuat object handler dan menambahkan dependensi mock
		handler := NewTodoService(mockRepo)

		reqBody1 := request.TodolistUpdateRequest{
			Title: "New Title",
		}
		requestBodyBytes, _ := json.Marshal(reqBody1)

		// create mock behavior
		mockRepo.On("GetByID", int64(2)).Return(nil, nil)

		// create test request
		req, _ := http.NewRequest(http.MethodPut, "/manage-todo/todo/2", bytes.NewBuffer(requestBodyBytes))
		rr := httptest.NewRecorder()

		// perform test request
		r := gin.Default()
		r.PUT("/manage-todo/todo/:id", handler.TodolistHandlerUpdate)
		r.ServeHTTP(rr, req)

		// check response
		assert.Equal(t, http.StatusNotFound, rr.Code)
		var resp1 respErr.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp1)
		require.NoError(t, err)
		assert.Equal(t, "ID not Found", resp1.Message)
		assert.Equal(t, http.StatusNotFound, resp1.Status)

		// assert mock behavior
		mockRepo.AssertExpectations(t)
	})
	// internal Server Error

	t.Run("Internal Server Error", func(t *testing.T) {
		mockRepo := mocks.NewTodoRepository(t)

		// membuat object handler dan menambahkan dependensi mock
		handler := NewTodoService(mockRepo)

		mockRepo.On("GetByID", int64(3)).Return(&entity.Todolist{}, nil)
		mockRepo.On("Update", int64(3), mock.Anything).Return(nil, errors.New("Internal Server Error"))

		// membuat handler dengan mock object

		// membuat request payload
		payload := request.TodolistUpdateRequest{
			Title:  "New Title",
			Status: false,
		}
		requestBody, _ := json.Marshal(payload)

		// membuat request http
		req, _ := http.NewRequest("PUT", "/manage-todo/todo/3", bytes.NewBuffer(requestBody))
		rr := httptest.NewRecorder()

		// perform test request
		r := gin.Default()
		r.PUT("/manage-todo/todo/:id", handler.TodolistHandlerUpdate)
		r.ServeHTTP(rr, req)

		// melakukan pengecekan status code dan response
		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var resp2 respErr.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp2)
		require.NoError(t, err)
		assert.Equal(t, "Internal Server Error", resp2.Message)
		assert.Equal(t, http.StatusInternalServerError, resp2.Status)

		// melakukan pengecekan apakah ekspektasi sudah terpanggil
		mockRepo.AssertExpectations(t)
	})

}

// Get By id
func TestGetByID(t *testing.T) {
	// inisiasi mocking
	t.Run("Success", func(t *testing.T) {
		mockTodoRepo := mocks.NewTodoRepository(t)

		// inisiasi handler
		handler := NewTodoService(mockTodoRepo)

		// testing success
		mockTodoRepo.On("GetByID", int64(1)).Return(&entity.Todolist{ID: 1, Title: "Test Todo"}, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/manage-todo/todo/1", nil)
		router := gin.Default()
		router.GET("/manage-todo/todo/:id", handler.TodolistHandlerGetByID)
		router.ServeHTTP(w, req)

		respBody, err := io.ReadAll(w.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, w.Code)

		var response request.TodoResponse
		err = json.Unmarshal(respBody, &response)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, response.Message, "Success Get Id")
	})

	// testing not found
	t.Run("Not Found", func(t *testing.T) {
		mockTodoRepo := mocks.NewTodoRepository(t)

		// inisiasi handler
		handler := NewTodoService(mockTodoRepo)

		mockTodoRepo.On("GetByID", int64(2)).Return(nil, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/manage-todo/todo/2", nil)
		router := gin.Default()
		router.GET("/manage-todo/todo/:id", handler.TodolistHandlerGetByID)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var res respErr.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.Equal(t, "Not Found", res.Message)
		assert.Equal(t, http.StatusNotFound, res.Status)
	})

	// testing internal server error
	t.Run("Internal Server Error", func(t *testing.T) {
		mockTodoRepo := mocks.NewTodoRepository(t)

		// inisiasi handler
		handler := NewTodoService(mockTodoRepo)

		mockTodoRepo.On("GetByID", int64(3)).Return(nil, errors.New("Internal Server Error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/manage-todo/todo/3", nil)
		router := gin.Default()
		router.GET("/manage-todo/todo/:id", handler.TodolistHandlerGetByID)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var internal respErr.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &internal)
		require.NoError(t, err)
		assert.Equal(t, "Internal Server Error", internal.Message)
		assert.Equal(t, http.StatusInternalServerError, internal.Status)
	})

}

// Delete
func TestDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockTodoRepo := mocks.NewTodoRepository(t)
		handler := NewTodoService(mockTodoRepo)

		// Testing Success
		mockTodoRepo.On("Delete", int64(1)).Return(int64(1), nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/manage-todo/todo/1", nil)
		router := gin.Default()
		router.DELETE("/manage-todo/todo/:id", handler.TodolistHandlerDelete)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp request.TodoDeleteResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Success Delete", resp.Message)

	})
	//Testing Not Found

	t.Run("Not Found", func(t *testing.T) {
		mockTodoRepo := mocks.NewTodoRepository(t)
		handler := NewTodoService(mockTodoRepo)

		mockTodoRepo.On("Delete", int64(2)).Return(int64(0), nil)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/manage-todo/todo/2", nil)
		router := gin.Default()
		router.DELETE("/manage-todo/todo/:id", handler.TodolistHandlerDelete)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var res respErr.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, res.Status)
		assert.Equal(t, "Not Found", res.Message)

	})
	// internal Server ERror
	t.Run("Internal Server Error", func(t *testing.T) {
		mockTodoRepo := mocks.NewTodoRepository(t)
		handler := NewTodoService(mockTodoRepo)

		mockTodoRepo.On("Delete", int64(3)).Return(int64(0), errors.New("Internal Server Error"))
		w := httptest.NewRecorder()

		req, _ := http.NewRequest(http.MethodDelete, "/manage-todo/todo/3", nil)
		router := gin.Default()
		router.DELETE("/manage-todo/todo/:id", handler.TodolistHandlerDelete)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var res1 respErr.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &res1)
		require.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, res1.Status)
		assert.Equal(t, "Internal Server Error", res1.Message)

		mockTodoRepo.AssertExpectations(t)
	})

}
