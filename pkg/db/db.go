package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yhy0/ChYing/conf/file"
	"github.com/yhy0/logging"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

/**
  @author: yhy
  @since: 2024/9/10
  @desc: //TODO
**/

var GlobalDB *gorm.DB

// dbWriteMutex 用于保护 SQLite 写操作，防止 "database is locked" 错误
var dbWriteMutex sync.Mutex

// DBError 数据库错误事件，用于通知前端
type DBError struct {
	Operation string `json:"operation"`
	Error     string `json:"error"`
	Time      string `json:"time"`
}

// DBErrorCallback 数据库错误回调函数
var DBErrorCallback func(err DBError)

// SetDBErrorCallback 设置数据库错误回调
func SetDBErrorCallback(callback func(err DBError)) {
	DBErrorCallback = callback
}

// notifyDBError 通知数据库错误
func notifyDBError(operation string, err error) {
	if DBErrorCallback != nil && err != nil {
		DBErrorCallback(DBError{
			Operation: operation,
			Error:     err.Error(),
			Time:      time.Now().Format("2006-01-02 15:04:05"),
		})
	}
}

// WithWriteLock 在写锁保护下执行数据库操作
func WithWriteLock(fn func() error) error {
	dbWriteMutex.Lock()
	defer dbWriteMutex.Unlock()
	return fn()
}

// RetryOnLocked 带重试的数据库写操作
func RetryOnLocked(operation string, fn func() error, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = WithWriteLock(fn)
		if err == nil {
			return nil
		}
		// 检查是否是锁定错误
		if strings.Contains(err.Error(), "database is locked") ||
			strings.Contains(err.Error(), "SQLITE_BUSY") {
			// 等待后重试
			time.Sleep(time.Duration(50*(i+1)) * time.Millisecond)
			continue
		}
		// 其他错误直接返回
		break
	}
	if err != nil {
		notifyDBError(operation, err)
	}
	return err
}

func Init(project string, DBType string) {
	InitSQL(project, DBType)
}

func InitSQL(project string, DBType string) {
	var dialector gorm.Dialector
	var err error
	var dbPath string

	// 项目名称（去掉 .db 后缀）
	projectName := strings.TrimSuffix(project, ".db")

	// 数据库文件放在 db/<projectName>/<projectName>.db 子目录中
	projectDir := path.Join(file.ChyingDir, "db", projectName)

	// 确保项目目录存在
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		logging.Logger.Errorf("创建项目目录失败: %v", err)
		return
	}

	dbPath = path.Join(projectDir, projectName+".db")
	// 启用 WAL 模式和设置 busy_timeout 来减少锁冲突
	// _journal_mode=WAL: 使用 Write-Ahead Logging，允许并发读写
	// _busy_timeout=5000: 等待锁释放的超时时间（毫秒）
	// _synchronous=NORMAL: 平衡性能和安全性
	dialector = sqlite.Open(dbPath + "?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL")

	// 创建 Gorm 的 logger 实现
	GlobalDB, err = gorm.Open(dialector, &gorm.Config{
		Logger: &Logger{
			Logger:                    logging.Logger,
			SlowThreshold:             3 * time.Second,
			IgnoreRecordNotFoundError: true,
			// 可以设置日志级别
			GormLogger: logger.New(
				logging.Logger, // io.Writer
				logger.Config{},
			),
		},
	})

	if err != nil {
		logging.Logger.Errorf("db.Setup err: %v", err)
		return
	}

	sqlDB, err := GlobalDB.DB()
	if err != nil {
		logging.Logger.Errorf("获取数据库连接失败: %v", err)
		return
	}

	// SQLite 并发优化：限制连接数为 1，避免多连接导致的锁冲突
	// SQLite 本身是单写多读的，多个连接写入会导致 "database is locked"
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)

	// 连接保持时间
	sqlDB.SetConnMaxLifetime(time.Hour)

	if GlobalDB == nil {
		logging.Logger.Errorln("db.Setup err: db connect failed")
		return
	}

	err = GlobalDB.AutoMigrate(&Request{}, &Response{}, &HTTPHistory{}, &SCopilot{}, &IPInfo{}, &ScanTarget{}, &Vulnerability{}, &ClaudeSession{})

	if err != nil {
		logging.Logger.Errorf("db AutoMigrate err: %v", err)
		return
	}
	logging.Logger.Infoln("Sql Connection Established.")
}

func CreateMySQLDB(dialector gorm.Dialector) error {
	// 连接到数据库
	db, err := gorm.Open(dialector)
	if err != nil {
		logging.Logger.Errorf("failed to connect to database: %v", err)
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 创建数据库
	err = db.Exec("CREATE DATABASE IF NOT EXISTS cy").Error
	if err != nil {
		logging.Logger.Errorf("failed to create database: %v", err)
		return fmt.Errorf("failed to create database: %w", err)
	}

	logging.Logger.Println("Database created successfully!")
	return nil
}

// Logger 实现了 Gorm 的 Logger 接口
type Logger struct {
	Logger                    *logrus.Logger
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	GormLogger                logger.Interface
}

var (
	infoStr      = "%s\n[info] "
	warnStr      = "%s\n[warn] "
	errStr       = "%s\n[error] "
	traceStr     = "%s\n[%.3fms] [rows:%v] %s"
	traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
	traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
)

func (l *Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.Logger.Level >= logrus.InfoLevel {
		l.Logger.Infof(warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.Logger.Level >= logrus.WarnLevel {
		l.Logger.Warnf(warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.Logger.Level >= logrus.ErrorLevel {
		l.Logger.Errorf(errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.Logger.Level <= logrus.FatalLevel {
		return
	}
	// 调用 fc 函数获取 SQL 语句和影响的行数
	sql, rows := fc()

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.Logger.Level >= logrus.ErrorLevel && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		if rows == -1 {
			l.Logger.Printf(traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Logger.Printf(traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.Logger.Level >= logrus.WarnLevel:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.Logger.Printf(traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Logger.Printf(traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.Logger.Level == logrus.InfoLevel:
		if rows == -1 {
			l.Logger.Printf(traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Logger.Printf(traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

// LogMode 这里全局由一开始日志初始化控制  logging.Logger = logging.New(true, "", "SuWen", true)
func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

// OpenDatabase 打开指定路径的数据库（用于统计查询等只读操作）
func OpenDatabase(dbPath string) (*gorm.DB, error) {
	// 启用 WAL 模式和设置 busy_timeout 来减少锁冲突
	dialector := sqlite.Open(dbPath + "?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&mode=ro")

	database, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 静默模式，不输出日志
	})

	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	return database, nil
}
