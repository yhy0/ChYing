package db

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// HTTPBody 结构体用于存储请求和响应的原始数据
type HTTPBody struct {
	ID          int64  `json:"id"`
	RequestRaw  string `json:"request_raw"`
	ResponseRaw string `json:"response_raw"`
}

// GetHistoryFromSQLite 从 SQLite 获取历史记录
func GetHistoryFromSQLite(page, limit int) ([]HTTPHistory, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var history []HTTPHistory
	offset := (page - 1) * limit

	err := GlobalDB.Model(&HTTPHistory{}).
		Order("id desc").
		Limit(limit).
		Offset(offset).
		Find(&history).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query history from SQLite: %w", err)
	}

	return history, nil
}

// GetHttpData 获取HTTP数据详情
func GetHttpData(id int) (*HTTPBody, error) {
	return getHttpDataFromSQLite(id)
}

// getHttpDataFromSQLite 从 SQLite 获取 HTTP 数据
func getHttpDataFromSQLite(id int) (*HTTPBody, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// 通过 request_id 查询请求和响应数据
	var req Request
	var res Response

	// 查询请求数据
	err := GlobalDB.Model(&Request{}).Where("request_id = ?", id).First(&req).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query request: %w", err)
	}

	// 查询响应数据
	err = GlobalDB.Model(&Response{}).Where("request_id = ?", id).First(&res).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query response: %w", err)
	}

	httpBody := &HTTPBody{
		ID:          int64(id),
		RequestRaw:  req.RequestRaw,
		ResponseRaw: res.ResponseRaw,
	}

	return httpBody, nil
}

// ClearAllHistoryData 清空所有历史数据
func ClearAllHistoryData() error {
	return clearSQLiteHistory()
}

// clearSQLiteHistory 清空 SQLite 历史数据
func clearSQLiteHistory() error {
	if GlobalDB == nil {
		return fmt.Errorf("database not initialized")
	}

	// 清空主历史表
	err := GlobalDB.Exec("DELETE FROM http_histories").Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to clear http_histories: %w", err)
	}

	// 清空其他相关表
	tables := []string{"requests", "responses"}
	for _, table := range tables {
		err := GlobalDB.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果表不存在，忽略错误
			continue
		}
	}

	return nil
}
