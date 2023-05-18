package utils

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

/**
  @author: yhy
  @since: 2023/5/11
  @desc: //TODO
**/

// IsPortOccupied 判断端口号是否被占用
func IsPortOccupied(port int) bool {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return true // 端口已被占用
	}
	listener.Close()
	return false // 端口未被占用
}

// GetRandomUnusedPort 随机获取一个在 60000 以上的端口号
func GetRandomUnusedPort() (int, error) {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())

	// 定义起始端口号和端口号范围，并生成一个随机整数
	basePort := 60000
	portRange := 65535 - basePort + 1
	randomOffset := rand.Intn(portRange)

	// 计算出最终的端口号
	port := basePort + randomOffset

	// 创建一个 TCP 地址对象
	addr := &net.TCPAddr{IP: nil, Port: port}

	// 使用 ListenTCP 函数创建一个新的 TCP 监听器，并返回一个 TCP 地址对象
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}

	// 关闭监听器
	defer listener.Close()

	// 获取监听器的地址对象，并返回其端口号作为结果
	return listener.Addr().(*net.TCPAddr).Port, nil
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
