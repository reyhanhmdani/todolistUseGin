package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"todoGin/mocks"
	"todoGin/model/entity"
	"todoGin/model/request"
	"todoGin/model/respErr"
)

func TestGetAll1(t *testing.T) {

	testCases := []struct {
		name               string
		expectedStatusCode int
		mockTodo           []entity.Todolist
		mockErr            error
		expectedResponse   request.TodoResponseToGetAll
	}{
		{
			name:               "Success",
			expectedStatusCode: http.StatusOK,
			mockTodo: []entity.Todolist{
				{ID: 1, Title: "Task 1", Status: false},
				{ID: 2, Title: "Task 2", Status: false},
			},
			mockErr: nil,
			expectedResponse: request.TodoResponseToGetAll{
				Message: "Success Get All",
				Data:    2,
				Todos: []entity.Todolist{
					{ID: 1, Title: "Task 1", Status: false},
					{ID: 2, Title: "Task 2", Status: false},
				},
			},
		},
		{
			name:               "Internal Server Error",
			expectedStatusCode: http.StatusInternalServerError,
			mockTodo:           []entity.Todolist{},
			mockErr:            errors.New("Internal Server Error"),
			expectedResponse: request.TodoResponseToGetAll{
				Message: "Internal Server Error",
				Data:    0,
				Todos:   []entity.Todolist(nil),
			},
		},
		{
			name:               "Empty",
			expectedStatusCode: http.StatusOK,
			mockTodo:           []entity.Todolist{},
			mockErr:            nil,
			expectedResponse: request.TodoResponseToGetAll{
				Message: "Success Get All",
				Data:    0,
				Todos:   []entity.Todolist{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewTodoRepository(t)
			repo.On("GetAll").Return(tc.mockTodo, tc.mockErr)

			handler := NewTodoService(repo)

			r, err := http.NewRequest("GET", "/manage-todos", nil)
			if err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			router := gin.Default()
			router.GET("/manage-todos", handler.TodolistHandlerGetAll)
			router.ServeHTTP(w, r)

			assert.Equal(t, tc.expectedStatusCode, w.Code)

			var resp request.TodoResponseToGetAll
			err = json.Unmarshal(w.Body.Bytes(), &resp)
			if err != nil {
				t.Fatal(err)
			}

			fmt.Printf("TYPE: %T\n", tc.expectedResponse.Todos)
			fmt.Printf("TYPE: %T\n", resp.Todos)
			/////////////////////////////////////////////////////

			assert.Equal(t, tc.expectedResponse.Message, resp.Message)
			//assert.IsEqual(t, reflect.DeepEqual(tc.expectedResponse.Todos, resp.Todos))
			assert.Equal(t, tc.expectedResponse.Todos, resp.Todos)
			assert.Equal(t, tc.expectedResponse.Data, resp.Data)
			//assert.Equal(t, tc.expectedResponse.Todos, resp.Todos)
		})
	}
}

func TestCreate1(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		mock           func(todoRepository *mocks.TodoRepository)
		expectedStatus int
		expectedData   entity.Todolist
		expectedError  string
	}{
		{
			name: "Success",
			body: `{"title": "Makan"}`,
			mock: func(mock *mocks.TodoRepository) {
				newTodo := &entity.Todolist{
					Title:  "Makan",
					Status: false,
				}
				mock.On("Create", "Makan").Return(newTodo, nil)
			},
			expectedStatus: http.StatusOK,
			expectedData: entity.Todolist{
				Title:  "Makan",
				Status: false,
			},
			expectedError: "",
		},
		{
			name:           "Invalid input",
			body:           `{"title": ""}`,
			mock:           func(mock *mocks.TodoRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedData:   entity.Todolist{},
			expectedError:  "Invalid input",
		},
		{
			name: "Internal Server Error",
			body: `{"title": "Test Todo"}`,
			mock: func(mock *mocks.TodoRepository) {
				expectedError := errors.New("Internal Server Error")
				mock.On("Create", "Test Todo").Return(nil, expectedError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedData:   entity.Todolist{},
			expectedError:  "Internal Server Error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			todoRepo := mocks.NewTodoRepository(t)
			tc.mock(todoRepo)
			handler := NewTodoService(todoRepo)

			endpoint := "/manage-todo"

			body := bytes.NewBufferString(tc.body)
			req, err := http.NewRequest(http.MethodPost, endpoint, body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)
			r.POST(endpoint, handler.TodolistHandlerCreate)

			c.Request = req
			r.ServeHTTP(w, req)

			respBody, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			if tc.expectedError != "" {
				var errResp respErr.ErrorResponse
				err = json.Unmarshal(respBody, &errResp)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedError, errResp.Message)
			} else {
				var result request.TodoResponse
				err = json.Unmarshal(respBody, &result)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedData, result.Data)
			}

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestTodolistHandlerDelete(t *testing.T) {
	mockRepo := mocks.NewTodoRepository(t)
	handler := NewTodoService(mockRepo)
	gin.SetMode(gin.TestMode)

	// Test cases
	tests := []struct {
		name      string
		todoID    int64
		isFound   int64
		repoError error
		expStatus int
		expResp   interface{}
	}{
		{
			name:      "Success",
			todoID:    1,
			isFound:   1,
			expStatus: http.StatusOK,
			repoError: nil,
			expResp: request.TodoDeleteResponse{
				Status:  http.StatusOK,
				Message: "Success Delete",
			},
		},
		{
			name:      "Not Found",
			todoID:    2,
			isFound:   0,
			repoError: nil,
			expStatus: http.StatusNotFound,
			expResp: request.TodoDeleteResponse{
				Message: "Not Found",
				Status:  http.StatusNotFound,
			},
		},
		{
			name:      "Internal Server Error",
			todoID:    3,
			isFound:   0,
			repoError: errors.New("Internal Server Error"),
			expStatus: http.StatusInternalServerError,
			expResp: request.TodoDeleteResponse{
				Message: "Internal Server Error",
				Status:  http.StatusInternalServerError,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo.On("Delete", tc.todoID).Return(tc.isFound, tc.repoError)

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodDelete, "/manage-todo/todo/"+strconv.FormatInt(tc.todoID, 10), nil)
			router := gin.Default()
			router.DELETE("/manage-todo/todo/:id", handler.TodolistHandlerDelete)
			router.ServeHTTP(w, r)

			assert.Equal(t, tc.expStatus, w.Code)

			var respBody request.TodoDeleteResponse
			err := json.Unmarshal(w.Body.Bytes(), &respBody)
			if err != nil {
				log.Print(err)
			}
			require.NoError(t, err)
			//assert.IsEqual(t, reflect.DeepEqual(tc.expResp, &respBody))
			assert.Equal(t, tc.expResp, respBody)

		})
	}
}

func TestGetByID1(t *testing.T) {
	testCases := []struct {
		name            string
		inputID         int64
		expectedStatus  int
		expectedMessage string
		expectedData    entity.Todolist
		mockError       error
		mockResult      *entity.Todolist
	}{
		{
			name:            "Success",
			inputID:         1,
			expectedStatus:  http.StatusOK,
			expectedMessage: "Success Get Id",
			expectedData:    entity.Todolist{ID: 1, Title: "Test Todo"},
			mockError:       nil,
			mockResult:      &entity.Todolist{ID: 1, Title: "Test Todo"},
		},
		{
			name:            "Not Found",
			inputID:         2,
			expectedStatus:  http.StatusNotFound,
			expectedMessage: "Not Found",
			expectedData:    entity.Todolist{},
			mockError:       nil,
			mockResult:      nil,
		},
		{
			name:            "Internal Server Error",
			inputID:         3,
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "Internal Server Error",
			expectedData:    entity.Todolist{},
			mockError:       errors.New("Internal Server Error"),
			mockResult:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockTodoRepo := mocks.NewTodoRepository(t)
			handler := NewTodoService(mockTodoRepo)

			mockTodoRepo.On("GetByID", tc.inputID).Return(tc.mockResult, tc.mockError)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", fmt.Sprintf("/manage-todo/todo/%d", tc.inputID), nil)
			router := gin.Default()
			router.GET("/manage-todo/todo/:id", handler.TodolistHandlerGetByID)
			router.ServeHTTP(w, req)

			respBody, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var response request.TodoResponse
			err = json.Unmarshal(respBody, &response)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Equal(t, tc.expectedMessage, response.Message)
			assert.Equal(t, tc.expectedData, response.Data)
		})
	}
}

func TestUpdate1(t *testing.T) {

	mockRepo := mocks.NewTodoRepository(t)

	// membuat object handler dan menambahkan dependensi mock
	handler := NewTodoService(mockRepo)

	testCases := []struct {
		name           string
		id             int64
		requestPayload request.TodolistUpdateRequest
		mockBehavior   func()
		expectedStatus int
		expectedResp   interface{}
		expectedError  string
	}{
		{
			name: "Success",
			id:   1,
			requestPayload: request.TodolistUpdateRequest{
				Title: "New Title",
			},
			mockBehavior: func() {
				expectedTodo := entity.Todolist{
					ID:     1,
					Title:  "New Title",
					Status: false,
				}
				mockRepo.On("GetByID", int64(1)).Return(&entity.Todolist{}, nil)
				mockRepo.On("Update", int64(1), mock.Anything).Return(&expectedTodo, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp: map[string]interface{}{
				"data":   "Success Update Todo",
				"status": 200,
				"todos":  entity.Todolist{},
			},
			expectedError: "",
		},
		{
			name: "Not Found",
			id:   2,
			requestPayload: request.TodolistUpdateRequest{
				Title: "New Title",
			},
			mockBehavior: func() {
				mockRepo.On("GetByID", int64(2)).Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedResp: respErr.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "ID not Found",
			},
			expectedError: "ID not Found",
		},
		{
			name: "Internal Server Error",
			id:   3,
			requestPayload: request.TodolistUpdateRequest{
				Title:  "New Title",
				Status: false,
			},
			mockBehavior: func() {
				mockRepo.On("GetByID", int64(3)).Return(&entity.Todolist{}, nil)
				mockRepo.On("Update", int64(3), mock.Anything).Return(nil, errors.New("Internal Server Error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp: respErr.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Internal Server Error",
			},
			expectedError: "Internal Server Error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			requestBodyBytes, _ := json.Marshal(tc.requestPayload)

			req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/manage-todo/todo/%d", tc.id), bytes.NewBuffer(requestBodyBytes))
			w := httptest.NewRecorder()

			r := gin.Default()
			r.PUT("/manage-todo/todo/:id", handler.TodolistHandlerUpdate)
			r.ServeHTTP(w, req)

			respBody, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, w.Code)

			//var resp interface{}
			//if tc.expectedStatus == http.StatusOK {
			//	resp = request.TodoUpdateResponse{}
			//} else {
			//	resp = respErr.ErrorResponse{}
			//}

			if tc.expectedError != "" {
				var errResp respErr.ErrorResponse
				err = json.Unmarshal(respBody, &errResp)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedError, errResp.Message)
			}

			//assert.IsEqual(t, reflect.DeepEqual(tc.expectedResp, resp))
			assert.Equal(t, tc.expectedStatus, w.Code)

			mockRepo.AssertExpectations(t)
		})
	}
}
