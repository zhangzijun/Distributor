package model

import (
	"database/sql"
	"fmt"
	"time"
)

type Task struct {
	ID               int       `json:"id"`
	TaskNumber       string    `json:"task_number"`
	MobilePlatform   string    `json:"mobile_platform"`
	UID              string    `json:"uid"`
	UIDKey           string    `json:"uid_key"`
	Remarks          string    `json:"remarks"`
	TaskStatus       int       `json:"task_status"`
	TaskCreationTime time.Time `json:"task_creation_time"`
	TaskUpdateTime   time.Time `json:"task_update_time"`
}

func (t *Task) FormattedCreationTime() string {
	return t.TaskCreationTime.Format("2006-01-02 15:04:05")
}

func (t *Task) FormattedUpdateTime() string {
	return t.TaskUpdateTime.Format("2006-01-02 15:04:05")
}

// GetTaskTableName 根据当前日期生成动态表名
func GetTaskTableName() string {
	currentTime := time.Now()
	return fmt.Sprintf("tasks_%d%02d", currentTime.Year(), int(currentTime.Month()))
}

// Retrieve tasks from the database
func GetTasks(db *sql.DB) ([]Task, error) {
	tableName := GetTaskTableName()
	query := fmt.Sprintf(`SELECT * FROM %s`, tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.TaskNumber, &task.MobilePlatform, &task.UID, &task.UIDKey, &task.Remarks, &task.TaskStatus, &task.TaskCreationTime, &task.TaskUpdateTime); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// 获取分页任务
func GetTasksLimit(db *sql.DB, offset, limit int) ([]Task, error) {
	tableName := GetTaskTableName()
	query := fmt.Sprintf(`SELECT id, task_number, mobile_platform, uid, uid_key, remarks, task_status, task_creation_time, task_update_time FROM %s LIMIT ? OFFSET ?`, tableName)
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.TaskNumber, &task.MobilePlatform, &task.UID, &task.UIDKey, &task.Remarks, &task.TaskStatus, &task.TaskCreationTime, &task.TaskUpdateTime); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// Update task in the database
func UpdateTask(db *sql.DB, task Task) error {
	tableName := GetTaskTableName()

	// 查询是否有其他记录使用相同的 task_number，但不是当前记录
	var existingTaskID int
	err := db.QueryRow("SELECT id FROM "+tableName+" WHERE task_number = ? AND id != ?", task.TaskNumber, task.ID).Scan(&existingTaskID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing task number: %w", err)
	}
	println(existingTaskID)
	if existingTaskID != 0 {
		return fmt.Errorf("task number %s already exists", task.TaskNumber)
	}

	query := fmt.Sprintf(`
        UPDATE %s
        SET uid = ?, uid_key = ?, task_status = ?, task_update_time = ?
        WHERE id = ?`, tableName)
	fmt.Print(query, task.UID, task.UIDKey, task.TaskStatus, task.TaskUpdateTime, task.ID)
	result, err := db.Exec(query,
		task.UID, task.UIDKey, task.TaskStatus, task.TaskUpdateTime, task.ID)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if affectedRows == 0 {
		return fmt.Errorf("no rows updated, possibly due to incorrect ID")
	}

	return nil
}

// InsertTask 插入新的任务
func InsertTask(db *sql.DB, task *Task) error {
	tableName := GetTaskTableName()
	query := fmt.Sprintf(`INSERT INTO %s (task_number, mobile_platform, uid, uid_key, task_status, task_creation_time, task_update_time)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`, tableName)
	_, err := db.Exec(query, task.TaskNumber, task.MobilePlatform, task.UID, task.UIDKey, task.TaskStatus, task.TaskCreationTime, task.TaskUpdateTime)
	return err
}
