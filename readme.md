# 一款基于ai的私人生活学习助理

D:.
│   go.mod
│   readme.md
│   
├───cmd
│   └───chronos
│           main.go
│           
└───internal
    ├───ai
    │       client.go
    │       
    ├───codeforces
    │       crawler.go
    │       default_problems.go
    │       service.go
    │       tracker.go
    │       translator.go
    │       types.go
    │       
    ├───config
    │       config.go
    │       
    ├───email
    │       render.go
    │       sender.go
    │       
    ├───scheduler
    │       scheduler.go
    │       
    └───storage
            storage.go

            
## 课表
第一阶段我决定先用官网的汇总课表
后续加上微信小程序的逆向抓取，每天校检具体课

最后还要有手动临时更改课表的


## 存储（全部用json格式）
1. 静态底座 (Base Schedule)：绝对真理
定位： 就像操作系统的 ROM。它存的是教务系统的固定课表，是整个日程的锚点。

特点： 极少修改。哪怕后面的临时任务安排得再乱，只要查这个文件，你就知道今天下午 2 点在教二确实有一节必修课。

2. 动态任务池 (Flexible Pool)：混沌变量
定位： 存你说的“临时实践、可以变、可以推的事件”。

特点： 它是高度动态的。比如突然有了个逆向工程的灵感要写代码，或者今晚想临时加一组引体向上，这些没有死板时间限制、但有优先级（Priority）的任务全扔进这里。

3. 生成的课表缓存 (Daily Snapshot)：最终呈现层
为什么这是神来之笔？ 很多人会犯的错，是每次在终端敲命令查看日程时，都让程序去重新合并课表、甚至重新调一次 AI。这就太慢了！

你的设计： 每天凌晨（或每次你执行 ava plan 时），程序把“底座”和“任务池”拿出来，或者扔给 AI 算一遍，然后把算好的最终时间轴写死存进这个缓存文件里。

结果： 白天你无数次在终端敲 ava show，程序只是瞬间去读这个缓存文件，毫秒级输出，体验直接拉满。



## ai处理

## 邮件功能

## 定期处理缓存

## 人物提示词
比如：东亚高考版，衡水版，养生版

晚上/周日：
AI 主动生成明天/下周计划
↓
早晨：
再给你一份当天最终计划
↓
你随时自然语言修改：
“下午有事。”
“今天不想学算法。”
“这个作业比预计难，至少还要 2 小时。”
“明天我要睡到 10 点。”
↓
Scheduler 重新规划
↓
AI 回复新的安排
↓
系统更新任务状态，并重新规划剩余时间。

## Codeforces 每日特训功能
1. **配置文件 (YAML 格式)**：
   - 配置文件存储于 `~/.chronos/config.yaml`（原 `config.json` 首次运行时自动无损迁移）。
   - 通过 `codeforces.enabled` (bool) 控制是否开启抓题功能。
   - 通过 `codeforces.daily_count` (int) 控制每日训练题目数量（默认 2 道）。

2. **题单与底座 (Base Schedule)**：
   - 题单集中存储在 `~/.chronos/tasks/base.json` 的 `codeforces_problems` 列表中，包含 60 道经典入门到进阶题（支持随时追加）。
   - 进度游标由 `tasks/cf_tracker.json` 自动记录，记住每天从哪里开始；**同日多次运行或微调保持幂等**，换日自动推进。

3. **独立题目翻译与持久化缓存**：
   - 翻译独立于每日日程安排，首次分配题目时自动抓取 Codeforces 英文原题，由算法教练 Prompt 进行规范中文翻译并永久写入 `tasks/cf_translations.json`。
   - 微调日程（如突发事件、重新规划时限）时零次调用翻译 AI，毫秒级复用。

4. **邮件网页卡片升级**：
   - 邮件采用现代化响应式卡片排版。
   - 配备 Codeforces 专属特训面板、经典天梯段位评分徽章、原题直达链接按钮与格式化中文题意解析。