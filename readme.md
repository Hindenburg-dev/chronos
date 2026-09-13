# Chronos

> An AI-powered personal scheduler written in Go.

Chronos is a personal scheduling tool I built for myself.

It takes my available time, tasks, existing schedule, and other constraints, then uses an AI model to generate a daily plan.

I originally started this as a small project while learning Go. It has grown quite a bit since then, and a significant part of the implementation was AI-assisted. I'm still learning and gradually trying to understand, improve, and refactor the code myself.

So this is **not a polished production-ready application**. It's primarily a personal project and a learning experiment.

## What can it do?

* 🗓️ Generate daily schedules using AI
* 🤖 Support multiple AI providers/models
* 📚 Manage a base schedule and flexible tasks
* 🔄 Re-plan when unexpected events happen
* 💾 Persist schedules and task data locally
* ⚡ Cache generated data to avoid unnecessary API calls
* 📧 Send schedules by email
* 🧩 Maintain a daily Codeforces practice plan
* 🌐 Fetch and translate Codeforces problems
* 🎭 Customize the AI's scheduling style/persona

The general idea is:

```text
             ┌─────────────────┐
             │  Base Schedule  │
             └────────┬────────┘
                      │
             ┌────────▼────────┐
             │ Flexible Tasks  │
             └────────┬────────┘
                      │
             ┌────────▼────────┐
             │   Codeforces    │
             │     Tasks       │
             └────────┬────────┘
                      │
                      ▼
                ┌───────────┐
                │    AI     │
                │ Scheduler │
                └─────┬─────┘
                      │
                      ▼
             ┌─────────────────┐
             │   Daily Plan    │
             └─────────────────┘
                      │
              unexpected event?
                      │
                      ▼
             ┌─────────────────┐
             │    Re-plan      │
             └─────────────────┘
```

## Why did I build this?

I wanted a scheduler that could deal with the fact that my days are not perfectly predictable.

A traditional calendar is good at answering:

> "What is scheduled at 2 PM?"

But I wanted something closer to:

> "I have these things to do, this much free time, different energy levels throughout the day, and something unexpected just happened. What should I do now?"

That's where the AI component comes in.

Chronos treats the AI as a **planner**, rather than the source of truth.

The actual schedule, tasks, and cached results are stored locally. The AI is used to make decisions based on that information.

## Project structure

The project is currently organized roughly by responsibility:

```text
cmd/
└── chronos/          # Application entry point

internal/
├── ai/               # AI client and model interaction
├── codeforces/       # Codeforces problems, tracking and translation
├── config/           # Configuration
├── email/            # Email functionality
├── scheduler/        # Scheduling logic
└── storage/          # Local data persistence
```

This structure is still evolving.

Some parts are cleaner than others, and there are definitely places that could be designed better.

## Tech stack

* **Go**
* JSON / YAML for local data
* AI APIs
* Codeforces
* SMTP/email

The project currently uses Go's standard library heavily, with only a small number of external dependencies.

## Running

### Requirements

* Go 1.26+
* An API key for a supported AI provider
* Optional: SMTP credentials if email functionality is enabled

### Build

```bash
git clone https://github.com/Hindenburg-dev/chronos.git
cd chronos

go build ./cmd/chronos
```

Then run the resulting executable.

> Configuration details are still being cleaned up. This section will probably change as the project develops.

## Configuration

Chronos uses a local configuration file for things such as:

* AI provider
* API key
* model
* email settings
* scheduler preferences

**Do not commit API keys or other secrets to the repository.**

## Codeforces integration

Chronos can maintain a daily Codeforces practice plan.

The current workflow is roughly:

```text
Problem pool
     ↓
Tracker
     ↓
Today's problems
     ↓
Translation / metadata
     ↓
Daily schedule
```

Translated problem information can also be cached locally so that the same problem does not need to be translated repeatedly.

## Current status

Chronos is very much a **work in progress**.

Some things I'm currently interested in improving:

* Simplifying the architecture
* Making the code easier to understand and maintain
* Improving error handling
* Reducing unnecessary AI API calls
* Improving configuration and installation
* Adding better tests
* Improving the CLI experience
* Eventually adding a GUI

I am intentionally not treating the current architecture as the final fixed version. To be honest, I feel this project is a bit of an unfinished work in rough shape. I didn't handle the task pool module very well — the worst part is that I directly dumped the entire problem set into base.json with no proper processing. Right now the prompt fed to the AI is ridiculously bloated, which even causes time offset errors during scheduling. On top of that, the total code volume has far exceeded what I can comfortably maintain at my current skill level. For now I can only treat this project like a hands-on Go learning workbook, and try my best to get a solid feel for what the workflow of a real-world project looks like.

## A note about AI-assisted development

A substantial amount of Chronos was written with the help of AI coding assistants.

I'm learning Go while working on this project, so AI has sometimes written code that I didn't fully understand at the time.

That's also one of the reasons I'm putting this project on GitHub.

I'd like to gradually move from:

```text
AI writes code → I make it run
```

towards:

```text
I understand the design → AI helps with implementation → I review the result
```

If you happen to look through the code and find something that is unnecessarily complicated, poorly designed, or just plain wrong, **feel free to tell me.**

I'm here to learn.

## Disclaimer

This project is primarily built for my own use.

It may contain rough edges, questionable design decisions, incomplete documentation, or things that only make sense because I wrote them at 2 AM.

Use it at your own risk.

And if you are a Go developer reading this:

**please roast my code responsibly.**
