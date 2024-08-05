package model

import (
	"database/sql"
	"log"
)

var db *sql.DB

func InitDB() {
	println("init db")
	//打开数据库连接
	var err error
	db, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// 连接池设置等
}
