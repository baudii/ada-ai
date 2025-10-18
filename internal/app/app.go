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
	"github.com/baudii/ada-ai/internal/project"
	"github.com/baudii/ada-ai/pkg/utils"
	"github.com/tmc/langchaingo/llms"
	"golang.org/x/sync/errgroup"
)

const (
	business  = "business"
	technical = "technical"
	scope     = "scope"
	filler    = "filler"
	ext       = "json"
)

var navNames = [4]string{business, technical, scope, project.Structure}

type app struct {
	deg       int
	new       bool
	cfgroot   string
	aicfgroot string

	ada    *adacore.Ada
	runner Runner
	proj   project.Manager
}

type option func(*app)

// Runner defines an interface for reading project data.
type Runner interface {
	ReadProjdata(chan adacore.ProjectData)
}

// WithAIConfigRoot sets the AI configuration root path for the application. default is
// set from the variable common.AiConfigPath.
func WithAIConfigRoot(path string) option {
	return func(a *app) {
		a.aicfgroot = path
	}
}

// WithConfigRoot sets the configuration root path for the application. default is
// set from the variable common.ConfigPath.
func WithConfigRoot(path string) option {
	return func(a *app) {
		a.cfgroot = path
	}
}

// WithNew sets whether to create a new project folder or reuse an existing one.
// Default is false (reuse existing).
func WithNew(new bool) option {
	return func(a *app) {
		a.new = new
	}
}

// WithDegree sets the maximum degree of concurrency for the application.
func WithDegree(deg int) option {
	return func(a *app) {
		a.deg = deg
	}
}

// New creates a new application instance with the provided options.
func New(opts ...option) *app {
	a := &app{
		deg:       1,
		cfgroot:   common.ConfigPath,
		aicfgroot: common.AiConfigPath,
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

// InitAda initializes the Ada AI model with the specified LLM provider
// and configuration. It sets up both a JSON-capable AI model and a standard
// text AI model, and configures the Ada workflow with the provided options.
func (a *app) InitAda(provider string) error {
	slog.Info("initializing ada", "provider", provider)
	path := filepath.Join(a.aicfgroot, fmt.Sprintf("%s.json", provider))
	cfg, err := utils.ParseJSONConfigWithLocal[ai.Config](path)
	if err != nil {
		return fmt.Errorf("parse llm config %q: %w", path, err)
	}

	ai, err := ai.Register(provider, cfg.Options)
	if err != nil {
		return fmt.Errorf("register ai %q: %w", provider, err)
	}

	optsPath := filepath.Join(a.cfgroot, "ada.json")
	opts, err := utils.ParseJSONConfigWithLocal[adacore.Options](optsPath)
	if err != nil {
		return fmt.Errorf("parse ada options %q: %w", optsPath, err)
	}
	applyAdaDefaults(opts)

	a.ada = adacore.New(ai, adacore.WithOptions(*opts))
	return nil
}

// InitProject initializes a new local project based on the current project data
// in the Ada session. It determines the project path and creates a new
// local project instance, incrementing the project folder index if
// the 'new' flag is set.
func (a *app) InitProject() error {
	slog.Info("initializing project")
	projPath := a.ada.ResolveProjectPath()
	lastIdx, err := project.LastFolder(projPath)
	if err != nil {
		return fmt.Errorf("determine last folder: %w", err)
	}
	if a.new {
		lastIdx++
	}
	lp, err := project.New(filepath.Join(projPath, strconv.Itoa(lastIdx)))
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

	if err := a.MaterializeProject(ctx); err != nil {
		return fmt.Errorf("materialize project: %w", err)
	}

	slog.Debug("project materialized")
	return nil
}

// MaterializeProject generates and creates the project files
// concurrently based on the project structure defined in the
// project instance.
func (a *app) MaterializeProject(ctx context.Context) error {
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
		project.DefaultFolderHandler,
	); err != nil {
		cancel()
		_ = g.Wait()
		return fmt.Errorf("materialize project: %w", err)
	}

	return g.Wait()
}

// AddNavs adds navigation files to the project. It checks for existing
// navigation files and generates new ones if they are not found.
func (a *app) AddNavs() error {
	m := make(map[string][]any)
	m[business] = []any{a.ada.Projdata.ProjName, a.ada.Projdata.Summary}
	m[technical] = []any{a.ada.Projdata.Language, 0}
	m[scope] = []any{0, 1}
	m[project.Structure] = []any{0, 1, 2}

	var res []byte
	var err error
	for _, navName := range navNames {
		filename := fmt.Sprintf("%v.%v", navName, ext)
		slog.Debug("checking nav file", "file", filename)
		if !a.proj.TryLoadNav(filename, &res) {
			slog.Debug("generating new nav file", "file", filename)
			args := m[navName]
			res, err = a.sendInstructions(context.Background(), navName, args, llms.WithJSONMode())
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
	res, err := a.sendInstructions(ctx, filler, []any{0, 1, 2, 3, path})
	if err != nil {
		return fmt.Errorf("generate file %q: %w", path, err)
	}
	if err := os.WriteFile(path, res, 0o644); err != nil {
		return fmt.Errorf("write file %q: %w", path, err)
	}
	slog.Info("success", "path", path)
	return nil
}

func (a *app) sendInstructions(ctx context.Context, p string, args []any, opts ...llms.CallOption) ([]byte, error) {
	args, err := a.injectNavContent(args...)
	if err != nil {
		return nil, fmt.Errorf("inject nav content %q: %w", p, err)
	}

	sys, hum, err := a.sysHumanPrompts(p, args...)
	if err != nil {
		return nil, fmt.Errorf("system and human prompts: %w", err)
	}

	resp, err := a.ada.GenerateWithSys(ctx, sys, hum, opts...)
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
