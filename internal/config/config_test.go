package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAMLConfigMigrationAndParsing(t *testing.T) {
	tempDir := t.TempDir()
	jsonFile := filepath.Join(tempDir, "config.json")
	yamlFile := filepath.Join(tempDir, "config.yaml")

	// 1. 模拟旧版 JSON 配置
	oldData := map[string]any{
		"platform": "deepseek",
		"api_key":   "sk-test-key-123456",
		"model":    "deepseek-chat",
		"email": map[string]any{
			"enabled":   true,
			"smtp_host": "smtp.qq.com",
			"smtp_port": 465,
			"from":      "test@qq.com",
			"password":  "authcode",
			"to":        "user@qq.com",
		},
	}
	jb, _ := json.MarshalIndent(oldData, "", "  ")
	_ = os.WriteFile(jsonFile, jb, 0600)

	// 2. 执行平滑转换逻辑
	jsonData, err := os.ReadFile(jsonFile)
	if err != nil {
		t.Fatalf("读取 JSON 失败: %v", err)
	}
	var migratedCfg Config
	if err := json.Unmarshal(jsonData, &migratedCfg); err != nil {
		t.Fatalf("反序列化 JSON 失败: %v", err)
	}

	migratedCfg.Codeforces = CodeforcesConfig{
		Enabled:    true,
		DailyCount: 2,
	}

	yamlBytes, err := yaml.Marshal(&migratedCfg)
	if err != nil {
		t.Fatalf("序列化 YAML 失败: %v", err)
	}
	_ = os.WriteFile(yamlFile, yamlBytes, 0600)

	// 3. 读取并验证 YAML
	readYaml, err := os.ReadFile(yamlFile)
	if err != nil {
		t.Fatalf("读取 YAML 失败: %v", err)
	}

	var loadedCfg Config
	if err := yaml.Unmarshal(readYaml, &loadedCfg); err != nil {
		t.Fatalf("反序列化 YAML 失败: %v", err)
	}

	if loadedCfg.APIKey != "sk-test-key-123456" {
		t.Errorf("APIKey 迁移不匹配: %s", loadedCfg.APIKey)
	}
	if !loadedCfg.Codeforces.Enabled || loadedCfg.Codeforces.DailyCount != 2 {
		t.Errorf("Codeforces 默认配置异常: %+v", loadedCfg.Codeforces)
	}
	if !loadedCfg.Email.Enabled || loadedCfg.Email.From != "test@qq.com" {
		t.Errorf("Email 配置迁移异常: %+v", loadedCfg.Email)
	}
}
