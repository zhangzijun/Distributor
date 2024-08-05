package Background

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"Distributor/model"
	"github.com/gin-gonic/gin"
)

func RegisterOrderHandlers(r *gin.Engine, db *sql.DB) {

	// HTML 渲染
	r.GET("/orders", func(c *gin.Context) {
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

		orders, err := model.GetOrdersLimit(db, offset, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.HTML(http.StatusOK, "orders_list.html", gin.H{
			"orders": orders,
			"page":   page,
			"limit":  limit,
		})
	})

	r.GET("/orders/create", func(c *gin.Context) {
		c.HTML(http.StatusOK, "orders_create.html", nil)
	})

	r.POST("/orders/create", func(c *gin.Context) {
		print("do create")
		var order model.Order
		if err := c.ShouldBind(&order); err != nil {
			c.HTML(http.StatusBadRequest, "orders_create.html", gin.H{"error": err.Error()})
			return
		}
		order.KeyGenerationTime = time.Now()
		order.KeyUpdateTime = time.Now()
		fmt.Printf("%#v\n", order)
		if err := model.AddOrder(db, &order); err != nil {
			fmt.Println("%#v", err)
			c.HTML(http.StatusInternalServerError, "orders_create.html", gin.H{"error": err.Error()})
			return
		}
		// 根据 TotalKeys 创建相应数量的 tasks
		for i := 0; i < order.TotalKeys; i++ {
			task := model.Task{
				TaskNumber:       fmt.Sprintf("Order-%d-%d", order.ID, i+1),
				MobilePlatform:   order.MobileBrand,
				UID:              fmt.Sprintf(""),
				UIDKey:           fmt.Sprintf(""),
				TaskStatus:       0,
				TaskCreationTime: time.Now(),
				TaskUpdateTime:   time.Now(),
			}

			if err := model.InsertTask(db, &task); err != nil {
				fmt.Println("%#v", err)
				c.HTML(http.StatusInternalServerError, "orders_create.html", gin.H{"error": err.Error()})
				return
			}
		}
		c.Redirect(http.StatusFound, "/orders")
	})

	r.GET("/orders/edit/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		println(id)
		if err != nil {
			c.HTML(http.StatusBadRequest, "orders_edit.html", gin.H{"error": "Invalid order ID"})
			return
		}
		order, err := model.GetOrder(db, id)
		fmt.Printf("%#v\n", order)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "orders_edit.html", gin.H{"error": err.Error()})
			return
		}
		c.HTML(http.StatusOK, "orders_edit.html", gin.H{"order": order})
	})

	r.POST("/orders/edit/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.HTML(http.StatusBadRequest, "orders_edit.html", gin.H{"error": "Invalid order ID"})
			return
		}
		var order model.Order
		if err := c.ShouldBind(&order); err != nil {
			c.HTML(http.StatusBadRequest, "orders_edit.html", gin.H{"error": err.Error()})
			return
		}
		order.ID = id
		order.KeyUpdateTime = time.Now()
		if err := model.UpdateOrder(db, &order); err != nil {
			c.HTML(http.StatusInternalServerError, "orders_edit.html", gin.H{"error": err.Error()})
			return
		}
		c.Redirect(http.StatusFound, "/orders")
	})

	r.POST("/orders/delete/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Redirect(http.StatusFound, "/orders")
			return
		}
		if err := model.DeleteOrder(db, id); err != nil {
			c.Redirect(http.StatusFound, "/orders")
			return
		}
		c.Redirect(http.StatusFound, "/orders")
	})
}
