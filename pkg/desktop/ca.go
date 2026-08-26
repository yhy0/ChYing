package desktop

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type CAInstallResult struct {
	Installed bool
	Message   string
}

// InstallCACertificate 把 Proxify MITM CA 写入当前用户信任库。失败不阻断代理启动。
func InstallCACertificate(certPath string) CAInstallResult {
	trimmed := strings.TrimSpace(certPath)
	if trimmed == "" {
		return CAInstallResult{Message: "未提供 CA 证书路径"}
	}
	info, err := os.Stat(trimmed)
	if err != nil || info.IsDir() {
		return CAInstallResult{Message: fmt.Sprintf("CA 证书不存在: %s", trimmed)}
	}

	switch runtime.GOOS {
	case "linux":
		return installLinuxCA(trimmed)
	case "darwin":
		return installDarwinCA(trimmed)
	default:
		return CAInstallResult{Message: fmt.Sprintf("%s 需手动安装 CA：%s", runtime.GOOS, trimmed)}
	}
}

func installDarwinCA(certPath string) CAInstallResult {
	home, err := os.UserHomeDir()
	if err != nil {
		return CAInstallResult{Message: "无法确定用户目录，未安装 CA"}
	}
	keychain := filepath.Join(home, "Library", "Keychains", "login.keychain-db")
	if _, err := os.Stat(keychain); err != nil {
		keychain = filepath.Join(home, "Library", "Keychains", "login.keychain")
	}
	cmd := exec.Command("security", "add-trusted-cert", "-d", "-r", "trustRoot", "-k", keychain, certPath)
	if output, err := runTimed(cmd, 15*time.Second); err != nil {
		return CAInstallResult{Message: fmt.Sprintf("macOS 登录钥匙串安装失败（可双击 %s 手动信任）: %s", certPath, compactOutput(output, err))}
	}
	return CAInstallResult{Installed: true, Message: "已写入 macOS 登录钥匙串"}
}

func installLinuxCA(certPath string) CAInstallResult {
	var notes []string
	if nss := installLinuxNSS(certPath); nss != "" {
		notes = append(notes, nss)
	}
	if system := installLinuxSystemCA(certPath); system != "" {
		notes = append(notes, system)
	}
	if len(notes) == 0 {
		return CAInstallResult{Message: fmt.Sprintf("未找到 certutil / update-ca-certificates，请手动信任 %s", certPath)}
	}
	installed := strings.Contains(strings.Join(notes, "; "), "已")
	return CAInstallResult{Installed: installed, Message: strings.Join(notes, "；")}
}

func installLinuxNSS(certPath string) string {
	certutil, err := exec.LookPath("certutil")
	if err != nil {
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "无法确定用户目录，跳过 NSS"
	}
	nssDir := filepath.Join(home, ".pki", "nssdb")
	if err := os.MkdirAll(nssDir, 0o700); err != nil {
		return fmt.Sprintf("无法创建 %s", nssDir)
	}
	db := "sql:" + nssDir
	if _, err := os.Stat(filepath.Join(nssDir, "cert9.db")); err != nil {
		if output, err := runTimed(exec.Command(certutil, "-N", "-d", db, "--empty-password"), 8*time.Second); err != nil {
			return fmt.Sprintf("初始化 NSS 失败: %s", compactOutput(output, err))
		}
	}
	_, _ = runTimed(exec.Command(certutil, "-D", "-d", db, "-n", "ChYing-Proxify-CA"), 5*time.Second)
	if output, err := runTimed(exec.Command(certutil, "-A", "-d", db, "-t", "C,,", "-n", "ChYing-Proxify-CA", "-i", certPath), 8*time.Second); err != nil {
		return fmt.Sprintf("NSS 安装失败: %s", compactOutput(output, err))
	}
	return "已写入用户 NSS（Chromium / 部分 curl）"
}

func installLinuxSystemCA(certPath string) string {
	if _, err := exec.LookPath("update-ca-certificates"); err != nil {
		if _, err := exec.LookPath("update-ca-trust"); err != nil {
			return ""
		}
		dest := "/etc/pki/ca-trust/source/anchors/chying-proxify.crt"
		if output, err := runTimed(exec.Command("sudo", "-n", "install", "-m", "644", certPath, dest), 8*time.Second); err != nil {
			return fmt.Sprintf("系统信任库需要免密 sudo（update-ca-trust）: %s", compactOutput(output, err))
		}
		if output, err := runTimed(exec.Command("sudo", "-n", "update-ca-trust", "extract"), 20*time.Second); err != nil {
			return fmt.Sprintf("update-ca-trust 失败: %s", compactOutput(output, err))
		}
		return "已写入系统信任库"
	}
	dest := "/usr/local/share/ca-certificates/chying-proxify.crt"
	if output, err := runTimed(exec.Command("sudo", "-n", "install", "-d", "-m", "755", filepath.Dir(dest)), 8*time.Second); err != nil {
		return fmt.Sprintf("系统信任库需要免密 sudo（update-ca-certificates）: %s", compactOutput(output, err))
	}
	if output, err := runTimed(exec.Command("sudo", "-n", "install", "-m", "644", certPath, dest), 8*time.Second); err != nil {
		return fmt.Sprintf("系统信任库需要免密 sudo（update-ca-certificates）: %s", compactOutput(output, err))
	}
	if output, err := runTimed(exec.Command("sudo", "-n", "update-ca-certificates"), 20*time.Second); err != nil {
		return fmt.Sprintf("update-ca-certificates 失败: %s", compactOutput(output, err))
	}
	return "已写入系统信任库"
}

func runTimed(cmd *exec.Cmd, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	timed := exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
	timed.Env = cmd.Env
	timed.Dir = cmd.Dir
	output, err := timed.CombinedOutput()
	return string(output), err
}

func compactOutput(output string, err error) string {
	text := strings.TrimSpace(output)
	if text == "" && err != nil {
		return err.Error()
	}
	if err != nil && text != "" {
		return fmt.Sprintf("%s: %s", err.Error(), text)
	}
	return text
}

func LocalhostServiceURL(addr string) string {
	host, port, ok := strings.Cut(strings.TrimSpace(addr), ":")
	if !ok {
		return "http://127.0.0.1/" + strings.TrimPrefix(addr, "/")
	}
	if host == "0.0.0.0" || host == "::" || host == "[::]" || host == "" {
		host = "127.0.0.1"
	}
	return fmt.Sprintf("http://%s:%s", host, port)
}
