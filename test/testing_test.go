package test

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"todoGin/database"
	"todoGin/router"
	"todoGin/service"
)

func setupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open("root:Pastibisa@tcp(localhost:3306)/Gin_test"))
	if err != nil {
		fmt.Println(err)
	}

	logrus.Info("Connect to Database")
	return db, nil
}

func setupRouter(db *gorm.DB) *gin.Engine {
	todoRepo := database.NewTodoRepository(db)
	todoService := service.NewTodoService(todoRepo)
	routeBuilder := router.NewRouteBuilder(todoService)
	routeInit := routeBuilder.RouteInit()

	return routeInit
}

func truncateTodolist(DB *gorm.DB) {
	DB.Exec("TRUNCATE todolists")
}

func TestCreateSuccess(t *testing.T) {
	db, _ := setupTestDB()

	truncateTodolist(db)
	router := setupRouter(db)

	requestBody := strings.NewReader(`{"title": "sholat isya"}`)
	request := httptest.NewRequest(http.MethodPost, "http://localhost:3000/manage-todo", requestBody)
	request.Header.Add("Authorization", "Basic a2V5OnZhbHVl")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	response := recorder.Result()
	assert.Equal(t, 200, response.StatusCode)
	body, _ := io.ReadAll(response.Body)
	var responseBody map[string]interface{}
	err := json.Unmarshal(body, &responseBody)
	if err != nil {
		logrus.Error(err)
	}
	fmt.Println(responseBody)

	//assert.Nil(t, err)
	assert.Equal(t, 200, int(responseBody["status"].(float64)))
	assert.Equal(t, "sholat isya", responseBody["data"].(map[string]interface{})["title"])
}
func TestCreateFailedValidation(t *testing.T) {
	db, _ := setupTestDB()
	truncateTodolist(db)
	router := setupRouter(db)

	requestBody := strings.NewReader(`{"title" : ""}`)
	request := httptest.NewRequest(http.MethodPost, "http://localhost:3000/manage-todo", requestBody)
	request.Header.Add("Authorization", "Basic a2V5OnZhbHVl")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	response := recorder.Result()
	assert.Equal(t, 400, response.StatusCode)
	body, _ := io.ReadAll(response.Body)
	var responseBody map[string]interface{}
	err := json.Unmarshal(body, &responseBody)
	if err != nil {
		logrus.Error(err)
	}
	fmt.Println(responseBody)

	assert.Equal(t, 400, int(responseBody["status"].(float64)))
}
func TestUpdateSuccess(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		log.Fatal(err)
	}
	truncateTodolist(db)
	router := setupRouter(db)

	tx := db.Begin()
	todolistRepository := database.NewTodoRepository(db)
	todolist, _ := todolistRepository.Create("halo")

	tx.Commit()

	requestBody := strings.NewReader(`{"title": "sholat isya","status": true}`)
	request := httptest.NewRequest(http.MethodPut, "http://localhost:3000/manage-todo/todo/"+strconv.Itoa(int(todolist.ID)), requestBody)
	request.Header.Add("Authorization", "Basic a2V5OnZhbHVl")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	response := recorder.Result()
	assert.Equal(t, 200, response.StatusCode)

	body, _ := io.ReadAll(response.Body)
	var responseBody map[string]interface{}
	_ = json.Unmarshal(body, &responseBody)
	fmt.Println(responseBody)

	assert.Equal(t, 200, int(responseBody["status"].(float64)))
	assert.Equal(t, "sholat isya", responseBody["todos"].(map[string]interface{})["title"])

}
func TestUpdateInvalid(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		log.Fatal(err)
	}
	truncateTodolist(db)
	router := setupRouter(db)

	tx := db.Begin()
	todolistRepository := database.NewTodoRepository(db)
	todolist, _ := todolistRepository.Create("holaa")

	tx.Commit()

	requestBody := strings.NewReader(`{"title":  }`)
	request := httptest.NewRequest(http.MethodPut, "http://localhost:3000/manage-todo/todo/"+strconv.Itoa(int(todolist.ID)), requestBody)
	request.Header.Add("Authorization", "Basic a2V5OnZhbHVl")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	response := recorder.Result()
	assert.Equal(t, 400, response.StatusCode)

	body, _ := io.ReadAll(response.Body)
	var responseBody map[string]interface{}
	_ = json.Unmarshal(body, &responseBody)
	fmt.Println(responseBody)

	assert.Equal(t, 400, int(responseBody["status"].(float64)))
}
func TestGetSuccess(t *testing.T) {
	db, err := setupTestDB()
	truncateTodolist(db)
	router := setupRouter(db)

	tx := db.Begin()
	todolistRepository := database.NewTodoRepository(db)
	todolist, _ := todolistRepository.Create("makan pagi")
	tx.Commit()

	request := httptest.NewRequest(http.MethodGet, "/manage-todo/todo/1", nil)
	request.Header.Add("Authorization", "Basic a2V5OnZhbHVl")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	response := recorder.Result()
	assert.Equal(t, 200, response.StatusCode)

	body, _ := io.ReadAll(response.Body)
	require.NoError(t, err)
	defer response.Body.Close()
	var responseBody map[string]interface{}
	_ = json.Unmarshal(body, &responseBody)
	fmt.Println(responseBody)

	assert.Equal(t, 200, int(responseBody["status"].(float64)))
	assert.Equal(t, "Success Get Id", responseBody["message"])
	assert.Equal(t, todolist.Title, responseBody["data"].(map[string]interface{})["title"])
}
func TestGetFailed(t *testing.T) {
	db, _ := setupTestDB()
	truncateTodolist(db)
	router := setupRouter(db)

	request := httptest.NewRequest(http.MethodGet, "/manage-todo/todo/404", nil)
	request.Header.Add("Authorization", "Basic a2V5OnZhbHVl")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	response := recorder.Result()
	assert.Equal(t, 404, response.StatusCode)

	body, _ := io.ReadAll(response.Body)
	var responseBody map[string]interface{}
	_ = json.Unmarshal(body, &responseBody)
	fmt.Println(responseBody)

	assert.Equal(t, 404, int(responseBody["status"].(float64)))
	assert.Equal(t, "Not Found", responseBody["message"])
}
func TestDeleteSuccess(t *testing.T) {
	db, _ := setupTestDB()
	truncateTodolist(db)

	tx := db.Begin()

	todolistRepo := database.NewTodoRepository(db)
	todolist, _ := todolistRepo.Create("hapus ini")

	tx.Commit()

	router := setupRouter(db)

	request := httptest.NewRequest(http.MethodDelete, "/manage-todo/todo/"+strconv.Itoa(int(todolist.ID)), nil)
	request.Header.Add("Authorization", "Basic a2V5OnZhbHVl")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	response := recorder.Result()
	assert.Equal(t, 200, response.StatusCode)

	body, _ := io.ReadAll(response.Body)
	var responseBody map[string]interface{}
	_ = json.Unmarshal(body, &responseBody)
	fmt.Println(responseBody)

	assert.Equal(t, 200, int(responseBody["status"].(float64)))
	assert.Equal(t, "Success Delete", responseBody["message"])
}
func TestDeleteFailedNotFound(t *testing.T) {
	db, _ := setupTestDB()
	truncateTodolist(db)

	router := setupRouter(db)

	request := httptest.NewRequest(http.MethodDelete, "/manage-todo/todo/404", nil)
	request.Header.Add("Authorization", "Basic a2V5OnZhbHVl")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	response := recorder.Result()
	assert.Equal(t, 404, response.StatusCode)

	body, _ := io.ReadAll(response.Body)
	var responseBody map[string]interface{}
	_ = json.Unmarshal(body, &responseBody)
	fmt.Println(responseBody)

	assert.Equal(t, 404, int(responseBody["status"].(float64)))
	assert.Equal(t, "Not Found", responseBody["message"])
}
func TestGetAll(t *testing.T) {
	db, _ := setupTestDB()
	truncateTodolist(db)

	tx := db.Begin()

	todolistRepo := database.NewTodoRepository(db)
	todolist1, _ := todolistRepo.Create("hapus ini")
	todolist2, _ := todolistRepo.Create("hapus itu")
	tx.Commit()

	router := setupRouter(db)

	request := httptest.NewRequest(http.MethodGet, "/manage-todos", nil)
	request.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6IlJleTEiLCJleHAiOjE2ODkxNzUwOTV9.WjM7D8pVt-T15OiZxClp_CZpeDHh-QFc_1FiqtePgwA")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	response := recorder.Result()
	assert.Equal(t, 200, response.StatusCode)

	body, _ := io.ReadAll(response.Body)
	var responseBody map[string]interface{}
	_ = json.Unmarshal(body, &responseBody)
	fmt.Println(responseBody)

	//assert.Equal(t, 200, int(responseBody["status"].(float64)))
	assert.Equal(t, "Success Get All", responseBody["message"])

	var Todolists = responseBody["todos"].([]interface{})

	TodolistsResponse1 := Todolists[0].(map[string]interface{})
	TodolistsResponse2 := Todolists[1].(map[string]interface{})

	assert.Equal(t, todolist1.ID, int64(TodolistsResponse1["id"].(float64)))
	assert.Equal(t, todolist1.Title, (TodolistsResponse1["title"]))

	assert.Equal(t, todolist2.ID, int64(TodolistsResponse2["id"].(float64)))
	assert.Equal(t, todolist2.Title, (TodolistsResponse2["title"]))
}
func TestUnauthorized(t *testing.T) {
	db, _ := setupTestDB()
	truncateTodolist(db)
	router := setupRouter(db)

	request := httptest.NewRequest(http.MethodGet, "/manage-todos", nil)
	request.Header.Add("Authorization", "")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	response := recorder.Result()
	assert.Equal(t, 401, response.StatusCode)

	body, _ := io.ReadAll(response.Body)
	var responseBody map[string]interface{}
	_ = json.Unmarshal(body, &responseBody)
	fmt.Println(responseBody)

	assert.Equal(t, 401, int(responseBody["status"].(float64)))
	assert.Equal(t, "UNAUTHORIZED", responseBody["message"])
}
