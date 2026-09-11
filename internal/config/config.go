package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type CodeforcesConfig struct {
	Enabled    bool `yaml:"enabled" json:"enabled"`         // 是否开启每日抓题功能 (bool)
	DailyCount int  `yaml:"daily_count" json:"daily_count"` // 每日抓题数量 (int, 默认 2)
}

type EmailConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`     // 是否启用邮件发送
	SMTPHost string `yaml:"smtp_host" json:"smtp_host"` // 比如 "smtp.qq.com" 或 "smtp.163.com"
	SMTPPort int    `yaml:"smtp_port" json:"smtp_port"` // 比如 465 (SSL) 或 587 (STARTTLS)
	From     string `yaml:"from" json:"from"`          // 发件人邮箱 (如 "xxx@qq.com")
	Password string `yaml:"password" json:"password"`  // 邮箱授权码/密码
	To       string `yaml:"to" json:"to"`              // 收件人邮箱 (留空默认发给发件人自己)
}

type Config struct {
	Platform   string           `yaml:"platform" json:"platform"` // 比如 "deepseek"
	APIKey     string           `yaml:"api_key" json:"api_key"`   // 你的密钥
	Model      string           `yaml:"model" json:"model"`       // 比如 "deepseek-chat"
	Email      EmailConfig      `yaml:"email" json:"email"`
	Codeforces CodeforcesConfig `yaml:"codeforces" json:"codeforces"`
}

// GetHomeDir 获取 ~/.chronos 根目录
func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".chronos")
}

// LoadOrInit 读取或初始化 YAML 配置，支持从旧版 config.json 自动无缝迁移
func LoadOrInit() (*Config, error) {
	dir := GetHomeDir()
	_ = os.MkdirAll(dir, 0755)

	yamlFile := filepath.Join(dir, "config.yaml")
	jsonFile := filepath.Join(dir, "config.json")

	// 1. 如果 config.yaml 不存在，检查旧版 config.json 进行平滑迁移
	if _, err := os.Stat(yamlFile); os.IsNotExist(err) {
		if _, jsonErr := os.Stat(jsonFile); jsonErr == nil {
			// 旧版 config.json 存在，执行无损迁移
			jsonData, readErr := os.ReadFile(jsonFile)
			if readErr == nil {
				var oldCfg Config
				if unmarshalErr := json.Unmarshal(jsonData, &oldCfg); unmarshalErr == nil {
					// 注入 Codeforces 默认配置
					oldCfg.Codeforces = CodeforcesConfig{
						Enabled:    true,
						DailyCount: 2,
					}
					yamlBytes, marshalErr := yaml.Marshal(&oldCfg)
					if marshalErr == nil {
						_ = os.WriteFile(yamlFile, yamlBytes, 0600)
						fmt.Printf("💡 检测到旧版 JSON 配置，已平滑迁移为 YAML 配置: %s\n", yamlFile)
					}
				}
			}
		} else {
			// 两个都不存在，生成全新的默认 config.yaml
			defaultCfg := &Config{
				Platform: "deepseek",
				APIKey:   "YOUR_API_KEY_HERE",
				Model:    "deepseek-chat",
				Email: EmailConfig{
					Enabled:  false,
					SMTPHost: "smtp.qq.com",
					SMTPPort: 465,
					From:     "your_email@qq.com",
					Password: "YOUR_EMAIL_AUTH_CODE",
					To:       "your_email@qq.com",
				},
				Codeforces: CodeforcesConfig{
					Enabled:    true,
					DailyCount: 2,
				},
			}
			data, _ := yaml.Marshal(defaultCfg)
			_ = os.WriteFile(yamlFile, data, 0600)
			return nil, fmt.Errorf("首次运行，已在 %s 创建默认配置文件，请填写 API Key 后重试", yamlFile)
		}
	}

	// 2. 读取 config.yaml
	data, err := os.ReadFile(yamlFile)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析 YAML 配置文件失败 (%s): %w", yamlFile, err)
	}

	// 补全默认值
	if cfg.Codeforces.DailyCount <= 0 {
		cfg.Codeforces.DailyCount = 2
	}

	if cfg.APIKey == "YOUR_API_KEY_HERE" || cfg.APIKey == "" {
		return nil, fmt.Errorf("请先打开 %s 填写你的真实 API Key", yamlFile)
	}

	return &cfg, nil
}