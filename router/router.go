package router

import (
	"github.com/gin-gonic/gin"
	"todoGin/middleware"
	todoservice "todoGin/service"
)

type RouteBuilder struct {
	todoService *todoservice.Handler
}

func NewRouteBuilder(todoService *todoservice.Handler) *RouteBuilder {
	return &RouteBuilder{todoService: todoService}
}

func (rb *RouteBuilder) RouteInit() *gin.Engine {

	r := gin.New()
	r.Use(middleware.RecoveryMiddleware(), middleware.Logger())
	//r.Use(gin.Recovery(), middleware.Logger(), middleware.BasicAuth())

	auth := r.Group("/", middleware.Authmiddleware())
	{
		auth.GET("/manage-todos", rb.todoService.TodolistHandlerGetAll)
		auth.GET("/access", rb.todoService.Access)
		auth.POST("/manage-todo", rb.todoService.TodolistHandlerCreate)
		auth.GET("/manage-todo/todo/:id", rb.todoService.TodolistHandlerGetByID)
		auth.PUT("/manage-todo/todo/:id", rb.todoService.TodolistHandlerUpdate)
		auth.DELETE("/manage-todo/todo/:id", rb.todoService.TodolistHandlerDelete)
		//auth.POST("/manage-todo/uploadS3", rb.todoService.TodoHandlerUploadFileS3)
		//auth.POST("/manage-todo/uploadLocal", rb.todoService.TodoHandlerUploadFileLocal)
		auth.POST("/uploadS3/:id", rb.todoService.UploadTodoFileS3AtchHandler)
		auth.POST("/uploadLocal/:id", rb.todoService.UploadTodoLocalAtchHandler)
		auth.GET("/list-Search", rb.todoService.TodolistsSearchHandler)
	}

	r.POST("/uploadBuckets", rb.todoService.UploadFileS3BucketsHandler)
	r.POST("/register", rb.todoService.Register)
	r.POST("/login", rb.todoService.Login)
	return r
}
