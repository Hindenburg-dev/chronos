package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadBaseCFProblems(t *testing.T) {
	tempDir := t.TempDir()
	oldBaseFile := BaseFile
	BaseFile = filepath.Join(tempDir, "base.json")
	defer func() { BaseFile = oldBaseFile }()

	// 模拟写入带课表的基础数据
	baseContent := `{"metadata": {"term": "2026-2027"}, "weekly_schedule": []}`
	_ = os.WriteFile(BaseFile, []byte(baseContent), 0600)

	// 1. 首次加载，应当自动注入 60 题
	items := LoadBaseCFProblems()
	if len(items) != 60 {
		t.Fatalf("预期 60 题，实际: %d", len(items))
	}
	if items[0].Code != "4A" || items[59].Code != "1472B" {
		t.Errorf("题目数据不匹配: 第一题 %s, 最后一题 %s", items[0].Code, items[59].Code)
	}

	// 2. 验证原有课表数据 metadata 和 weekly_schedule 是否依然完好无损
	var checkMap map[string]any
	_ = loadJSON(BaseFile, &checkMap)
	if _, ok := checkMap["metadata"]; !ok {
		t.Error("metadata 丢失")
	}
	if _, ok := checkMap["weekly_schedule"]; !ok {
		t.Error("weekly_schedule 丢失")
	}

	// 3. 第二次加载，验证直接从 base.json 读取
	items2 := LoadBaseCFProblems()
	if len(items2) != 60 {
		t.Fatalf("第二次读取预期 60 题，实际: %d", len(items2))
	}
}
