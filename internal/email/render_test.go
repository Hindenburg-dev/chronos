package email

import (
	"strings"
	"testing"

	"chronos/internal/codeforces"
)

func TestMarkdownToHTMLWithCodeforces(t *testing.T) {
	date := "2026-09-07"
	planMD := `## 今日关键日程
08:00 - 09:30 汇编语言(I)
10:05 - 11:40 离散数学
- [ ] 算法特训刷题两道: [4A](https://codeforces.com/problemset/problem/4/A)
- [x] 晨间拉伸与背单词`

	cfProblems := []codeforces.ProblemDetail{
		{
			Item: codeforces.ProblemItem{
				ID:        1,
				Code:      "4A",
				ContestID: 4,
				Index:     "A",
				Name:      "Watermelon",
				Rating:    800,
				URL:       "https://codeforces.com/problemset/problem/4/A",
			},
			Translation: `### 【题目大意】
给定一个西瓜的重量 \(w\) 公斤，判断能否将这个西瓜切成两个正整数的部分。

### 【输入输出与数据范围】
- **输入**：一行一个整数 \(w\)，满足 \(1 \le w \le 100\)。
- **输出**：输出 ` + "`YES`" + ` 或 ` + "`NO`" + `。

### 【💡 思维切入点】
- 两个偶数相加一定是偶数。
- 核心判断：` + "`w % 2 == 0 && w > 2`" + `。`,
		},
	}

	htmlResult := MarkdownToHTML(date, planMD, cfProblems...)

	// 1. 验证基础卡片元素
	if !strings.Contains(htmlResult, "Chronos · 智能日程规划") {
		t.Error("缺少头部品牌标识")
	}
	if !strings.Contains(htmlResult, date) {
		t.Error("缺少目标日期信息")
	}

	// 2. 验证 Codeforces 特训卡片
	if !strings.Contains(htmlResult, "Codeforces 每日特训") {
		t.Error("缺少 Codeforces 专区标题")
	}
	if !strings.Contains(htmlResult, "Watermelon") || !strings.Contains(htmlResult, "4A") {
		t.Error("缺少题目名称或题号")
	}
	if !strings.Contains(htmlResult, "Rating: 800") {
		t.Error("缺少难度徽章")
	}
	if !strings.Contains(htmlResult, "https://codeforces.com/problemset/problem/4/A") {
		t.Error("缺少原题直达链接")
	}
	if !strings.Contains(htmlResult, "给定一个西瓜的重量") {
		t.Error("缺少中文题意解析内容")
	}

	// 3. 验证 Markdown 格式渲染：标题、列表、加粗、行内代码
	if !strings.Contains(htmlResult, "【题目大意】") {
		t.Error("题目大意标题未渲染")
	}
	if !strings.Contains(htmlResult, "<ul") || !strings.Contains(htmlResult, "<li") {
		t.Error("列表项未渲染为 <ul> <li>")
	}
	if !strings.Contains(htmlResult, "<strong") {
		t.Error("加粗标签未渲染")
	}
	if !strings.Contains(htmlResult, "<code") {
		t.Error("行内代码标签未渲染")
	}

	// 4. 验证数学公式渲染
	if !strings.Contains(htmlResult, "1 ≤ w ≤ 100") {
		t.Error("LaTeX 数学公式未正确转为 1 ≤ w ≤ 100")
	}

	// 5. 验证超链接渲染
	if !strings.Contains(htmlResult, `<a href="https://codeforces.com/problemset/problem/4/A"`) {
		t.Error("正文中的 Markdown 链接未渲染为 <a> 标签")
	}

	// 6. 验证自适应宽度属性存在 (防止横向滑动)
	if !strings.Contains(htmlResult, "max-width: 620px") || !strings.Contains(htmlResult, "box-sizing: border-box") {
		t.Error("缺少自适应布局及 box-sizing 约束")
	}
}
