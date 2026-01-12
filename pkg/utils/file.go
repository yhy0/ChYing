package utils

import (
    "github.com/yhy0/logging"
    "os"
)

/**
   @author yhy
   @since 2024/9/23
   @desc //TODO
**/

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

func ReadFile(path string) (string, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        logging.Logger.Errorln("read file error:", err)
        return "", err
    }
    return string(data), nil
}

func WriteFile(path string, content string) error {
    file, err := os.Create(path)
    if err != nil {
        logging.Logger.Errorln("write file error:", err)
        return err
    }
    defer file.Close() // 确保在函数结束时关闭文件
    _, err = file.Write([]byte(content))
    if err != nil {
        logging.Logger.Errorln("write file error:", err)
        return err
    }
    return nil
}
