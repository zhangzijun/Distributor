package Api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

func keyUsed(c *gin.Context) {
	mobileBrand := c.Query("mobile_brand")
	macAddr := c.Query("mac_addr")
	uid := c.Query("uid")
	uidKey := c.Query("uidkey")

	// Step 1: 解密 UIDKEY 得到 task_number
	plainText, err := decrypt(uid, uidKey)
	if err != nil {
		print(`%#v`, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decrypting UIDKEY"})
		return
	}

	parts := strings.Split(plainText, "-_-")
	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UIDKEY"})
		return
	}

	taskNumber := parts[1]

	// Step 2: 查找 tasks 表
	var task Task
	month := time.Now().Format("200601")
	tableName := "tasks_" + month

	query := fmt.Sprintf("SELECT id FROM %s WHERE task_number = ?", tableName)
	err = db.QueryRow(query, taskNumber).Scan(&task.ID)
	if err != nil {
		// 尝试从上个月的表中查找
		lastMonth := time.Now().AddDate(0, -1, 0).Format("200601")
		lastTableName := "tasks_" + lastMonth
		query = fmt.Sprintf("SELECT id FROM %s WHERE task_number = ?", lastTableName)
		err = db.QueryRow(query, taskNumber).Scan(&task.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No task found"})
			return
		}
	}

	// Step 3: 更新 task 状态
	updateQuery := fmt.Sprintf("UPDATE %s SET task_status = 2 WHERE id = ?", tableName)
	_, err = db.Exec(updateQuery, task.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating task"})
		return
	}

	// Step 4: 更新 orders 表
	_, err = db.Exec("UPDATE orders SET successful_keys = successful_keys + 1 WHERE mobile_brand = ? AND mac_addr = ?", mobileBrand, macAddr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task completed successfully"})
}
