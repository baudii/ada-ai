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

	"github.com/baudii/ada-ai/internal/adacore"
	"github.com/baudii/ada-ai/internal/ai"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/internal/projects"
	"github.com/baudii/ada-ai/pkg/utils"
	"golang.org/x/sync/errgroup"
)

const ext = "json"

const (
	business  = "business"
	technical = "technical"
	scope     = "scope"
	filler    = "filler"
)

var navNames = [4]string{business, technical, scope, projects.Structure}

type Runner interface {
	ReadProjdata(chan adacore.ProjectData)
}

type app struct {
	deg    int
	ada    *adacore.Ada
	runner Runner
	proj   projects.Project
}

// New creates a new application instance with the specified degree of concurrency.
func New(deg int) *app {
	if deg < 1 {
		deg = 1
	}
	return &app{deg: deg}
}

// InitAda initializes the Ada AI model with the specified LLM provider
// and configuration. It sets up both a JSON-capable AI model and a standard
// text AI model, and configures the Ada workflow with the provided options.
func (a *app) InitAda(provider string) error {
	slog.Info("initializing ada", "provider", provider)
	path := filepath.Join(common.AiConfigPath, fmt.Sprintf("%s.json", provider))
	cfg, err := utils.ParseJSONConfigWithLocal[ai.Config](path)
	if err != nil {
		return fmt.Errorf("parse llm config %q: %w", path, err)
	}

	jsonAI, err := ai.RegisterJSON(provider, cfg.Options)
	if err != nil {
		return fmt.Errorf("register json ai %q: %w", provider, err)
	}

	textAI, err := ai.Register(provider, cfg.Options)
	if err != nil {
		return fmt.Errorf("register normal ai %q: %w", provider, err)
	}

	optsPath := filepath.Join(common.ConfigPath, "ada.json")
	opts, err := utils.ParseJSONConfigWithLocal[adacore.Options](optsPath)
	if err != nil {
		return fmt.Errorf("parse ada options %q: %w", optsPath, err)
	}
	applyAdaDefaults(opts)

	a.ada = adacore.New(jsonAI, adacore.WithModel("text", textAI), adacore.WithOptions(*opts))
	return nil
}

// InitProject initializes a new local project based on the current project data
// in the Ada session. It determines the project path and creates a new
// local project instance, incrementing the project folder index if
// the 'new' flag is set.
func (a *app) InitProject(new bool) error {
	slog.Info("initializing project", "new", new)
	projPath := a.ada.ResolveProjectPath()
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

	a.proj = lp
	slog.Debug("created local project", "instance", lp)
	return nil
}

// ReceiveProjdata retrieves project data using the provided Runner
// implementation and adds it to the Ada session.
func (a *app) ReceiveProjdata(runner Runner) {
	slog.Info("receiving project data", "runner", fmt.Sprintf("%T", runner))
	c := make(chan adacore.ProjectData)
	go runner.ReadProjdata(c)
	a.ada.AddProjectData(<-c)
	a.runner = runner
}

// Run initializes and runs the CLI application. It sets up the AI model,
// configures the Ada AI workflow, and handles user input to generate and
// materialize a project based on the provided description.
//
// It also manages configuration loading and error handling throughout the process.
//
// Additional arguments can be passed to modify the behavior of the application.
func (a *app) Run(ctx context.Context) error {
	slog.Info("starting app session", "user", a.ada.Projdata.UserName, "project", a.ada.Projdata.ProjName)
	if err := a.AddNavs(); err != nil {
		return fmt.Errorf("add navs: %w", err)
	}

	if err := utils.PrintTree(os.Stdout, a.proj.Structure(), ""); err != nil {
		return fmt.Errorf("print project tree: %w", err)
	}

	if err := a.materializeProject(ctx); err != nil {
		return fmt.Errorf("materialize project: %w", err)
	}

	slog.Debug("project materialized")
	return nil
}

func (a *app) materializeProject(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(a.deg)

	if err := a.proj.Materialize(
		func(path string) error {
			g.Go(func() error {
				return a.fileGen(ctx, path)
			})
			return nil
		},
		projects.DefaultFolderHandler,
	); err != nil {
		cancel()
		_ = g.Wait()
		return fmt.Errorf("materialize project: %w", err)
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("materialize project: %w", err)
	}

	return nil
}

// AddNavs generates and adds navigation files to the local project. It checks for existing
// navigation files and generates new ones using the Ada AI model if they are not found.
// The generated navigation files are added to the project for later use.
func (a *app) AddNavs() error {
	m := make(map[string][]any)
	m[business] = []any{a.ada.Projdata.ProjName, a.ada.Projdata.Summary}
	m[technical] = []any{a.ada.Projdata.Language, 0}
	m[scope] = []any{0, 1}
	m[projects.Structure] = []any{0, 1, 2}

	var res []byte
	var err error
	for _, navName := range navNames {
		filename := fmt.Sprintf("%v.%v", navName, ext)
		slog.Debug("checking nav file", "file", filename)
		if !a.proj.TryLoadNav(filename, &res) {
			slog.Debug("generating new nav file", "file", filename)
			args := m[navName]
			res, err = a.sendInstructions(context.Background(), "main", navName, args...)
			if err != nil {
				return fmt.Errorf("business description: %w", err)
			}
		}

		if err := a.proj.AddNav(navName, ext, res); err != nil {
			return fmt.Errorf("add nav file %q: %w", filename, err)
		}
		slog.Debug("added nav file", "file", filename)
	}
	return nil
}

func (a *app) fileGen(ctx context.Context, path string) error {
	slog.Debug("generating file", "path", path)
	res, err := a.sendInstructions(ctx, "text", filler, []any{0, 1, 2, 3, path}...)
	if err != nil {
		return fmt.Errorf("generate file %q: %w", path, err)
	}
	if err := os.WriteFile(path, res, 0o644); err != nil {
		return fmt.Errorf("write file %q: %w", path, err)
	}
	slog.Info("success", "path", path)
	return nil
}

func (a *app) sendInstructions(ctx context.Context, aiKey, p string, args ...any) ([]byte, error) {
	args, err := a.injectNavContent(args...)
	if err != nil {
		return nil, fmt.Errorf("inject nav content %q: %w", p, err)
	}

	sys, hum, err := a.sysHumanPrompts(p, args...)
	if err != nil {
		return nil, fmt.Errorf("system and human prompts: %w", err)
	}

	resp, err := a.ada.GenerateWithSys(ctx, aiKey, sys, hum)
	if err != nil {
		return nil, fmt.Errorf("generate with sys: %w", err)
	}

	return []byte(resp.Choices[0].Content), nil
}

func (a *app) sysHumanPrompts(p string, args ...any) (string, string, error) {
	sys, err := a.ada.PromptFromTemplate(filepath.Join(p, "system.txt"))
	if err != nil {
		return "", "", fmt.Errorf("load system template: %w", err)
	}

	hum, err := a.ada.PromptFromTemplate(filepath.Join(p, "human.txt"), args...)
	if err != nil {
		return "", "", fmt.Errorf("load human template: %w", err)
	}

	return sys, hum, nil
}

func (a *app) injectNavContent(args ...any) ([]any, error) {
	for i := range args {
		if idx, ok := args[i].(int); ok {
			r, err := a.proj.NavContent(navNames[idx])
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

func applyAdaDefaults(o *adacore.Options) {
	if o.PromptsRoot == "" {
		o.PromptsRoot = common.PromptsPath
	}
	if o.ProjectsRoot == "" {
		o.ProjectsRoot = common.ProjectsPath
	}
}
