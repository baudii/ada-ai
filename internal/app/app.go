package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/baudii/ada-ai/internal/adacore"
	"github.com/baudii/ada-ai/internal/ai"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/internal/projects"
	"github.com/baudii/ada-ai/pkg/utils"
)

const ext = "json"

const (
	business  = "business"
	technical = "technical"
	scope     = "scope"
	filler    = "filler"
)

var navNames = [4]string{business, technical, scope, projects.Structure}

var (
	new bool = false
)

var defaultCfg adacore.Options = adacore.Options{
	Timeout: "3m",
}

type Runner interface {
	Projdata(chan adacore.ProjectData)
}

type result struct {
	content []byte
	path    string
	err     error
}

// Run initializes and runs the CLI application. It sets up the AI model,
// configures the Ada AI workflow, and handles user input to generate and
// materialize a project based on the provided description.
//
// It also manages configuration loading and error handling throughout the process.
//
// Additional arguments can be passed to modify the behavior of the application.
func Run(ada *adacore.Ada, runner Runner, args ...any) error {
	parseArgs(args...)
	projPath := ada.ResolveProjectPath()
	lastIdx, err := projects.LastFolder(projPath)
	if err != nil {
		return fmt.Errorf("determine last folder: %w", err)
	}
	if new {
		lastIdx++
	}

	lp, err := projects.New(filepath.Join(projPath, strconv.Itoa(lastIdx)))
	if err != nil {
		return fmt.Errorf("create local project: %w", err)
	}

	slog.Debug("created local project", "instance", lp)
	m := make(map[string][]any)
	m[business] = []any{ada.Projdata.ProjName, ada.Projdata.Summary}
	m[technical] = []any{ada.Projdata.Language, 0}
	m[scope] = []any{0, 1}
	m[projects.Structure] = []any{0, 1, 2}

	for _, navName := range navNames {
		filename := fmt.Sprintf("%v.%v", navName, ext)
		slog.Debug("checking nav file", "file", filename)
		var res []byte
		if new || !lp.TryLoadNav(filename, &res) {
			slog.Debug("generating new nav file", "file", filename)
			args := m[navName]
			args, err = updateArgs(args, lp.NavContent)
			if err != nil {
				return fmt.Errorf("update args for %q: %w", navName, err)
			}
			res, err = sendInstructions(context.Background(), ada, "main", navName, args...)
			if err != nil {
				return fmt.Errorf("business description: %w", err)
			}
		}

		err := lp.AddNav(navName, ext, res)
		if err != nil {
			return fmt.Errorf("add nav file %q: %w", filename, err)
		}
		slog.Debug("added nav file", "file", filename)
	}

	err = utils.PrintTree(os.Stdout, lp.Structure(), "")
	if err != nil {
		return fmt.Errorf("print project tree: %w", err)
	}

	wg := sync.WaitGroup{}
	c := make(chan result)
	sem := make(chan struct{}, 4) // semaphore
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	matErr := lp.Materialize(func(path string) error {
		slog.Debug("generating file", "path", path)
		args := []any{0, 1, 2, 3}
		args, err = updateArgs(args, lp.NavContent)
		args = append(args, path)
		if err != nil {
			return fmt.Errorf("update args for file %q: %w", path, err)
		}
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()
			res, err := sendInstructions(ctx, ada, "text", filler, args...)
			c <- result{content: res, err: err, path: path}
		})
		return nil
	}, func(path string) error {
		return os.MkdirAll(path, 0755)
	})

	go func() {
		wg.Wait()
		close(c)
	}()

	if matErr != nil {
		slog.Debug("cancelling generation due to fail", "error", matErr)
		cancel()
	}

	for v := range c {
		if v.err != nil {
			slog.Error("generate file", "path", v.path, "error", v.err)
			continue
		}
		err = os.WriteFile(v.path, v.content, 0644)
		if err != nil {
			slog.Error("write file", "path", v.path, "error", err)
		}
		slog.Info("success", "path", v.path)
	}

	if matErr != nil {
		return fmt.Errorf("materialize project: %w", matErr)
	}

	slog.Debug("project materialized")
	return nil
}

func InitAda(provider string, runner Runner) (*adacore.Ada, error) {
	path := filepath.Join(common.AiConfigPath, fmt.Sprintf("%s.json", provider))
	cfg, err := utils.ParseJSONConfigWithLocal[ai.Config](path)
	if err != nil {
		return nil, fmt.Errorf("parse llm config %q: %w", path, err)
	}

	cfg.Options["format"] = "json"
	jsonai, err := ai.Register(provider, cfg.Options)
	if err != nil {
		return nil, fmt.Errorf("register json llm: %w", err)
	}

	cfg.Options["format"] = ""
	normalai, err := ai.Register(provider, cfg.Options)
	if err != nil {
		return nil, fmt.Errorf("register normal llm: %w", err)
	}

	var opts *adacore.Options
	cfgPath := filepath.Join(common.ConfigPath, "ada.json")
	opts, err = utils.ParseJSONConfigWithLocal[adacore.Options](cfgPath)
	if err != nil {
		slog.Error("failed to parse json configuration", "path", cfgPath, "error", err)
		opts = &defaultCfg
	} else {
		slog.Info("succesffully parsed json configuration", "cfg", opts)
	}

	if opts.PromptsRoot == "" {
		opts.PromptsRoot = common.PromptsPath
	}
	if opts.ProjectsRoot == "" {
		opts.ProjectsRoot = common.ProjectsPath
	}

	ada := adacore.New(jsonai, adacore.WithModel("text", normalai), adacore.WithOptions(*opts))
	c := make(chan adacore.ProjectData)
	go runner.Projdata(c)
	data := <-c
	ada.AddProjectData(data)
	slog.Info("starting app session", "user", ada.Projdata.UserName, "project", ada.Projdata.ProjName)
	return ada, nil
}

func sendInstructions(ctx context.Context, ada *adacore.Ada, aiKey, p string, args ...any) ([]byte, error) {
	sys, hum, err := sysHumanPrompts(ada, p, args...)
	if err != nil {
		return nil, fmt.Errorf("system and human prompts: %w", err)
	}

	resp, err := ada.GenerateWithSys(ctx, aiKey, sys, hum)
	if err != nil {
		return nil, fmt.Errorf("generate with sys: %w", err)
	}

	return []byte(resp.Choices[0].Content), nil
}

func sysHumanPrompts(ada *adacore.Ada, p string, args ...any) (string, string, error) {
	sys, err := ada.PromptFromTemplate(filepath.Join(p, "system.txt"))
	if err != nil {
		return "", "", fmt.Errorf("load system template: %w", err)
	}

	hum, err := ada.PromptFromTemplate(filepath.Join(p, "human.txt"), args...)
	if err != nil {
		return "", "", fmt.Errorf("load human template: %w", err)
	}

	return sys, hum, nil
}

func updateArgs(args []any, getNav func(string) ([]byte, error)) ([]any, error) {
	for i := range args {
		if idx, ok := args[i].(int); ok {
			r, err := getNav(navNames[idx])
			if err != nil {
				return nil, fmt.Errorf("retrieve nav %q: %w", navNames[idx], err)
			}

			buf := &bytes.Buffer{}
			if err := json.Compact(buf, r); err != nil {
				return nil, fmt.Errorf("compact system prompt: %w", err)
			}
			args[i] = buf.String()
		}
	}
	return args, nil
}

func parseArgs(args ...any) {
	for _, v := range args {
		switch v := v.(type) {
		case bool:
			new = v
		default:
			panic("unknown argument")
		}
	}
}
