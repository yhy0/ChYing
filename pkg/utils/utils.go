package utils

import (
    "fmt"
    "github.com/google/uuid"
    regexp "github.com/wasilibs/go-re2"
    "math/rand"
    "net"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "time"
)

/**
  @author: yhy
  @since: 2023/5/11
  @desc: //TODO
**/

func init() {
    rand.Seed(time.Now().Unix())
}

func UUID() string {
    return uuid.NewString()
}

func GetTitle(body string) string {
    titleReg := regexp.MustCompile(`<title>([\s\S]{1,200})</title>`)
    title := titleReg.FindStringSubmatch(body)
    if len(title) > 1 {
        return title[1]
    }
    return ""
}

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

func OpenFolder(path string) error {
    var err error
    
    switch runtime.GOOS {
    case "windows":
        err = exec.Command("cmd", "/c", "explorer", path).Start()
    case "linux":
        err = execCmd("xdg-open", path)
    case "darwin":
        err = execCmd("open", path)
    default:
        err = fmt.Errorf("unsupported platform")
    }
    
    return err
}

func execCmd(cmd string, args ...string) error {
    command := exec.Command(cmd, args...)
    command.Stdout = os.Stdout
    command.Stderr = os.Stderr
    return command.Run()
}

// GetDBFiles 返回指定目录下所有 .db 文件的文件名
func GetDBFiles(dir string) (map[string]string, error) {
    var dbFiles = make(map[string]string)
    
    // 使用 filepath.Walk 遍历目录
    err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        // 检查文件是否是 .db 文件
        if !info.IsDir() && filepath.Ext(info.Name()) == ".db" {
            dbFiles[info.Name()] = formatSize(info.Size())
        }
        return nil
    })
    
    if err != nil {
        return nil, err
    }
    
    return dbFiles, nil
}

// formatSize 将字节数转换为更易读的格式
// formatSize 将字节数转换为更易读的格式
func formatSize(size int64) string {
    const (
        _KB = 1024
        _MB = _KB * 1024
        _GB = _MB * 1024
    )
    
    switch {
    case size >= _GB:
        return fmt.Sprintf("%.2f GB", float64(size)/float64(_GB))
    case size >= _MB:
        return fmt.Sprintf("%.2f MB", float64(size)/float64(_MB))
    case size >= _KB:
        return fmt.Sprintf("%.2f KB", float64(size)/float64(_KB))
    default:
        return fmt.Sprintf("%d bytes", size)
    }
}
