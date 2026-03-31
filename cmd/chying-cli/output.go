package main

import (
	"fmt"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

func timestamp() string {
	return time.Now().Format("15:04:05")
}

func printStatus(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s[%s]%s %s%s%s\n", colorGray, timestamp(), colorReset, colorCyan, msg, colorReset)
}

func printTraffic(method, url, status, length string) {
	statusColor := colorGreen
	if len(status) > 0 && status[0] != '2' {
		statusColor = colorYellow
	}
	fmt.Printf("%s[%s]%s %s->%s %s %s %s[%s]%s %s\n",
		colorGray, timestamp(), colorReset,
		colorBlue, colorReset,
		method, url,
		statusColor, status, colorReset,
		length,
	)
}

func printVuln(level, vulnType, target, param string) {
	var icon, levelColor string
	switch level {
	case "Critical", "High":
		icon = "[!]"
		levelColor = colorRed
	case "Medium":
		icon = "[*]"
		levelColor = colorYellow
	case "Low":
		icon = "[+]"
		levelColor = colorGreen
	default:
		icon = "[-]"
		levelColor = colorGray
	}

	paramInfo := ""
	if param != "" {
		paramInfo = fmt.Sprintf(" | Param: %s", param)
	}

	fmt.Printf("%s[%s]%s %s %s[%s]%s %s | %s%s\n",
		colorGray, timestamp(), colorReset,
		icon,
		levelColor, level, colorReset,
		vulnType, target, paramInfo,
	)
}
