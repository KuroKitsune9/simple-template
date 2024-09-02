package routes

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"

	"todos/controller"
	"todos/db"
	mddw "todos/middleware"
)

func Init() error {
	e := echo.New()

	db, err := db.Init()
	if err != nil {
		return err
	}
	defer db.Close()

	e.GET("", func(ctx echo.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Application is Running",
		})
	})

	task := e.Group("/task")
	task.Use(mddw.ValidateToken) // Function untuk

	category := e.Group("/category")
	category.Use(mddw.ValidateToken)

	task.POST("", controller.AddTaskController(db))
	task.GET("", controller.GetAllTaskController(db))
	task.GET("/:id", controller.GetTaskById(db))
	task.DELETE("/:id", controller.DeleteTaskControll(db))
	task.DELETE("", controller.BulkDeleteTask(db))
	task.PUT("/:id", controller.EditTaskController(db))
	task.POST("/search", controller.SearchTask(db))
	task.GET("/count", controller.CountStatus(db))

	category.GET("", controller.GetAllCategory(db))
	category.GET("/:id", controller.GETcategoryById(db))
	category.POST("", controller.AddCategoryController(db))
	category.PUT("/:id", controller.UpdateCategoryController(db))
	category.DELETE("/:id", controller.DeleteCategorycontroller(db))

	e.POST("/register", controller.RegisterController(db))
	e.POST("/login", controller.LoginController(db))
	e.POST("/logout", controller.LogoutController(db))
	e.Static("/uploads", "/uploads")
	return e.Start(fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")))
}
