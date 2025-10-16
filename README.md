# Ada AI

**Ada AI** generates complete software projects from a short idea using an LLM.
It plans the structure, reflects to improve quality, and creates real folders and files.

---

## Features

* Prompt-based project generation
* Reflection loop for quality checks
* Configurable LLM via Ollama
* Auto-writes projects under `artifacts/.projects/`
* Daily rotating logs

---

## Setup

```bash
git clone <repo>
make all
# or manually
go build -o artifacts/ada-cli ./cmd/ada-cli
cp -r data/* artifacts/data/
```

---

## Run

```bash
./artifacts/ada-cli
```

Answer the prompts; Ada will generate the project and logs under `artifacts/`.

---

## Config

`artifacts/data/configuration/`

* `llm-provider.json` – model and API settings
* `ada.json` – workflow and reflection options
* `dilog.json` – logging

Use `.local` files to override defaults.

---

## Roadmap

* More LLM providers
* Richer scaffolding (README, CI, etc.)
* GUI interface
