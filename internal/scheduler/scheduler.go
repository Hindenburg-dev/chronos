package scheduler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"chronos/internal/ai"
	"chronos/internal/codeforces"
	"chronos/internal/storage"
)

type Scheduler struct {
	aiClient *ai.Client
}  

func New(client *ai.Client) *Scheduler {
	return &Scheduler{aiClient: client}
}

// 注意：这里新增了 targetDate 与 cfProblems 参数
func (s *Scheduler) Plan(persona, timeRange string, tasks []string, targetDate time.Time, cfProblems []codeforces.ProblemDetail) (*storage.DayState, error) {
	// 1. 读取系统 Prompt
	promptPath := filepath.Join(storage.PromptsDir, persona+".md")
	promptData, err := os.ReadFile(promptPath)
	systemPrompt := "你是高效时间规划助手 Chronos。"
	if err == nil {
		systemPrompt = string(promptData)
	}
  
	// 2. 读取上下文库
	baseStr, poolStr, unexpStr := storage.LoadContextStrings()
	
	now := time.Now()
	weekdays := []string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}
	targetDateStr := targetDate.Format("2006-01-02")
	targetWeekday := weekdays[targetDate.Weekday()]

	// 3. 构建超级上下文
	var userPrompt strings.Builder
	userPrompt.WriteString(fmt.Sprintf("【现实当前时间】: %s\n", now.Format("2006-01-02 15:04")))
	userPrompt.WriteString(fmt.Sprintf("【🎯 规划目标日期】: %s %s\n", targetDateStr, targetWeekday))

	// 智能推导时间范围：自动区分是排“今天余下时间”还是“明天一整天”
	if strings.TrimSpace(timeRange) == "" {
		if targetDateStr == now.Format("2006-01-02") {
			userPrompt.WriteString(fmt.Sprintf("【规划时间范围】: 从当前现实时刻（%s）开始，直到今晚就寝时间。\n", now.Format("15:04")))
		} else {
			userPrompt.WriteString("【规划时间范围】: 目标日全天（请根据基础习惯和课表自动规划作息）。\n")
		}
	} else {
		userPrompt.WriteString(fmt.Sprintf("【规划时间范围】: %s\n", timeRange))
	}

	userPrompt.WriteString("\n【固定课表/习惯限制 (绝对不可冲突)】:\n")
	userPrompt.WriteString(baseStr)
	
	userPrompt.WriteString("\n【突发事件 (高优先级排入)】:\n")
	userPrompt.WriteString(unexpStr)
	
	userPrompt.WriteString("\n【中短期任务池参考 (如有空闲可抽取)】:\n")
	userPrompt.WriteString(poolStr)

	// Codeforces 题目注入
	if len(cfProblems) > 0 {
		userPrompt.WriteString(fmt.Sprintf("\n\n【⚡ 今日 Codeforces 每日算法特训（共 %d 道指定题目）】:\n", len(cfProblems)))
		userPrompt.WriteString(fmt.Sprintf("（系统已严格指定今日需完成以下 %d 道题目，请务必在日程中为每道题目规划出专注刷题时段。安排的题目必须且只能是以下这 %d 道，严禁自行杜撰或增加其他题目编号）：\n", len(cfProblems), len(cfProblems)))
		for i, p := range cfProblems {
			userPrompt.WriteString(fmt.Sprintf("%d. [%s] %s (难度: %d)\n   原题链接: %s\n",
				i+1, p.Item.Code, p.Item.Name, p.Item.Rating, p.Item.URL))
			if p.Translation != "" {
				userPrompt.WriteString(fmt.Sprintf("   题意点拨: %s\n", cleanBrief(p.Translation)))
			}
		}
	}

	userPrompt.WriteString("\n\n【用户指定待办任务】:\n")
	for i, task := range tasks {
		userPrompt.WriteString(fmt.Sprintf("%d. %s\n", i+1, task))
	}
	userPrompt.WriteString("\n请基于上述规则、固定课表与可用时间，为我生成该目标日的结构化执行时间表。")

	messages := []ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt.String()},
	}

	reply, err := s.aiClient.Chat(messages)
	if err != nil {
		return nil, err
	}

	messages = append(messages, ai.Message{Role: "assistant", Content: reply})

	state := &storage.DayState{
		Date:        targetDateStr, // 这里将生成的计划精准绑定到目标日期
		CurrentPlan: reply,
		Messages:    messages,
		CFProblems:  cfProblems,
	}
	return state, nil
}

func cleanBrief(trans string) string {
	lines := strings.Split(trans, "\n")
	var brief []string
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t != "" && !strings.HasPrefix(t, "#") {
			brief = append(brief, t)
			if len(brief) >= 3 {
				break
			}
		}
	}
	return strings.Join(brief, " ")
}

func (s *Scheduler) Adjust(state *storage.DayState, tweak string) error {
	msg := fmt.Sprintf("【%s 提出微调】: %s。请重新输出完整时间表。", time.Now().Format("15:04"), tweak)
	state.Messages = append(state.Messages, ai.Message{Role: "user", Content: msg})
	
	reply, err := s.aiClient.Chat(state.Messages)
	if err != nil {
		return err
	}
	
	state.Messages = append(state.Messages, ai.Message{Role: "assistant", Content: reply})
	state.CurrentPlan = reply
	return nil
}