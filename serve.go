package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"github.com/chromedp/chromedp"
	"github.com/russross/blackfriday/v2"
	_ "image/png"
	"io/ioutil"
	"log"
)

// MarkdownToHTML 将Markdown文本转换为HTML。
func MarkdownToHTML(md string) string {
	output := blackfriday.Run([]byte(md))
	return string(output)
}
func HtmlToImage(htmlContent string, message string) {
	// HTMLToImage 将HTML内容渲染成图片。
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// 生成 HTML 内容

	// 设置截图选项
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath("C:\\Program Files (x86)\\chromiumbrowser\\Chromium.exe"),
		chromedp.Flag("headless", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.WindowSize(780, 2600),
	)

	// 创建新的上下文
	ctxt, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// 创建chromedp的上下文
	ctx, cancel = chromedp.NewContext(ctxt)
	defer cancel()

	// 打开网页
	var buf []byte
	if err := chromedp.Run(ctx, chromedp.Navigate("data:text/html,"+"<meta charset=\"UTF-8\">"+htmlContent), chromedp.CaptureScreenshot(&buf)); err != nil {
		log.Fatal(err)
	}

	// 将截图数据写入文件
	if err := ioutil.WriteFile("C://xampp/htdocs/image/"+getMD5Hash(message)+".png", buf, 0644); err != nil {
		log.Fatal(err)
	}
}

func getMD5Hash(text string) string {

	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
