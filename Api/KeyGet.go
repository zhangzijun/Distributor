package Api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type Order struct {
	ID             int
	MobileBrand    string
	MacAddr        string
	RemainingKeys  int
	SuccessfulKeys int
}

type Task struct {
	ID         int
	TaskNumber string
	TaskStatus int
	UID        string
	UIDKey     string
}

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./Data/order_management.db")
	if err != nil {
		panic(err)
	}
}

func encrypt(key, text string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(text))

	return hex.EncodeToString(ciphertext), nil
}

func decrypt(key, cipherHex string) (string, error) {
	ciphertext, _ := hex.DecodeString(cipherHex)

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

func getUIDKey(c *gin.Context) {
	mobileBrand := c.Query("mobile_brand")
	macAddr := c.Query("mac_addr")
	uid := c.Query("uid")

	// Step 1: 查找 orders 表
	var order Order
	err := db.QueryRow("SELECT id, mobile_brand, mac_addr, remaining_keys FROM orders WHERE mobile_brand = ? AND mac_addr = ?", mobileBrand, macAddr).Scan(&order.ID, &order.MobileBrand, &order.MacAddr, &order.RemainingKeys)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding order"})
		return
	}

	if order.RemainingKeys <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough remaining keys"})
		return
	}

	// Step 2: 查找 tasks 表
	var task Task
	month := time.Now().Format("200601")
	tableName := "tasks_" + month

	query := fmt.Sprintf("SELECT id, task_number FROM %s WHERE task_number LIKE 'Order_%d_%%' AND task_status = 0 ORDER BY id LIMIT 1", tableName, order.ID)
	err = db.QueryRow(query).Scan(&task.ID, &task.TaskNumber)
	if err != nil {
		// 尝试从上个月的表中查找
		lastMonth := time.Now().AddDate(0, -1, 0).Format("200601")
		lastTableName := "tasks_" + lastMonth
		query = fmt.Sprintf("SELECT id, task_number FROM %s WHERE task_number LIKE 'Order_%d_%%' AND task_status = 0 ORDER BY id LIMIT 1", lastTableName, order.ID)
		err = db.QueryRow(query).Scan(&task.ID, &task.TaskNumber)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No available tasks"})
			return
		}
	}

	// Step 3: 生成 UIDKEY
	uidKey, err := encrypt(uid, macAddr+"-_-"+task.TaskNumber)
	if err != nil {
		fmt.Print("%#v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating UIDKEY"})
		return
	}

	// Step 4: 更新数据库
	_, err = db.Exec("UPDATE orders SET remaining_keys = remaining_keys - 1 WHERE id = ?", order.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating order"})
		return
	}

	updateQuery := fmt.Sprintf("UPDATE %s SET task_status = 1, uid = ?, uid_key = ? WHERE id = ?", tableName)
	_, err = db.Exec(updateQuery, uid, uidKey, task.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating task"})
		return
	}

	// Step 5: 返回 UIDKEY
	c.JSON(http.StatusOK, gin.H{
		"uidkey":      uidKey,
		"remain_keys": order.RemainingKeys - 1,
	})
}
