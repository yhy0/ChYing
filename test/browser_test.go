package test

import (
	"context"
	"fmt"
	"github.com/yhy0/logging"
	"testing"

	"github.com/chromedp/chromedp"
)

/**
   @author yhy
   @since 2025/4/28
   @desc //TODO
**/

var browserCancel context.CancelFunc

func InitBrowser(proxy string) error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// 无头模式
		chromedp.Flag("headless", false),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.WindowSize(1920, 1080),
		chromedp.ProxyServer(proxy),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	bctx, ctxCancel := chromedp.NewContext(allocCtx,
		chromedp.WithLogf(logging.Logger.Printf),
	)

	// 保存两个取消函数，确保资源正确释放
	browserCancel = func() {
		ctxCancel()
		cancel()
	}

	// 启动浏览器
	if err := chromedp.Run(bctx); err != nil {
		browserCancel() // 出错时释放资源
		return fmt.Errorf("启动Chrome失败: %w", err)
	}

	return nil
}

func TestBrowser(t *testing.T) {
	InitBrowser("http://127.0.0.1:8888")
}
