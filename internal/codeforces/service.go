package codeforces

import (
	"fmt"

	"chronos/internal/ai"
)

type Service struct {
	tracker    *Tracker
	translator *Translator
}

func NewService(trackerPath, transCachePath string, aiClient *ai.Client) *Service {
	return &Service{
		tracker:    NewTracker(trackerPath),
		translator: NewTranslator(transCachePath, aiClient),
	}
}

// GetDailyProblems 获取或分配某日的题目，并获取其独立中文翻译（带持久化缓存）
func (s *Service) GetDailyProblems(date string, count int, problems []ProblemItem) ([]ProblemDetail, error) {
	if len(problems) == 0 {
		return nil, nil
	}

	assignedItems, err := s.tracker.GetOrAssignDailyProblems(date, count, problems)
	if err != nil {
		return nil, fmt.Errorf("分配每日题目失败: %w", err)
	}

	var details []ProblemDetail
	for _, item := range assignedItems {
		trans, err := s.translator.GetOrTranslate(item)
		if err != nil {
			// 若某道题翻译出现网络或临时错误，不中断整个流程，使用降级提示
			trans = fmt.Sprintf("⚠️ 暂未获取到翻译内容（可直接前往原题查看: %s）\n错误详情: %v", item.URL, err)
		}
		details = append(details, ProblemDetail{
			Item:        item,
			Translation: trans,
		})
	}

	return details, nil
}
