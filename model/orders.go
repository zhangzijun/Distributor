package model

import (
	"database/sql"
	"time"
)

type Order struct {
	ID                int       `json:"id"`
	MobileBrand       string    `json:"mobile_brand" form:"mobile_brand"`
	MacAddr           string    `json:"mac_addr" form:"mac_addr"`
	KeyGenerationTime time.Time `json:"key_generation_time"`
	TotalKeys         int       `json:"total_keys" form:"total_keys"`
	RemainingKeys     int       `json:"remaining_keys" form:"remaining_keys"`
	SuccessfulKeys    int       `json:"successful_keys" form:"successful_keys"`
	KeyUpdateTime     time.Time `json:"key_update_time"`
}

func (o *Order) FormattedGenerationTime() string {
	return o.KeyGenerationTime.Format("2006-01-02 15:04:05")
}

func (o *Order) FormattedUpdateTime() string {
	return o.KeyUpdateTime.Format("2006-01-02 15:04:05")
}

func CreateOrderTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		mobile_brand TEXT NOT NULL,
		mac_addr TEXT NOT NULL,
		key_generation_time DATETIME NOT NULL,
		total_keys INTEGER NOT NULL DEFAULT 0,
		remaining_keys INTEGER NOT NULL DEFAULT 0,
		successful_keys INTEGER NOT NULL DEFAULT 0,
		key_update_time DATETIME NOT NULL
	);`
	_, err := db.Exec(query)
	return err
}

func AddOrder(db *sql.DB, order *Order) error {
	query := `
	INSERT INTO orders (mobile_brand, mac_addr, key_generation_time, total_keys, remaining_keys, successful_keys, key_update_time)
	VALUES (?, ?, ?, ?, ?, ?, ?);`
	result, err := db.Exec(query, order.MobileBrand, order.MacAddr, order.KeyGenerationTime, order.TotalKeys, order.RemainingKeys, order.SuccessfulKeys, order.KeyUpdateTime)
	if err != nil {
		return err
	}

	// 获取插入的 order ID
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	order.ID = int(lastInsertID)
	return nil
}

func UpdateOrder(db *sql.DB, order *Order) error {
	query := `
	UPDATE orders
	SET mobile_brand = ?,mac_addr = ?, key_generation_time = ?, total_keys = ?, remaining_keys = ?, successful_keys = ?, key_update_time = ?
	WHERE id = ?;`
	_, err := db.Exec(query, order.MobileBrand, order.MacAddr, order.KeyGenerationTime, order.TotalKeys, order.RemainingKeys, order.SuccessfulKeys, order.KeyUpdateTime, order.ID)
	return err
}

func DeleteOrder(db *sql.DB, id int) error {
	query := `DELETE FROM orders WHERE id = ?;`
	_, err := db.Exec(query, id)
	return err
}

func GetOrder(db *sql.DB, id int) (*Order, error) {
	query := `SELECT id, mobile_brand, mac_addr, key_generation_time, total_keys, remaining_keys, successful_keys, key_update_time FROM orders WHERE id = ?;`
	row := db.QueryRow(query, id)
	var order Order
	err := row.Scan(&order.ID, &order.MobileBrand, &order.MacAddr, &order.KeyGenerationTime, &order.TotalKeys, &order.RemainingKeys, &order.SuccessfulKeys, &order.KeyUpdateTime)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// 获取分页订单
func GetOrdersLimit(db *sql.DB, offset, limit int) ([]Order, error) {
	rows, err := db.Query("SELECT id, mobile_brand, mac_addr, key_generation_time, total_keys, remaining_keys, successful_keys, key_update_time FROM orders LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, err
	}

	orders := []Order{}
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.ID, &order.MobileBrand, &order.MacAddr, &order.KeyGenerationTime, &order.TotalKeys, &order.RemainingKeys, &order.SuccessfulKeys, &order.KeyUpdateTime); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func GetAllOrders(db *sql.DB) ([]Order, error) {
	query := `SELECT id, mobile_brand, mac_addr, key_generation_time, total_keys, remaining_keys, successful_keys, key_update_time FROM orders;`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []Order
	for rows.Next() {
		var order Order
		err := rows.Scan(&order.ID, &order.MobileBrand, &order.MacAddr, &order.KeyGenerationTime, &order.TotalKeys, &order.RemainingKeys, &order.SuccessfulKeys, &order.KeyUpdateTime)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}
