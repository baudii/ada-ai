package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/ada"
	"github.com/baudii/ada-ai/internal/ai"
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
	deg      int
	mode     project.Mode
	gen      Generator
	dp       DirProvider
	proj     project.Manager
	projData ProjectData
	opts     *Options
}

// Generator defines methods for generating AI content and handling prompts.
type Generator interface {
	PromptFromTemplate(filename string, input ...any) (string, error)
	GenerateWithSys(ctx context.Context, sys, user string, callOptions ...llms.CallOption) (*llms.ContentResponse, error)
}

// DirProvider defines an interface for providing folder names based on a base path
// and a flag indicating whether to create a new folder.
type DirProvider interface {
	ProjectFolder(base string, mode project.Mode) (string, error)
}

// Runner defines an interface for reading project data.
type Runner interface {
	ReadProjdata() ProjectData
}

// InitAda initializes the Ada AI model with the specified LLM provider
// and configuration. It sets up both a JSON-capable AI model and a standard
// text AI model, and configures the Ada workflow with the provided options.
func (a *app) InitAda(provider, configPath string) error {
	slog.Info("initializing ada", "provider", provider)
	cfg, err := utils.ParseJSONConfigWithLocal[ai.Config](configPath)
	if err != nil {
		return fmt.Errorf("parse llm config %q: %w", configPath, err)
	}

	ai, err := ai.Register(provider, cfg.Options)
	if err != nil {
		return fmt.Errorf("register ai %q: %w", provider, err)
	}

	ada := ada.New(ai,
		ada.WithPromptsRoot(a.opts.PromptsRoot),
		ada.WithTimeout(a.opts.Timeout),
		ada.WithReflection(a.opts.Reflection))
	a.gen = ada
	return nil
}

// InitLocalProject initializes a new local project based on the current project data
// in the Ada session. It determines the project path and creates a new
// local project instance, incrementing the project folder index if
// the 'new' flag is set.
func (a *app) InitLocalProject() error {
	slog.Info("initializing project")
	base := filepath.Join(a.opts.ProjectsRoot, a.projData.UserName, a.projData.ProjName)
	path, err := a.dp.ProjectFolder(base, a.mode)
	if err != nil {
		return fmt.Errorf("project folder: %w", err)
	}
	lp, err := project.New(path)
	if err != nil {
		return fmt.Errorf("create local project %q: %w", path, err)
	}
	a.proj = lp
	slog.Debug("created local project", "path", path)
	return nil
}

// Run initializes and runs the CLI application. It sets up the AI model,
// configures the Ada AI workflow, and handles user input to generate and
// materialize a project based on the provided description.
//
// It also manages configuration loading and error handling throughout the process.
//
// Additional arguments can be passed to modify the behavior of the application.
func (a *app) Run(ctx context.Context) error {
	slog.Info("starting app session", "user", a.projData.UserName, "project", a.projData.ProjName)
	if err := a.AddNavs(ctx); err != nil {
		return fmt.Errorf("add navs: %w", err)
	}

	if err := a.MaterializeProject(ctx); err != nil {
		return fmt.Errorf("materialize project: %w", err)
	}

	slog.Debug("project materialized")
	return nil
}

// AddNavs adds navigation files to the project. It checks for existing
// navigation files and generates new ones if they are not found.
func (a *app) AddNavs(ctx context.Context) error {
	m := make(map[string][]any)
	m[business] = []any{a.projData.ProjName, a.projData.Summary}
	m[technical] = []any{a.projData.Language, 0}
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
			res, err = a.sendInstructions(ctx, navName, args, llms.WithJSONMode())
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

	resp, err := a.gen.GenerateWithSys(ctx, sys, hum, opts...)
	if err != nil {
		return nil, fmt.Errorf("generate with sys: %w", err)
	}

	return []byte(resp.Choices[0].Content), nil
}

func (a *app) sysHumanPrompts(p string, args ...any) (string, string, error) {
	sys, err := a.gen.PromptFromTemplate(filepath.Join(p, "system.txt"))
	if err != nil {
		return "", "", fmt.Errorf("load system template: %w", err)
	}

	hum, err := a.gen.PromptFromTemplate(filepath.Join(p, "human.txt"), args...)
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
