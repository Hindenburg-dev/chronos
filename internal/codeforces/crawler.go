package codeforces

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	htmlTagRegex = regexp.MustCompile(`<[^>]+>`)
)

type Crawler struct {
	httpClient *http.Client
}

func NewCrawler() *Crawler {
	return &Crawler{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// FetchProblemStatement 从 Codeforces 网页抓取题目英文原题与要求说明
func (c *Crawler) FetchProblemStatement(contestID int, index string) (string, error) {
	url := fmt.Sprintf("https://codeforces.com/problemset/problem/%d/%s", contestID, index)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Chronos/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("网络请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Codeforces 返回状态码异常: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	html := string(bodyBytes)

	// 提取 class="problem-statement"
	startIdx := strings.Index(html, `<div class="problem-statement">`)
	if startIdx == -1 {
		// 备用查找
		startIdx = strings.Index(html, `problem-statement`)
	}

	if startIdx != -1 {
		raw := html[startIdx:]
		endIdx := strings.Index(raw, `</div></div></div>`)
		if endIdx != -1 {
			raw = raw[:endIdx]
		} else if len(raw) > 6000 {
			raw = raw[:6000]
		}
		clean := cleanHTMLText(raw)
		if len(clean) > 50 {
			return clean, nil
		}
	}

	// 兜底返回基础信息
	return fmt.Sprintf("Problem %d%s (Details at %s)", contestID, index, url), nil
}

func cleanHTMLText(s string) string {
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "<br />", "\n")
	s = strings.ReplaceAll(s, "</p>", "\n\n")
	s = strings.ReplaceAll(s, "</div>", "\n")
	s = strings.ReplaceAll(s, "</pre>", "\n")
	s = htmlTagRegex.ReplaceAllString(s, "")

	// HTML 实体替换
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&le;", "≤")
	s = strings.ReplaceAll(s, "&ge;", "≥")
	s = strings.ReplaceAll(s, "&ne;", "≠")
	s = strings.ReplaceAll(s, "&times;", "×")
	s = strings.ReplaceAll(s, "&hellip;", "...")

	lines := strings.Split(s, "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleanedLines = append(cleanedLines, trimmed)
		}
	}
	return strings.Join(cleanedLines, "\n")
}
