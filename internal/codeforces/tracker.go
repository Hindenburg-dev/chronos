package codeforces

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Tracker struct {
	filePath string
	mu       sync.Mutex
}

func NewTracker(filePath string) *Tracker {
	return &Tracker{filePath: filePath}
}

// Load 读取追踪状态
func (t *Tracker) Load() (*TrackerState, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.loadUnlocked()
}

func (t *Tracker) loadUnlocked() (*TrackerState, error) {
	data, err := os.ReadFile(t.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &TrackerState{
				NextIndex:     0,
				DailyAssigned: make(map[string][]string),
			}, nil
		}
		return nil, err
	}

	var state TrackerState
	if err := json.Unmarshal(data, &state); err != nil {
		return &TrackerState{
			NextIndex:     0,
			DailyAssigned: make(map[string][]string),
		}, nil
	}
	if state.DailyAssigned == nil {
		state.DailyAssigned = make(map[string][]string)
	}
	return &state, nil
}

// Save 原子写入追踪状态
func (t *Tracker) Save(state *TrackerState) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.saveUnlocked(state)
}

func (t *Tracker) saveUnlocked(state *TrackerState) error {
	state.LastUpdated = time.Now().Format("2006-01-02 15:04:05")
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	tmpFile := t.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmpFile, t.filePath)
}

// GetOrAssignDailyProblems 获取或分配某日的 CF 题目
// 1. 同一天重复调用具有幂等性：若配置数量不变，直接返回当天已分配题目，游标不推进
// 2. 若用户修改了 daily_count（如从 2 改为 3），自动顺延补齐缺失题目，保证与配置完全一致
// 3. 新的一天调用：从 NextIndex 开始取出 count 道题，游标前移并持久化保存
func (t *Tracker) GetOrAssignDailyProblems(date string, count int, allProblems []ProblemItem) ([]ProblemItem, error) {
	if len(allProblems) == 0 {
		return nil, nil
	}
	if count <= 0 {
		count = 2
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	state, err := t.loadUnlocked()
	if err != nil {
		return nil, err
	}

	total := len(allProblems)
	if state.NextIndex >= total || state.NextIndex < 0 {
		state.NextIndex = 0
	}

	codeMap := make(map[string]ProblemItem, total)
	for _, p := range allProblems {
		codeMap[p.Code] = p
	}

	// 1. 检查当日是否已经分配过题目
	if assignedCodes, ok := state.DailyAssigned[date]; ok && len(assignedCodes) > 0 {
		if len(assignedCodes) == count {
			var result []ProblemItem
			for _, code := range assignedCodes {
				if item, exists := codeMap[code]; exists {
					result = append(result, item)
				}
			}
			if len(result) == count {
				return result, nil
			}
		} else if len(assignedCodes) > count {
			// 用户在配置中调小了 daily_count，返回前 count 道
			var result []ProblemItem
			for _, code := range assignedCodes[:count] {
				if item, exists := codeMap[code]; exists {
					result = append(result, item)
				}
			}
			return result, nil
		} else {
			// 用户在配置中调大了 daily_count（如从 2 改为 3），顺延补齐
			diff := count - len(assignedCodes)
			var result []ProblemItem
			for _, code := range assignedCodes {
				if item, exists := codeMap[code]; exists {
					result = append(result, item)
				}
			}

			for i := 0; i < diff; i++ {
				idx := (state.NextIndex + i) % total
				item := allProblems[idx]
				result = append(result, item)
				assignedCodes = append(assignedCodes, item.Code)
			}

			state.NextIndex = (state.NextIndex + diff) % total
			state.DailyAssigned[date] = assignedCodes
			_ = t.saveUnlocked(state)
			return result, nil
		}
	}

	// 2. 当日未分配：从 NextIndex 开始截取 count 道题
	var assignedItems []ProblemItem
	var assignedCodes []string

	for i := 0; i < count; i++ {
		idx := (state.NextIndex + i) % total
		item := allProblems[idx]
		assignedItems = append(assignedItems, item)
		assignedCodes = append(assignedCodes, item.Code)
	}

	// 游标前移
	state.NextIndex = (state.NextIndex + count) % total
	state.DailyAssigned[date] = assignedCodes

	if err := t.saveUnlocked(state); err != nil {
		return assignedItems, err
	}

	return assignedItems, nil
}
