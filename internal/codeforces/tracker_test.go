package codeforces

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMarkdownTable(t *testing.T) {
	items := GetDefaultProblemItems()
	if len(items) != 60 {
		t.Fatalf("预期 60 道题，实际解析出: %d", len(items))
	}

	// 验证第 1 题
	if items[0].Code != "4A" || items[0].Name != "Watermelon" || items[0].Rating != 800 {
		t.Errorf("第 1 题解析异常: %+v", items[0])
	}

	// 验证第 24 题 (266B Queue at the School)
	if items[23].Code != "266B" || items[23].Index != "B" || items[23].ContestID != 266 {
		t.Errorf("第 24 题 (266B) 解析异常: %+v", items[23])
	}

	// 验证第 60 题 (1472B Fair Division)
	if items[59].Code != "1472B" || items[59].Index != "B" || items[59].Rating != 800 {
		t.Errorf("第 60 题 (1472B) 解析异常: %+v", items[59])
	}
}

func TestTrackerIdempotencyAndProgression(t *testing.T) {
	tempDir := t.TempDir()
	trackerFile := filepath.Join(tempDir, "cf_tracker.json")
	tracker := NewTracker(trackerFile)

	items := GetDefaultProblemItems()

	// 1. 模拟第一天分配 (2026-09-07)
	day1, err := tracker.GetOrAssignDailyProblems("2026-09-07", 2, items)
	if err != nil {
		t.Fatalf("第一天分配失败: %v", err)
	}
	if len(day1) != 2 {
		t.Fatalf("预期分配 2 道题，实际: %d", len(day1))
	}
	if day1[0].Code != "4A" || day1[1].Code != "71A" {
		t.Errorf("第一天题目不符合预期: %s, %s", day1[0].Code, day1[1].Code)
	}

	// 2. 同一天再次调用（模拟微调日程或重新运行），验证幂等性
	day1Again, err := tracker.GetOrAssignDailyProblems("2026-09-07", 2, items)
	if err != nil {
		t.Fatalf("同一天重复获取失败: %v", err)
	}
	if len(day1Again) != 2 || day1Again[0].Code != "4A" || day1Again[1].Code != "71A" {
		t.Errorf("幂等性校验失败: %+v", day1Again)
	}

	// 验证游标依然停留在 2，没有在同日错误递增
	state, _ := tracker.Load()
	if state.NextIndex != 2 {
		t.Errorf("游标状态异常，预期为 2，实际为: %d", state.NextIndex)
	}

	// 3. 模拟第二天分配 (2026-09-08)，验证顺延推进
	day2, err := tracker.GetOrAssignDailyProblems("2026-09-08", 2, items)
	if err != nil {
		t.Fatalf("第二天分配失败: %v", err)
	}
	if len(day2) != 2 || day2[0].Code != "231A" || day2[1].Code != "282A" {
		t.Errorf("第二天题目未正确顺延: %+v", day2)
	}

	state2, _ := tracker.Load()
	if state2.NextIndex != 4 {
		t.Errorf("第二天游标预期推进到 4，实际为: %d", state2.NextIndex)
	}

	// 4. 测试用户临时在当天调大题量：从 2 改为 3
	day2Expanded, err := tracker.GetOrAssignDailyProblems("2026-09-08", 3, items)
	if err != nil {
		t.Fatalf("调大题量失败: %v", err)
	}
	if len(day2Expanded) != 3 {
		t.Fatalf("调大题量后预期返回 3 道题，实际为: %d", len(day2Expanded))
	}
	if day2Expanded[0].Code != "231A" || day2Expanded[1].Code != "282A" || day2Expanded[2].Code != "158A" {
		t.Errorf("调大题量后补齐题目不正确: %+v", day2Expanded)
	}

	state3, _ := tracker.Load()
	if state3.NextIndex != 5 {
		t.Errorf("补齐后游标预期推进到 5，实际为: %d", state3.NextIndex)
	}

	_ = os.RemoveAll(tempDir)
}
