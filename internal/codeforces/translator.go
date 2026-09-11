package codeforces

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"chronos/internal/ai"
)

type Translator struct {
	cachePath string
	crawler   *Crawler
	aiClient  *ai.Client
	mu        sync.Mutex
}

func NewTranslator(cachePath string, aiClient *ai.Client) *Translator {
	return &Translator{
		cachePath: cachePath,
		crawler:   NewCrawler(),
		aiClient:  aiClient,
	}
}

func (t *Translator) loadCacheUnlocked() (map[string]TranslationRecord, error) {
	data, err := os.ReadFile(t.cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]TranslationRecord), nil
		}
		return nil, err
	}

	var records map[string]TranslationRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return make(map[string]TranslationRecord), nil
	}
	return records, nil
}

func (t *Translator) saveCacheUnlocked(records map[string]TranslationRecord) error {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	tmpFile := t.cachePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmpFile, t.cachePath)
}

// GetOrTranslate 获取题目的中文翻译。
// 若本地已有持久化缓存，直接返回；若无，则抓取网页原题并调用 AI 翻译，随后存盘。
func (t *Translator) GetOrTranslate(item ProblemItem) (string, error) {
	t.mu.Lock()
	records, err := t.loadCacheUnlocked()
	if err != nil {
		records = make(map[string]TranslationRecord)
	}

	// 1. 检查本地持久化缓存 (命中则零 AI 调用，秒级返回)
	if record, ok := records[item.Code]; ok && record.Translation != "" {
		t.mu.Unlock()
		return record.Translation, nil
	}
	t.mu.Unlock()

	// 2. 未命中，抓取原题内容
	rawStatement, err := t.crawler.FetchProblemStatement(item.ContestID, item.Index)
	if err != nil {
		rawStatement = fmt.Sprintf("Problem: %s (%s), Rating: %d, URL: %s", item.Name, item.Code, item.Rating, item.URL)
	}

	// 3. 构建专业的 ACM-ICPC 算法教练提示词
	systemPrompt := `你是一名资深 ACM-ICPC / Codeforces 竞赛教练和算法专家。
请将给定的 Codeforces 题目英文原文准确、通俗、结构清晰地翻译为规范中文。

格式要求（纯 Markdown，请直接输出以下结构，不要任何多余寒暄）：
### 【题目大意】
（用精练通俗的中文提炼问题背景与核心目标，滤除无用故事冗余，突出数学与逻辑本质）

### 【输入输出与数据范围】
- 输入：说明各变量含义
- 输出：说明输出格式要求
- 限制与范围：准确保留所有变量范围与时空限制（如 $1 \le n \le 10^5$）

### 【💡 思维切入点】
（提供 1~2 句话的核心思考方向或解题启示，如贪心、双指针、前缀和、数学分析等，重在启发思维，切勿直接贴出整段代码）`

	userPrompt := fmt.Sprintf("【题目代号】: %s\n【题目名称】: %s\n【难度 Rating】: %d\n【题目直达链接】: %s\n\n【英文原题文本】:\n%s\n\n请按规范翻译该题目：",
		item.Code, item.Name, item.Rating, item.URL, rawStatement)

	messages := []ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	translation, err := t.aiClient.Chat(messages)
	if err != nil {
		return "", fmt.Errorf("调用 AI 翻译失败 (%s): %w", item.Code, err)
	}

	// 4. 立即持久化至本地缓存
	t.mu.Lock()
	defer t.mu.Unlock()

	currentRecords, _ := t.loadCacheUnlocked()
	currentRecords[item.Code] = TranslationRecord{
		Code:         item.Code,
		ContestID:    item.ContestID,
		Index:        item.Index,
		Name:         item.Name,
		Rating:       item.Rating,
		URL:          item.URL,
		Translation:  translation,
		TranslatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	_ = t.saveCacheUnlocked(currentRecords)

	return translation, nil
}
