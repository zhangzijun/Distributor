package Background

import (
	"Distributor/model"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

func RegisterTaskRoutes(r *gin.Engine, db *sql.DB) {
	r.GET("/tasks", func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "10")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 10
		}

		offset := (page - 1) * limit

		tasks, err := model.GetTasksLimit(db, offset, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.HTML(http.StatusOK, "tasks_list.html", gin.H{
			"tasks": tasks,
			"page":  page,
			"limit": limit,
		})
	})

	r.GET("/tasks/edit/:id", func(c *gin.Context) {
		idStr := c.Param("id")         // 从 URL 获取 ID 参数，类型为 string
		id, err := strconv.Atoi(idStr) // 将字符串转换为整数
		if err != nil {
			c.HTML(http.StatusBadRequest, "tasks_edit.html", gin.H{"error": "Invalid task ID"})
			return
		}
		tasks, err := model.GetTasks(db)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "tasks_edit.html", gin.H{"error": err.Error()})
			return
		}
		var task model.Task
		for _, t := range tasks {
			if t.ID == id {
				task = t
				break
			}
		}
		c.HTML(http.StatusOK, "tasks_edit.html", gin.H{"task": task})
	})

	r.POST("/tasks/edit/:id", func(c *gin.Context) {
		idStr := c.Param("id")         // 从 URL 获取 ID 参数，类型为 string
		id, err := strconv.Atoi(idStr) // 将字符串转换为整数
		if err != nil {
			c.HTML(http.StatusBadRequest, "tasks_edit.html", gin.H{"error": "Invalid task ID"})
			return
		}
		var task model.Task
		if err := c.ShouldBind(&task); err != nil {
			c.HTML(http.StatusBadRequest, "tasks_edit.html", gin.H{"error": err.Error()})
			return
		}

		task.ID = id
		task.TaskUpdateTime = time.Now()

		if err := model.UpdateTask(db, task); err != nil {
			fmt.Print(`%#v`, err)
			c.HTML(http.StatusInternalServerError, "tasks_edit.html", gin.H{"error": "Failed to update task"})
			return
		}

		c.Redirect(http.StatusSeeOther, "/tasks")
	})
}
