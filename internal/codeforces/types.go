package codeforces

// ProblemItem 对应 base.json 中题单中的单个题目项
type ProblemItem struct {
	ID        int    `json:"id"`         // 序号 (1, 2, 3...)
	Code      string `json:"code"`       // 题目代号 (如 "4A", "266B")
	ContestID int    `json:"contest_id"` // 比赛ID (如 4, 266)
	Index     string `json:"index"`      // 题号 (如 "A", "B")
	Name      string `json:"name"`       // 英文名称 (如 "Watermelon")
	Rating    int    `json:"rating"`     // 难度分 (如 800)
	URL       string `json:"url"`        // 原题直达链接
}

// ProblemDetail 包含题目元信息和中文翻译的综合详情
type ProblemDetail struct {
	Item        ProblemItem `json:"item"`
	Translation string      `json:"translation"` // 独立生成的中文题意与思维点拨
}

// TrackerState 记录每日刷题进度游标与每日分配记录
type TrackerState struct {
	NextIndex     int                 `json:"next_index"`     // 下一次从第几个题目开始取 (0-indexed)
	DailyAssigned map[string][]string `json:"daily_assigned"` // 日期 -> 题号列表，例如 "2026-09-07": ["4A", "71A"]
	LastUpdated   string              `json:"last_updated"`
}

// TranslationRecord 独立题目翻译的持久化结构
type TranslationRecord struct {
	Code        string `json:"code"`        // "4A"
	ContestID   int    `json:"contest_id"`
	Index       string `json:"index"`
	Name        string `json:"name"`
	Rating      int    `json:"rating"`
	URL         string `json:"url"`
	Translation string `json:"translation"` // 翻译文本 (Markdown 格式)
	TranslatedAt string `json:"translated_at"`
}
