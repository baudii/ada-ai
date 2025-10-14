# Ada AI

Ada is the AI-agent that is designed to generate a project from scratch using a Large Language Model (LLM). It collects a short brief about the product you want to build, asks the configured LLM to propose a complete project structure, iteratively reflects on the response to raise its quality, and finally materializes the folders and files on disk. The repository also contains helpers for prompt management and daily log rotation so you can tune and observe the workflow end to end.

## Key Capabilities
- **Prompt-driven architecture planning** - reusable prompt templates turn a short idea into a coherent product brief and project tree.
- **LLM abstraction layer** - integrates with Ollama via `langchaingo`, keeping provider options in JSON configuration files.
- **Reflection loop** - optional multi-pass evaluation (relevance, accuracy, completeness) with retries until thresholds are met.
- **Project materialization** - converts the JSON tree into a real directory hierarchy under `artifacts/.projects/<user>/<project>/<index>/`.
- **Configuration layering** - every config file can be overridden via a `.local` companion without touching the defaults.
- **Structured logging** - `pkg/dilog` writes daily rotating slog logs to the runtime directory so you can trace LLM conversations and filesystem work.

## Repository Layout
- `cmd/ada-cli` - entry point that wires flags and launches the app.
- `internal/app` - top-level orchestration (config loading, AI initialization, run loop).
- `internal/cli` - interactive runner that captures project metadata.
- `internal/adacore` - core workflow (prompt resolution, reflection, LLM calls).
- `internal/projects` - adapters that materialize project descriptions on disk.
- `pkg/utils` - helpers for JSON configs, console I/O, and filesystem utilities.
- `pkg/dilog` - daily log writer and slog setup.
- `data/` - source of prompt templates and default configuration (copied to `artifacts/` at build time).

## Prerequisites
- Go 1.25.x (see `go.mod`).
- An Ollama runtime with a chat model available (defaults expect `gemma3:4b`).
  - Only ollama option is available at the moment.
- `make` (optional but recommended for repeatable build and copy steps).

## Setup
1. Clone the repository.
2. Run `make all` to build the CLI and copy configuration and prompts into `artifacts/`.
   - On Windows the Makefile relies on `xcopy`; ensure it is available.
3. (Optional) Run `make test` to execute the Go test suite and view coverage.

The compiled binary lives at `artifacts/ada-cli` (or `artifacts/ada-cli.exe` on Windows).
All runtime data (configs, prompts, logs, generated projects) is read relative to this directory because `internal/common` resolves paths from `os.Executable()`.

### Manual build without `make`
```powershell
go build -ldflags="-s -w" -o artifacts/ada-cli ./cmd/ada-cli
xcopy /E /I /Y data\configuration artifacts\data\configuration
xcopy /E /I /Y data\prompts artifacts\data\prompts
```
On Unix-like shells replace the `xcopy` commands with `cp -r`.

## Configuration
Configuration lives under `artifacts/data/configuration/` at runtime:

| File | Purpose |
| --- | --- |
| `llm-provider.json` | Selects the LLM backend (`provider`) and its options. For Ollama set `model`, `url`, `keep_alive`, `format`, `http_timeout`, and similar keys. |
| `ada.json` | Controls Ada's workflow: request timeout, where projects are written, reflection settings (`reflectionDepth`, `reflectionThreshold`), and default model call options. |
| `dilog.json` | Logging strategy (timezone, log directory, filename prefix, log level). |

Add a `.local` suffix (for example `llm-provider.local.json`) to override specific fields locally without editing the defaults. Files are merged with the base JSON using `pkg/utils.ParseJSONConfigWithLocal`.

`artifacts/data/user_data.json` caches the last project brief. Delete it or edit the JSON to start fresh; otherwise the CLI reuses the stored data and skips interactive prompts.

## Running the CLI
```powershell
artifacts\ada-cli.exe          # Windows
./artifacts/ada-cli            # macOS/Linux
```
- First run: you are prompted for nickname, project name, preferred language, and a short summary. The answers are saved to `artifacts/data/user_data.json`.
- Ada assembles the prompt templates from `artifacts/data/prompts`, asks the LLM for a project structure, performs reflection cycles if they are enabled, prints the resulting tree, and writes the directories and files to `artifacts/.projects/<user>/<project>/<n>/`.
- All stdout and stderr is mirrored to daily log files via `pkg/dilog`; check the `logs` folder next to the binary for historical runs.

### Debug mode
The CLI ships a helper mode for inspecting intermediate artifacts without invoking the full workflow:
```
artifacts/ada-cli -debug -stage=0   # replay a saved JSON structure
artifacts/ada-cli -debug -stage=1   # run prompt-to-structure generation and print the tree
```
Debug assets are loaded from `artifacts/data/debugging/`. Copy sample payloads there to iterate on prompts safely.

## Prompt Library
Prompt templates live in `artifacts/data/prompts/` (copied from `data/prompts/`). Notable files include:
- `project-structure-template.txt` - asks the LLM for a complete directory layout.
- `reflect-template.txt` and `reflect-template-short.txt` - drive the scoring and feedback loop.
- `improve-template.txt` - feeds reflection feedback back into the model.
- `product/system.txt` - generates a structured product specification that can seed later stages.

You can tune these templates to suit new stacks or evaluation strategies; updates take effect on the next run.

## Development
- Run `go test ./...` (or `make test`) before submitting changes; coverage output is written to `coverage.out`.
- Run `go fmt ./...` to keep formatting consistent (not wired into the Makefile yet).
- Logs and generated projects accumulate under `artifacts/`; clean them with `make clean`.

## Roadmap Ideas
- Finish the generation part. The goal is to create a project that can be executed from a simple user prompt.
- Add more LLM providers by implementing new adapters in `internal/ai`.
- Extend `internal/projects` with generators that go beyond structure (scaffold README, CI files).
- Persist reflection feedback for analysis and human review.
- Add GUI interface. Console UX is only there for the debugging.

