package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"chronos/internal/ai"
	"chronos/internal/codeforces"
)

var (
	BaseDir            string
	TasksDir           string
	CacheDir           string
	PromptsDir         string
	BaseFile           string
	PoolFile           string
	UnexpectedFile     string
	CFTrackerFile      string
	CFTranslationsFile string

	// 🚨 专门保护突发事件读写的互斥锁，防止后台 GC 与前台写入发生竞态碰撞
	unexpectedMutex sync.Mutex
)

// 初始化所有路径
func initPaths() {
	home, _ := os.UserHomeDir()
	BaseDir = filepath.Join(home, ".chronos")
	TasksDir = filepath.Join(BaseDir, "tasks")
	CacheDir = filepath.Join(BaseDir, "cache") // 所有的计划文件都会存在这里
	PromptsDir = filepath.Join(BaseDir, "prompts")

	BaseFile = filepath.Join(TasksDir, "base.json")
	PoolFile = filepath.Join(TasksDir, "pool.json")
	UnexpectedFile = filepath.Join(TasksDir, "unexpected.json")
	CFTrackerFile = filepath.Join(TasksDir, "cf_tracker.json")
	CFTranslationsFile = filepath.Join(TasksDir, "cf_translations.json")
}

// 结构体定义
type DayState struct {
	Date        string                     `json:"date"` // 目标日期，例如 "2026-08-26"
	CurrentPlan string                     `json:"current_plan"`
	Messages    []ai.Message               `json:"messages"`
	CFProblems  []codeforces.ProblemDetail `json:"cf_problems,omitempty"`
}

type UnexpectedEvent struct {
	Date    string `json:"date"`
	Content string `json:"content"`
}

// InitAndGC 执行目录初始化与并发清理
func InitAndGC() {
	initPaths()
	// 1. 创建基础架构目录 (目录需要 0755 权限才能进入)
	_ = os.MkdirAll(TasksDir, 0755)
	_ = os.MkdirAll(CacheDir, 0755)
	_ = os.MkdirAll(PromptsDir, 0755)

	// 2. 初始化空文件 (不存在则创建，采用安全的 0600 权限)
	createIfNotExist(BaseFile, "{}")
	createIfNotExist(PoolFile, `{"tasks": []}`)
	createIfNotExist(UnexpectedFile, `[]`)
	createIfNotExist(CFTrackerFile, "{\n  \"next_index\": 0,\n  \"daily_assigned\": {}\n}")
	createIfNotExist(CFTranslationsFile, `{}`)

	// 3. 初始化提示词文件
	createIfNotExist(filepath.Join(PromptsDir, "gentle.md"), "你是高效温和的日程规划师 Chronos，请生成条理清晰的计划。")
	createIfNotExist(filepath.Join(PromptsDir, "anime.md"), "你是元气二次元搭子 Chronos，请称呼用户为前辈，用撒娇可爱的语气排计划！")

	// 4. 后台并发 GC (垃圾回收)
	go func() {
		todayStr := time.Now().Format("2006-01-02")
		
		// 🚨 清理突发事件必须加锁：保留今天以及未来的突发事件，删除昨天的
		unexpectedMutex.Lock()
		var events []UnexpectedEvent
		if err := loadJSON(UnexpectedFile, &events); err == nil {
			var validEvents []UnexpectedEvent
			for _, e := range events {
				if e.Date >= todayStr {
					validEvents = append(validEvents, e)
				}
			}
			_ = saveJSON(UnexpectedFile, validEvents)
		}
		unexpectedMutex.Unlock()

		// 🚨 清理 CacheDir 中 7 天前的过期计划文件
		cutoff := time.Now().Add(-7 * 24 * time.Hour)
		files, _ := os.ReadDir(CacheDir)
		for _, f := range files {
			info, err := f.Info()
			if err == nil && info.ModTime().Before(cutoff) {
				_ = os.Remove(filepath.Join(CacheDir, f.Name()))
			}
		}
	}()
}

// ================= API 封装 (全面支持动态日期) =================

// SavePlan 根据 Date 字段动态保存文件 (如 2026-08-26.json)
func SavePlan(state *DayState) error {
	path := filepath.Join(CacheDir, state.Date+".json")
	return saveJSON(path, state)
}

// LoadPlan 根据日期字符串动态加载文件
func LoadPlan(dateStr string) (*DayState, error) {
	path := filepath.Join(CacheDir, dateStr+".json")
	var s DayState
	err := loadJSON(path, &s)
	return &s, err
}

// AddUnexpected 接收目标日期，将突发事件精准绑定到对应天
func AddUnexpected(dateStr, content string) error {
	unexpectedMutex.Lock()
	defer unexpectedMutex.Unlock()

	var events []UnexpectedEvent
	_ = loadJSON(UnexpectedFile, &events)
	events = append(events, UnexpectedEvent{
		Date:    dateStr,
		Content: content,
	})
	return saveJSON(UnexpectedFile, events)
}

// 读取基础排程供 AI 参考
func LoadContextStrings() (baseStr, poolStr, unexpStr string) {
	b, _ := os.ReadFile(BaseFile)
	p, _ := os.ReadFile(PoolFile)
	u, _ := os.ReadFile(UnexpectedFile)
	return string(b), string(p), string(u)
}

// ================= 底层安全辅助函数 =================

func createIfNotExist(path, defaultContent string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 🚨 修改权限为 0600 (仅本人可读写)，防止在 VPS 被盗取数据
		_ = os.WriteFile(path, []byte(defaultContent), 0600)
	}
}

// saveJSON 升级为原子写入（Atomic Write）
func saveJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err // 绝不静默吞掉序列化错误
	}

	// 1. 先写到临时文件 (.tmp)
	tempFile := path + ".tmp"
	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		return err
	}

	// 2. 原子的重命名覆盖，保证断电或报错时原文件绝不损坏
	return os.Rename(tempFile, path)
}

func loadJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// LoadBaseCFProblems 从 base.json 中读取 codeforces_problems 题单
// 若尚未配置，自动将 60 道推荐题单注入 base.json 并原子存盘，绝不破坏已有课表等内容
func LoadBaseCFProblems() []codeforces.ProblemItem {
	var rawMap map[string]any
	if err := loadJSON(BaseFile, &rawMap); err != nil || rawMap == nil {
		rawMap = make(map[string]any)
	}

	if cfRaw, exists := rawMap["codeforces_problems"]; exists {
		data, err := json.Marshal(cfRaw)
		if err == nil {
			var items []codeforces.ProblemItem
			if err := json.Unmarshal(data, &items); err == nil && len(items) > 0 {
				return items
			}
		}
	}

	// base.json 尚未配置题单，自动注入默认 60 题并持久化保存
	defaultItems := codeforces.GetDefaultProblemItems()
	rawMap["codeforces_problems"] = defaultItems
	_ = saveJSON(BaseFile, rawMap)
	return defaultItems
}

// SaveBaseCFProblems 保存或更新 base.json 中的 codeforces_problems 题单
func SaveBaseCFProblems(problems []codeforces.ProblemItem) error {
	var rawMap map[string]any
	if err := loadJSON(BaseFile, &rawMap); err != nil || rawMap == nil {
		rawMap = make(map[string]any)
	}
	rawMap["codeforces_problems"] = problems
	return saveJSON(BaseFile, rawMap)
}