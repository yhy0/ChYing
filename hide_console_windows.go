//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

func init() {
	// 只在非终端环境下隐藏控制台窗口（即双击运行时）
	// 如果是从 CMD/PowerShell 运行，不隐藏
	kernel32 := syscall.NewLazyDLL("kernel32.dll")

	// 检查是否有父进程的控制台
	// GetConsoleProcessList 返回附加到控制台的进程数量
	// 如果只有 1 个进程（当前进程），说明是双击运行，系统自动创建了控制台
	// 如果有多个进程，说明是从 CMD/PowerShell 运行，继承了父进程的控制台
	getConsoleProcessList := kernel32.NewProc("GetConsoleProcessList")
	pids := make([]uint32, 16)
	count, _, _ := getConsoleProcessList.Call(
		uintptr(unsafe.Pointer(&pids[0])),
		uintptr(len(pids)),
	)

	// 只有当控制台只属于当前进程时才隐藏
	if count == 1 {
		user32 := syscall.NewLazyDLL("user32.dll")
		getConsoleWindow := kernel32.NewProc("GetConsoleWindow")
		showWindow := user32.NewProc("ShowWindow")

		hwnd, _, _ := getConsoleWindow.Call()
		if hwnd != 0 {
			showWindow.Call(hwnd, 0) // SW_HIDE = 0
		}
	}
}
