package email

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"chronos/internal/codeforces"
)

var (
	reBold          = regexp.MustCompile(`\*\*(.*?)\*\*`)
	reItalic        = regexp.MustCompile(`\*([^*]+)\*`)
	reStrike        = regexp.MustCompile(`~~(.*?)~~`)
	reInlineCode    = regexp.MustCompile("`([^`]+)`")
	reLink          = regexp.MustCompile(`\[([^\]]+)\]\((https?:\/\/[^\s\)]+)\)`)
	reMathParen     = regexp.MustCompile(`\\\((\s*[^)]+?\s*)\\\)`)
	reMathDollar    = regexp.MustCompile(`\$([^$\n]+)\$`)
	reTaskDone      = regexp.MustCompile(`^-\s*\[[xX]\]\s*(.*)`)
	reTaskTodo      = regexp.MustCompile(`^-\s*\[\s*\]\s*(.*)`)
	reBulletList    = regexp.MustCompile(`^[-*+]\s+(.*)`)
	reNumberedList  = regexp.MustCompile(`^\d+\.\s+(.*)`)
	reHeading1      = regexp.MustCompile(`^#\s+(.*)`)
	reHeading2      = regexp.MustCompile(`^##\s+(.*)`)
	reHeading3      = regexp.MustCompile(`^###\s+(.*)`)
	reBlockquote    = regexp.MustCompile(`^>\s*(.*)`)
	reHorizontalSep = regexp.MustCompile(`^[-*_]{3,}\s*$`)
	reTableSepCell  = regexp.MustCompile(`^:?-+:?$`)
	reTimeSpan      = regexp.MustCompile(`^(?:[【\[]?\s*)?(\d{1,2}:\d{2}\s*[-~至到]\s*\d{1,2}:\d{2})(?:[】\]]?\s*)(.*)`)
)

// MarkdownToHTML 将日程与 Codeforces 特训转为移动端撑满友好、排版精致的 HTML 邮件
func MarkdownToHTML(date, md string, cfProblems ...codeforces.ProblemDetail) string {
	lines := strings.Split(md, "\n")
	var bodyHTML strings.Builder

	inUl := false
	inOl := false
	inBlockquote := false
	inTable := false
	tableRowCount := 0

	closeLists := func() {
		if inUl {
			bodyHTML.WriteString("</ul>\n")
			inUl = false
		}
		if inOl {
			bodyHTML.WriteString("</ol>\n")
			inOl = false
		}
	}

	closeBlockquote := func() {
		if inBlockquote {
			bodyHTML.WriteString("</blockquote>\n")
			inBlockquote = false
		}
	}

	closeTable := func() {
		if inTable {
			bodyHTML.WriteString("</tbody>\n</table>\n</div>\n")
			inTable = false
			tableRowCount = 0
		}
	}

	closeAll := func() {
		closeLists()
		closeBlockquote()
		closeTable()
	}

	for _, line := range lines {
		rawLine := strings.TrimRight(line, "\r")
		trimmed := strings.TrimSpace(rawLine)

		if trimmed == "" {
			closeAll()
			continue
		}

		// 1. 表格支持 (Markdown Table: | 列1 | 列2 |)
		if isTableRow(trimmed) {
			closeLists()
			closeBlockquote()

			cells := splitTableRow(trimmed)
			if isTableSeparatorRow(cells) {
				continue
			}

			if !inTable {
				inTable = true
				tableRowCount = 0

				bodyHTML.WriteString(`<div style="overflow-x: auto; margin: 14px 0; border-radius: 8px; border: 1px solid #e2e8f0; background-color: #ffffff; -webkit-overflow-scrolling: touch;">` + "\n")
				bodyHTML.WriteString(`<table style="width: 100%; max-width: 100%; border-collapse: collapse; font-size: 12.5px; text-align: left; table-layout: auto;">` + "\n")
				bodyHTML.WriteString(`<thead><tr style="background-color: #f1f5f9; border-bottom: 2px solid #e2e8f0;">` + "\n")
				for _, cell := range cells {
					bodyHTML.WriteString(fmt.Sprintf(`<th style="padding: 8px 10px; font-weight: 700; color: #1e293b; border-right: 1px solid #e2e8f0; word-break: break-word;">%s</th>`+"\n", renderInline(cell)))
				}
				bodyHTML.WriteString("</tr></thead>\n<tbody>\n")
				continue
			}

			tableRowCount++
			bg := "#ffffff"
			if tableRowCount%2 == 0 {
				bg = "#f8fafc"
			}
			bodyHTML.WriteString(fmt.Sprintf(`<tr style="background-color: %s; border-bottom: 1px solid #f1f5f9;">`+"\n", bg))
			for _, cell := range cells {
				bodyHTML.WriteString(fmt.Sprintf(`<td style="padding: 8px 10px; color: #334155; line-height: 1.5; border-right: 1px solid #f1f5f9; word-break: break-word;">%s</td>`+"\n", renderInline(cell)))
			}
			bodyHTML.WriteString("</tr>\n")
			continue
		} else {
			closeTable()
		}

		// 2. 时间轴识别 (非表格中的时间行，如 08:00 - 09:30 任务说明)
		if m := reTimeSpan.FindStringSubmatch(trimmed); len(m) > 2 {
			closeAll()
			timePart := strings.TrimSpace(m[1])
			descPart := renderInline(strings.TrimSpace(m[2]))
			bodyHTML.WriteString(fmt.Sprintf(`
<div style="margin: 8px 0; padding: 6px 10px; background: #f8fafc; border-left: 4px solid #2563eb; border-radius: 0 6px 6px 0;">
  <span style="display: inline-block; background-color: #2563eb; color: #ffffff; font-size: 12px; font-weight: 700; padding: 2px 6px; border-radius: 4px; font-family: Consolas, Monaco, monospace; margin-right: 8px;">%s</span>
  <span style="color: #1e293b; font-size: 13.5px; font-weight: 500; word-break: break-word;">%s</span>
</div>`+"\n", timePart, descPart))
			continue
		}

		// 3. 水平分割线 ---
		if reHorizontalSep.MatchString(trimmed) {
			closeAll()
			bodyHTML.WriteString(`<hr style="border: none; border-top: 1px solid #e2e8f0; margin: 18px 0;" />` + "\n")
			continue
		}

		// 4. 标题 # / ## / ###
		if m := reHeading1.FindStringSubmatch(trimmed); len(m) > 1 {
			closeAll()
			content := renderInline(m[1])
			bodyHTML.WriteString(fmt.Sprintf(`<h1 style="font-size: 17px; font-weight: 700; color: #0f172a; margin: 20px 0 10px 0; border-bottom: 2px solid #2563eb; padding-bottom: 6px;">%s</h1>`+"\n", content))
			continue
		}
		if m := reHeading2.FindStringSubmatch(trimmed); len(m) > 1 {
			closeAll()
			content := renderInline(m[1])
			bodyHTML.WriteString(fmt.Sprintf(`<h2 style="font-size: 15px; font-weight: 700; color: #1e293b; margin: 16px 0 8px 0; border-left: 4px solid #2563eb; padding-left: 8px;">%s</h2>`+"\n", content))
			continue
		}
		if m := reHeading3.FindStringSubmatch(trimmed); len(m) > 1 {
			closeAll()
			content := renderInline(m[1])
			bodyHTML.WriteString(fmt.Sprintf(`<h3 style="font-size: 13.5px; font-weight: 600; color: #334155; margin: 14px 0 6px 0;">%s</h3>`+"\n", content))
			continue
		}

		// 5. 引用块 >
		if m := reBlockquote.FindStringSubmatch(trimmed); len(m) > 1 {
			closeLists()
			if !inBlockquote {
				bodyHTML.WriteString(`<blockquote style="margin: 12px 0; padding: 10px 14px; background-color: #f0fdf4; border-left: 4px solid #10b981; color: #166534; border-radius: 0 8px 8px 0; font-size: 13px; line-height: 1.5; word-break: break-word;">` + "\n")
				inBlockquote = true
			}
			content := renderInline(m[1])
			bodyHTML.WriteString(fmt.Sprintf(`<div style="margin: 3px 0;">%s</div>`+"\n", content))
			continue
		} else {
			closeBlockquote()
		}

		// 6. 任务清单 - [x] / - [ ]
		if m := reTaskDone.FindStringSubmatch(trimmed); len(m) > 1 {
			if !inUl {
				closeLists()
				bodyHTML.WriteString(`<ul style="list-style: none; padding-left: 2px; margin: 8px 0;">` + "\n")
				inUl = true
			}
			content := renderInline(m[1])
			bodyHTML.WriteString(fmt.Sprintf(`<li style="margin: 5px 0; line-height: 1.5; font-size: 13.5px; color: #94a3b8; text-decoration: line-through; word-break: break-word;"><span style="color: #10b981; margin-right: 6px; font-weight: bold; font-size: 14px;">☑</span> %s</li>`+"\n", content))
			continue
		}
		if m := reTaskTodo.FindStringSubmatch(trimmed); len(m) > 1 {
			if !inUl {
				closeLists()
				bodyHTML.WriteString(`<ul style="list-style: none; padding-left: 2px; margin: 8px 0;">` + "\n")
				inUl = true
			}
			content := renderInline(m[1])
			bodyHTML.WriteString(fmt.Sprintf(`<li style="margin: 5px 0; line-height: 1.5; font-size: 13.5px; color: #334155; word-break: break-word;"><span style="color: #64748b; margin-right: 6px; font-size: 14px;">☐</span> %s</li>`+"\n", content))
			continue
		}

		// 7. 无序列表
		if m := reBulletList.FindStringSubmatch(trimmed); len(m) > 1 {
			if !inUl {
				closeLists()
				bodyHTML.WriteString(`<ul style="padding-left: 18px; margin: 8px 0; color: #334155;">` + "\n")
				inUl = true
			}
			content := renderInline(m[1])
			bodyHTML.WriteString(fmt.Sprintf(`<li style="margin: 4px 0; line-height: 1.5; font-size: 13.5px; word-break: break-word;">%s</li>`+"\n", content))
			continue
		}

		// 8. 有序列表
		if m := reNumberedList.FindStringSubmatch(trimmed); len(m) > 1 {
			if !inOl {
				closeLists()
				bodyHTML.WriteString(`<ol style="padding-left: 18px; margin: 8px 0; color: #334155;">` + "\n")
				inOl = true
			}
			content := renderInline(m[1])
			bodyHTML.WriteString(fmt.Sprintf(`<li style="margin: 4px 0; line-height: 1.5; font-size: 13.5px; word-break: break-word;">%s</li>`+"\n", content))
			continue
		}

		// 9. 普通段落
		closeAll()
		content := renderInline(trimmed)
		bodyHTML.WriteString(fmt.Sprintf(`<p style="margin: 6px 0; line-height: 1.6; font-size: 13.5px; color: #334155; word-break: break-word;">%s</p>`+"\n", content))
	}

	closeAll()

	// 渲染 Codeforces 专区 (若有题目)
	cfSectionHTML := renderCFSection(cfProblems)

	// 组装完整的现代响应式 HTML 邮件卡片（控制宽度撑满，防水平滑动）
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
  <title>Chronos 执行计划</title>
  <style>
    * { box-sizing: border-box; }
    body { margin: 0; padding: 0; -webkit-text-size-adjust: 100%%; }
    table { width: 100%% !important; max-width: 100%% !important; border-collapse: collapse !important; }
    td, th { word-break: break-word !important; }
    @media only screen and (max-width: 640px) {
      .card-wrap { width: 100%% !important; border-radius: 0 !important; border-left: none !important; border-right: none !important; }
      .section-box { padding: 14px 10px !important; }
      .banner-box { padding: 18px 12px !important; }
      th, td { padding: 6px 6px !important; font-size: 12px !important; }
    }
  </style>
</head>
<body style="margin: 0; padding: 10px 4px; background-color: #f1f5f9; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; -webkit-font-smoothing: antialiased;">
  <div class="card-wrap" style="max-width: 620px; width: 100%%; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 14px rgba(0, 0, 0, 0.05); border: 1px solid #e2e8f0; box-sizing: border-box;">
    
    <!-- 头部横幅 Banner -->
    <div class="banner-box" style="background: linear-gradient(135deg, #0f172a 0%%, #1e293b 50%%, #2563eb 100%%); padding: 20px 20px; color: #ffffff;">
      <div style="display: flex; align-items: center; justify-content: space-between;">
        <span style="font-size: 19px; font-weight: 800; letter-spacing: 0.5px;">
          ⏳ Chronos · 智能日程规划
        </span>
        <span style="background-color: rgba(255,255,255,0.18); font-size: 11px; font-weight: 600; padding: 3px 8px; border-radius: 9999px;">
          DAILY DISPATCH
        </span>
      </div>
      <div style="margin-top: 10px; font-size: 13.5px; color: #e2e8f0;">
        🎯 规划目标日: <strong style="color: #ffffff; background-color: rgba(37,99,235,0.4); padding: 2px 8px; border-radius: 4px; font-family: Consolas, Monaco, monospace; font-size: 13.5px;">%s</strong>
      </div>
    </div>

    <!-- Codeforces 每日特训专属区域 -->
    %s

    <!-- 正文执行计划区域 -->
    <div class="section-box" style="padding: 20px 16px; color: #1e293b; box-sizing: border-box;">
      <div style="margin-bottom: 12px;">
        <span style="font-size: 15px; font-weight: 700; color: #0f172a; border-left: 4px solid #2563eb; padding-left: 8px;">📋 目标日执行计划</span>
      </div>
      %s
    </div>

    <!-- 底部版权信息 Footer -->
    <div style="background-color: #f8fafc; padding: 14px 16px; text-align: center; border-top: 1px solid #e2e8f0; font-size: 12px; color: #64748b; line-height: 1.5;">
      <div>本文档由 <strong>Chronos AI</strong> 自动生成 · 祝今日高效专注，通关愉快 ✨</div>
    </div>
  </div>
</body>
</html>`, date, cfSectionHTML, bodyHTML.String())
}

// renderCFSection 渲染 Codeforces 每日特训卡片模块（流式自适应、防溢出）
func renderCFSection(problems []codeforces.ProblemDetail) string {
	if len(problems) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(`<div class="section-box" style="padding: 16px 16px 8px 16px; background-color: #f8fafc; border-bottom: 1px solid #e2e8f0; box-sizing: border-box;">` + "\n")
	sb.WriteString(`  <div style="margin-bottom: 12px;">` + "\n")
	sb.WriteString(`    <span style="font-size: 15px; font-weight: 700; color: #0f172a;">⚡ Codeforces 每日特训</span>` + "\n")
	sb.WriteString(`    <span style="margin-left: 6px; font-size: 11px; font-weight: 600; color: #2563eb; background-color: #dbeafe; padding: 2px 7px; border-radius: 9999px;">精选练习</span>` + "\n")
	sb.WriteString(`  </div>` + "\n")

	for i, p := range problems {
		ratingBg := "#dcfce7"
		ratingColor := "#166534"
		if p.Item.Rating >= 1200 && p.Item.Rating < 1400 {
			ratingBg = "#cffafe"
			ratingColor = "#0e7490"
		} else if p.Item.Rating >= 1400 && p.Item.Rating < 1600 {
			ratingBg = "#dbeafe"
			ratingColor = "#1d4ed8"
		} else if p.Item.Rating >= 1600 {
			ratingBg = "#f3e8ff"
			ratingColor = "#7e22ce"
		}

		sb.WriteString(fmt.Sprintf(`  <!-- Problem Item %d -->
  <div style="margin-bottom: 14px; background-color: #ffffff; border: 1px solid #e2e8f0; border-left: 4px solid #2563eb; border-radius: 8px; padding: 12px 14px; box-sizing: border-box;">
    <div style="margin-bottom: 8px;">
      <div style="margin-bottom: 6px;">
        <span style="display: inline-block; background-color: #0f172a; color: #ffffff; font-family: Consolas, Monaco, monospace; font-size: 12px; font-weight: 700; padding: 2px 6px; border-radius: 4px; margin-right: 4px;">%s</span>
        <strong style="color: #0f172a; font-size: 14.5px; vertical-align: middle;">%s</strong>
        <span style="display: inline-block; background-color: %s; color: %s; font-size: 11px; font-weight: 700; padding: 2px 6px; border-radius: 9999px; margin-left: 4px;">Rating: %d</span>
      </div>
      <div>
        <a href="%s" target="_blank" style="display: inline-block; background-color: #2563eb; color: #ffffff; text-decoration: none; font-size: 12px; font-weight: 600; padding: 3px 8px; border-radius: 5px;">🔗 前往 Codeforces 原题</a>
      </div>
    </div>
    <div style="font-size: 13px; color: #334155; background-color: #f8fafc; border: 1px solid #f1f5f9; padding: 10px 12px; border-radius: 6px; line-height: 1.6; word-break: break-word;">
      %s
    </div>
  </div>`+"\n", i+1, p.Item.Code, renderInline(p.Item.Name), ratingBg, ratingColor, p.Item.Rating, p.Item.URL, renderTranslationMarkdown(p.Translation)))
	}

	sb.WriteString("</div>\n")
	return sb.String()
}

// renderTranslationMarkdown 将题目翻译 Markdown 完整规范地渲染为带内联样式的 HTML
func renderTranslationMarkdown(trans string) string {
	if strings.TrimSpace(trans) == "" {
		return "<span style='color: #94a3b8;'>暂无中文解析</span>"
	}
	lines := strings.Split(trans, "\n")
	var sb strings.Builder

	inList := false

	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed == "" {
			if inList {
				sb.WriteString("</ul>\n")
				inList = false
			}
			continue
		}

		// 标题：### 或 ##
		if strings.HasPrefix(trimmed, "###") || strings.HasPrefix(trimmed, "##") {
			if inList {
				sb.WriteString("</ul>\n")
				inList = false
			}
			head := strings.TrimLeft(trimmed, "# ")
			sb.WriteString(fmt.Sprintf(`<div style="font-weight: 700; color: #1e293b; margin: 10px 0 4px 0; font-size: 13px; border-left: 3px solid #2563eb; padding-left: 6px;">%s</div>`+"\n", renderInline(head)))
			continue
		}

		// 列表项：- 或 *
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			if !inList {
				sb.WriteString(`<ul style="margin: 4px 0; padding-left: 18px; color: #334155;">` + "\n")
				inList = true
			}
			item := strings.TrimPrefix(trimmed, "- ")
			item = strings.TrimPrefix(item, "* ")
			sb.WriteString(fmt.Sprintf(`<li style="margin: 3px 0; line-height: 1.5; font-size: 13px; word-break: break-word;">%s</li>`+"\n", renderInline(item)))
			continue
		}

		if inList {
			sb.WriteString("</ul>\n")
			inList = false
		}

		// 普通段落
		sb.WriteString(fmt.Sprintf(`<div style="margin: 4px 0; line-height: 1.55; font-size: 13px; color: #475569; word-break: break-word;">%s</div>`+"\n", renderInline(trimmed)))
	}

	if inList {
		sb.WriteString("</ul>\n")
	}

	return sb.String()
}

// isTableRow 判定是否为 Markdown 表格行
func isTableRow(line string) bool {
	return strings.Contains(line, "|") && (strings.HasPrefix(line, "|") || strings.Count(line, "|") >= 2)
}

// splitTableRow 将表格行按 | 切分为单元格
func splitTableRow(line string) []string {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "|") {
		trimmed = trimmed[1:]
	}
	if strings.HasSuffix(trimmed, "|") {
		trimmed = trimmed[:len(trimmed)-1]
	}
	raw := strings.Split(trimmed, "|")
	cells := make([]string, 0, len(raw))
	for _, c := range raw {
		cells = append(cells, strings.TrimSpace(c))
	}
	return cells
}

// isTableSeparatorRow 判定是否为表格的对齐/分割行 (如 | :--- | :--- |)
func isTableSeparatorRow(cells []string) bool {
	if len(cells) == 0 {
		return false
	}
	for _, c := range cells {
		if !reTableSepCell.MatchString(c) {
			return false
		}
	}
	return true
}

// renderInline 处理 Markdown 行内格式：链接、粗体、斜体、删除线、行内代码及数学符号
func renderInline(text string) string {
	escaped := html.EscapeString(text)

	// 1. Markdown 链接 [text](url)
	escaped = reLink.ReplaceAllString(escaped, `<a href="$2" target="_blank" style="color: #2563eb; text-decoration: none; font-weight: 500; border-bottom: 1px dashed #2563eb;">$1</a>`)

	// 2. 粗体 **bold**
	escaped = reBold.ReplaceAllString(escaped, `<strong style="color: #0f172a; font-weight: 700;">$1</strong>`)

	// 3. 斜体 *italic*
	escaped = reItalic.ReplaceAllString(escaped, `<em>$1</em>`)

	// 4. 删除线 ~~del~~
	escaped = reStrike.ReplaceAllString(escaped, `<del style="color: #94a3b8;">$1</del>`)

	// 5. 行内代码 `code`
	escaped = reInlineCode.ReplaceAllString(escaped, `<code style="background-color: #f1f5f9; color: #2563eb; padding: 2px 5px; border-radius: 4px; font-family: Consolas, Monaco, monospace; font-size: 12.5px;">$1</code>`)

	// 6. 数学公式 \( ... \) 与 $ ... $ 规范化转换
	escaped = reMathParen.ReplaceAllStringFunc(escaped, func(m string) string {
		sub := strings.TrimPrefix(m, `\(`)
		sub = strings.TrimSuffix(sub, `\)`)
		return renderMathSymbol(sub)
	})
	escaped = reMathDollar.ReplaceAllStringFunc(escaped, func(m string) string {
		sub := strings.TrimPrefix(m, `$`)
		sub = strings.TrimSuffix(sub, `$`)
		return renderMathSymbol(sub)
	})

	return escaped
}

// renderMathSymbol 格式化 LaTeX 数学符号为纯 HTML 可视化标签
func renderMathSymbol(mathExpr string) string {
	expr := strings.TrimSpace(mathExpr)
	expr = strings.ReplaceAll(expr, `\le`, "≤")
	expr = strings.ReplaceAll(expr, `\ge`, "≥")
	expr = strings.ReplaceAll(expr, `\ne`, "≠")
	expr = strings.ReplaceAll(expr, `\neq`, "≠")
	expr = strings.ReplaceAll(expr, `\times`, "×")
	expr = strings.ReplaceAll(expr, `\cdot`, "·")
	expr = strings.ReplaceAll(expr, `\dots`, "...")
	expr = strings.ReplaceAll(expr, `\ldots`, "...")
	expr = strings.ReplaceAll(expr, `\cdots`, "...")
	expr = strings.ReplaceAll(expr, `\approx`, "≈")
	expr = strings.ReplaceAll(expr, `\in`, "∈")
	expr = strings.ReplaceAll(expr, `\subset`, "⊂")
	expr = strings.ReplaceAll(expr, `\to`, "→")

	reText := regexp.MustCompile(`\\text\{([^}]+)\}`)
	expr = reText.ReplaceAllString(expr, "$1")

	return fmt.Sprintf(`<span style="font-family: 'Cambria Math', 'Times New Roman', Consolas, serif; font-size: 12.5px; background-color: #f1f5f9; color: #0f172a; padding: 1px 4px; border-radius: 3px; border: 1px solid #e2e8f0;">%s</span>`, expr)
}
