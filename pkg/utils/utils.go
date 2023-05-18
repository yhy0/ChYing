package utils

import (
	"fmt"
	"net"
	"os"
)

/**
  @author: yhy
  @since: 2023/5/11
  @desc: //TODO
**/

func IsPortOccupied(port int) bool {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return true // 端口已被占用
	}
	listener.Close()
	return false // 端口未被占用
}

// Exists 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
