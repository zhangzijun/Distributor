package model

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

func InitTable() {
	var err error
	db, err = sql.Open("sqlite3", "./Data/order_management.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)
	// Initialize the orders table
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS orders (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        mac_addr TEXT NOT NULL,
        mobile_brand TEXT NOT NULL,
        key_generation_time DATETIME NOT NULL,
        total_keys INTEGER NOT NULL DEFAULT 0,
        remaining_keys INTEGER NOT NULL DEFAULT 0,
        successful_keys INTEGER NOT NULL DEFAULT 0,
        key_update_time DATETIME NOT NULL
    )`)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize a sample orders table for the current month
	month := time.Now().Format("200601")

	tableName := "tasks_" + month
	createTableSQL := `
CREATE TABLE IF NOT EXISTS ` + tableName + ` (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_number TEXT NOT NULL,
    mobile_platform TEXT NOT NULL,
    uid TEXT NOT NULL,
    uid_key TEXT NOT NULL,
    remarks TEXT DEFAULT '',
    task_status TEXT NOT NULL,
    task_creation_time DATETIME NOT NULL,
    task_update_time DATETIME NOT NULL
);`

	createIndexSQL := `
CREATE UNIQUE INDEX IF NOT EXISTS uidx_task_number
ON ` + tableName + ` (
    task_number
);`

	// 执行创建表的SQL
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// 执行创建索引的SQL
	_, err = db.Exec(createIndexSQL)
	if err != nil {
		log.Fatal(err)
	}
}

func CreateOrderTableForMonth(month string) {

}

func GetData() (bool, string) {
	// 获取一个秘钥并创建订单
	var secretKeyID int
	var remainingKeys int

	tx, err := db.Begin()
	if err != nil {
		return false, err.Error()
	}
	defer tx.Rollback()

	err = tx.QueryRow("SELECT id, remaining_keys FROM orders WHERE remaining_keys > 0 LIMIT 1").Scan(&secretKeyID, &remainingKeys)
	if err != nil {
		return false, "No available keys"
	}

	_, err = tx.Exec("UPDATE orders SET remaining_keys = remaining_keys - 1, key_update_time = ? WHERE id = ?", time.Now(), secretKeyID)
	if err != nil {
		return false, err.Error()
	}

	month := time.Now().Format("200601")
	CreateOrderTableForMonth(month)

	_, err = tx.Exec("INSERT INTO tasks_"+month+" (task_number, mobile_platform, uid, uid_key, task_status, order_creation_time, order_update_time) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"ORDER12345", "Android", "UID123", "UIDKEY123", "burning", time.Now(), time.Now())
	if err != nil {
		return false, err.Error()
	}

	err = tx.Commit()
	if err != nil {
		return false, err.Error()
	}

	return true, "ok"
}

func OrderCallBack() (bool, string) {
	var orderStatus struct {
		OrderNumber string `json:"task_number" binding:"required"`
		Status      string `json:"status" binding:"required"`
		Reason      string `json:"reason"`
	}

	tx, err := db.Begin()
	if err != nil {
		return false, err.Error()
	}
	defer tx.Rollback()

	month := time.Now().Format("200601")
	tableName := "tasks_" + month

	_, err = tx.Exec("UPDATE orders SET successful_keys = successful_keys + 1, key_update_time = ? WHERE id = (SELECT secret_key_id FROM "+tableName+" WHERE task_number = ?)", time.Now(), orderStatus.OrderNumber)
	if err != nil {
		return false, err.Error()
	}

	_, err = tx.Exec("UPDATE "+tableName+" SET task_status = ?, order_update_time = ?, remarks = ? WHERE task_number = ?",
		orderStatus.Status, time.Now(), orderStatus.Reason, orderStatus.OrderNumber)
	if err != nil {
		return false, err.Error()
	}

	err = tx.Commit()
	if err != nil {
		return false, err.Error()
	}

	return true, "ok"
}
