package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetFileSize 获取文件大小（字节）
func GetFileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

// FormatFileSize 格式化文件大小为可读字符串
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetDBFileInfo 获取数据库文件信息
func GetDBFileInfo(dbPath string) (map[string]interface{}, error) {
	fileInfo, err := os.Stat(dbPath)
	if err != nil {
		return nil, err
	}

	size := fileInfo.Size()

	return map[string]interface{}{
		"size_bytes":     size,
		"size_formatted": FormatFileSize(size),
		"modified_time":  fileInfo.ModTime(),
		"file_name":      filepath.Base(dbPath),
		"file_path":      dbPath,
	}, nil
}
