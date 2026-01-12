package util

import (
	"bufio"
	"github.com/yhy0/logging"
	"io"

	"os"
	"strings"
)

// Exists 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) // os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// IsDir 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// IsFile 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}

// ReadFileAsLine 读取文件并返回一个字符串切片
func ReadFileAsLine(path string) ([]string, error) {
	lineSlice := make([]string, 0)

	if !IsFile(path) {
		return nil, os.ErrNotExist
	}
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		logging.Logger.Errorf("Open file %s error", path)
		return nil, err
	}

	buf := bufio.NewReader(file)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		lineSlice = append(lineSlice, line)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				logging.Logger.Errorf("Read file %s error", path)
				return nil, err
			}
		}
	}

	return lineSlice, nil
}

func WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0)
}
