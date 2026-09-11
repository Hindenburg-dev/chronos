package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"chronos/internal/ai"
	"chronos/internal/codeforces"
	"chronos/internal/config"
	"chronos/internal/email"
	"chronos/internal/scheduler"
	"chronos/internal/storage"
)

func main() {
	// 1. 初始化配置与存储层
	cfg, err := config.LoadOrInit()
	if err != nil {
		log.Fatal(err)
	}
	storage.InitAndGC()

	aiClient := ai.NewClient(cfg.Platform, cfg.APIKey, cfg.Model)
	sch := scheduler.New(aiClient)
	scanner := bufio.NewScanner(os.Stdin)
	var currentState *storage.DayState

	// 初始化 Codeforces 题目服务
	var cfService *codeforces.Service
	var allCFProblems []codeforces.ProblemItem
	if cfg.Codeforces.Enabled {
		cfService = codeforces.NewService(storage.CFTrackerFile, storage.CFTranslationsFile, aiClient)
		allCFProblems = storage.LoadBaseCFProblems()
	}

	fmt.Println("========================================")
	fmt.Println("       ⏳ Chronos 智能日程规划引擎      ")
	fmt.Println("========================================")

	// 2. 🌟 智能推断目标日期
	now := time.Now()
	targetDate := now
	isLate := now.Hour() >= 21 // 晚上 9 点以后，默认规划明天

	fmt.Println("👉 请选择要规划的日期:")
	if isLate {
		fmt.Printf("   [1] 今天 (%s)\n", now.Format("01-02"))
		fmt.Printf("   [2] 明天 (%s) [默认]\n", now.AddDate(0, 0, 1).Format("01-02"))
		fmt.Print("请选择 (输入 1 或 2，回车默认选 2): ")
		scanner.Scan()
		if strings.TrimSpace(scanner.Text()) != "1" {
			targetDate = now.AddDate(0, 0, 1)
		}
	} else {
		fmt.Printf("   [1] 今天 (%s) [默认]\n", now.Format("01-02"))
		fmt.Printf("   [2] 明天 (%s)\n", now.AddDate(0, 0, 1).Format("01-02"))
		fmt.Print("请选择 (输入 1 或 2，回车默认选 1): ")
		scanner.Scan()
		if strings.TrimSpace(scanner.Text()) == "2" {
			targetDate = now.AddDate(0, 0, 1)
		}
	}

	targetDateStr := targetDate.Format("2006-01-02")

	// 3. 准备 Codeforces 题目 (根据最新的 daily_count 动态自适应同步)
	var todaysCFProblems []codeforces.ProblemDetail
	if cfg.Codeforces.Enabled && cfService != nil {
		details, cfErr := cfService.GetDailyProblems(targetDateStr, cfg.Codeforces.DailyCount, allCFProblems)
		if cfErr != nil {
			fmt.Printf("⚠️ 提取 Codeforces 题目时遇到警告: %v\n", cfErr)
		} else {
			todaysCFProblems = details
		}
	}

	// 检查目标日期的缓存状态
	existingState, err := storage.LoadPlan(targetDateStr)
	hasCache := err == nil && existingState.Date == targetDateStr

	// ==========================================
	// 第一阶段：生成或读取【初版底稿】
	// ==========================================
	if hasCache {
		fmt.Printf("💡 检测到 [%s] 已有执行计划。\n", targetDateStr)
		// 动态同步最新的 CF 题目列表与题量
		if cfg.Codeforces.Enabled && len(todaysCFProblems) > 0 {
			existingState.CFProblems = todaysCFProblems
		}

		fmt.Print("👉 是否重新生成初版排程？(直接 [回车] 读取已有计划，输入 r 重新生成): ")
		scanner.Scan()
		choice := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if choice == "r" || choice == "new" {
			hasCache = false
		} else {
			currentState = existingState
			if len(currentState.CFProblems) > 0 {
				fmt.Printf("\n⚡ 今日 Codeforces 特训任务已同步 (%d 题):\n", len(currentState.CFProblems))
				for i, p := range currentState.CFProblems {
					fmt.Printf("   [%d] %s - %s (Rating: %d)\n       原题: %s\n", i+1, p.Item.Code, p.Item.Name, p.Item.Rating, p.Item.URL)
				}
			}
			fmt.Println("\n================ 📋 当前计划概要 ================")
			fmt.Println(currentState.CurrentPlan)
		}
	}

	if !hasCache {
		if cfg.Codeforces.Enabled && len(todaysCFProblems) > 0 {
			fmt.Printf("\n✅ 今日 Codeforces 特训题目就绪 (%d 题):\n", len(todaysCFProblems))
			for i, p := range todaysCFProblems {
				fmt.Printf("   [%d] %s - %s (Rating: %d)\n       原题: %s\n", i+1, p.Item.Code, p.Item.Name, p.Item.Rating, p.Item.URL)
			}
		}

		currentState = createInitialPlan(sch, scanner, targetDate, todaysCFProblems)
	}

	// ==========================================
	// 第二阶段：循环录入【突发事件】
	// ==========================================
	fmt.Println("\n================ 🚨 突发事件录入 ================")
	fmt.Println("有什么不在计划内的突发/强制事件吗？(可连续输入多个，直接 [回车] 跳过或结束)")
	
	var newEvents []string
	for {
		fmt.Printf("突发事件 %d: ", len(newEvents)+1)
		if !scanner.Scan() {
			break
		}
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			break // 敲回车，结束录入
		}
		newEvents = append(newEvents, text)
		
		// 🚨 修复：捕获突发事件保存错误（注意这里传入了目标日期）
		if err := storage.AddUnexpected(targetDateStr, text); err != nil {
			fmt.Printf("⚠️ 记录突发事件失败 (可能无法持久化): %v\n", err)
		}
	}

	// 如果有新录入的突发事件，让 AI 强行重排
	if len(newEvents) > 0 {
		fmt.Println("\n⏳ 正在将突发事件强行插入时间轴，请稍候...")
		
		// 构建强插 Prompt
		prompt := "我新增了以下突发事件，它们是最高优先级（必须严格执行），请帮我重新规划时间表，被挤占的任务可以顺延或退回：\n"
		for i, ev := range newEvents {
			prompt += fmt.Sprintf("%d. %s\n", i+1, ev)
		}

		if err := sch.Adjust(currentState, prompt); err != nil {
			fmt.Println("❌ 插入突发事件失败:", err)
		} else {
			// 🚨 修复：捕获重排后的保存错误
			if err := storage.SavePlan(currentState); err != nil {
				fmt.Printf("❌ 计划保存失败: %v\n", err)
			}
			fmt.Println("\n================ 📋 插入突发后的计划 ================")
			fmt.Println(currentState.CurrentPlan)
		}
	}

	// ==========================================
	// 第三阶段：无限循环【精细微调】
	// ==========================================
	fmt.Println("\n================ 🛠️ 计划最后微调 ================")
	for {
		fmt.Print("👉 对当前计划满意吗？\n(直接 [回车] 存盘退出，或输入微调要求如 \"小说看久一点\"): ")
		if !scanner.Scan() {
			break
		}
		tweak := strings.TrimSpace(scanner.Text())
		
		if tweak == "" {
			// 🚨 修复：捕获最终保存的错误并给予正确反馈
			if err := storage.SavePlan(currentState); err != nil {
				fmt.Printf("❌ 计划保存失败: %v\n", err)
			} else {
				fmt.Printf("✅ [%s] 计划已成功保存！祝一切顺利，通关愉快！✨\n", targetDateStr)
			}

			// 📧 邮件推送
			if cfg.Email.Enabled {
				targetTo := cfg.Email.To
				if targetTo == "" {
					targetTo = cfg.Email.From
				}
				fmt.Printf("\n📧 正在将排程发送至邮箱 [%s] ...\n", targetTo)
				if err := email.SendPlan(&cfg.Email, targetDateStr, currentState.CurrentPlan, currentState.CFProblems...); err != nil {
					fmt.Printf("⚠️ 邮件发送失败: %v\n", err)
				} else {
					fmt.Println("✅ 邮件已成功发送！请查收。")
				}
			}
			break
		}

		fmt.Println("\n⏳ 正在让 Chronos 为你精修时间表...")
		if err := sch.Adjust(currentState, tweak); err != nil {
			fmt.Println("❌ 调整失败:", err)
			continue
		}
		
		// 🚨 修复：捕获微调循环中的保存错误
		if err := storage.SavePlan(currentState); err != nil {
			fmt.Printf("❌ 计划保存失败: %v\n", err)
		}
		fmt.Println("\n================ 📋 调整后的计划 ================")
		fmt.Println(currentState.CurrentPlan)
	}
}

// createInitialPlan 纯净的初版生成（加入目标日期推断与 Codeforces 题目绑定）
func createInitialPlan(sch *scheduler.Scheduler, scanner *bufio.Scanner, targetDate time.Time, cfProblems []codeforces.ProblemDetail) *storage.DayState {
	fmt.Print("\n👉 请选择风格 (1:严谨温和 2:元气看板娘，回车默认2): ")
	scanner.Scan()
	persona := "anime"
	if strings.TrimSpace(scanner.Text()) == "1" {
		persona = "gentle"
	}

	// 智能提示时间区间
	now := time.Now()
	if targetDate.Format("2006-01-02") == now.Format("2006-01-02") {
		fmt.Printf("\n👉 请输入今天剩余可用时间段 (如: 7:00-22:00，回车默认从现在 %s 到睡前): ", now.Format("15:04"))
	} else {
		fmt.Print("\n👉 请输入目标日期的可用时间段 (回车默认根据作息排全天): ")
	}
	scanner.Scan()
	timeRange := strings.TrimSpace(scanner.Text())

	fmt.Println("\n👉 请输入该日的主线任务 (每行一个，空行结束):")
	var tasks []string
	for {
		fmt.Printf("任务 %d: ", len(tasks)+1)
		if !scanner.Scan() {
			break
		}
		t := strings.TrimSpace(scanner.Text())
		if t == "" {
			break
		}
		tasks = append(tasks, t)
	}

	fmt.Println("\n⏳ Chronos 正在结合基础课表、CF 题单和任务池，生成【初版排程底稿】...")
	state, err := sch.Plan(persona, timeRange, tasks, targetDate, cfProblems)
	if err != nil {
		log.Fatal("生成失败:", err)
	}

	// 🚨 修复：捕获初版生成的保存错误
	if err := storage.SavePlan(state); err != nil {
		fmt.Printf("⚠️ 初版计划保存至本地失败: %v\n", err)
	}
	
	fmt.Println("\n================ 📋 初版排程出炉 ================")
	fmt.Println(state.CurrentPlan)
	return state
}