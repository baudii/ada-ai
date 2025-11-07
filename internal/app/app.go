package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/core/gen"
	"github.com/baudii/ada-ai/internal/core/project"
	"golang.org/x/sync/errgroup"
)

// Constants that define folder names containing prompts.
const (
	business  = "business"
	technical = "technical"
	scope     = "scope"
	filler    = "filler"
)

const ext = "json"

// Run initializes and runs the CLI application. It sets up the AI model,
// configures the Ada AI workflow, and handles user input to generate and
// materialize a project based on the provided description.
//
// It also manages configuration loading and error handling throughout the process.
//
// Additional arguments can be passed to modify the behavior of the application.
func (a *app) Run(ctx context.Context) error {
	a.logger.Info("starting app session", "user", a.projectContext.UserName, "project", a.projectContext.Name)
	if err := a.AddNavs(ctx); err != nil {
		if saveErr := a.materializer.Materialize(nil, nil); saveErr != nil {
			return errors.Join(
				fmt.Errorf("add navs: %w", err),
				fmt.Errorf("materialize during recovery: %w", saveErr),
			)
		}
		return fmt.Errorf("add navs: %w", err)
	}

	if err := a.MaterializeProject(ctx); err != nil {
		return fmt.Errorf("materialize project: %w", err)
	}

	a.logger.Debug("project materialized")
	return nil
}

// AddNavs adds navigation files to the project. It checks for existing
// navigation files and generates new ones if they are not found.
func (a *app) AddNavs(ctx context.Context) error {
	m := make(map[string][]any)
	m[business] = []any{a.projectContext.Name, a.projectContext.Summary}
	m[technical] = []any{a.projectContext.Language, 0}
	m[scope] = []any{0, 1}
	m[project.Structure] = []any{0, 1, 2}

	var res []byte
	var err error
	for _, navName := range a.navNames {
		filename := fmt.Sprintf("%v.%v", navName, ext)
		a.logger.Debug("checking nav file", "file", filename)
		if res, err = a.navigator.LoadNav(filename); err != nil {
			a.logger.Debug("generating new nav file", "file", filename)
			args, ok := m[navName]
			if !ok {
				return fmt.Errorf("no args for nav %q", navName)
			}
			if res, err = a.sendInstructions(ctx, navName, args, gen.WithJSONMode()); err != nil {
				return fmt.Errorf("send instructions: %w", err)
			}
		}

		if err := a.navigator.AddNav(navName, ext, res); err != nil {
			return fmt.Errorf("add nav file: %w", err)
		}
		a.logger.Debug("added nav file", "file", filename)
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
	if err := a.materializer.Materialize(
		func(name string) error {
			a.logger.Debug("handling materialization", "name", name)
			g.Go(func() error {
				return a.fileGen(ctx, name)
			})
			return nil
		},
		nil,
	); err != nil {
		cancel()
		_ = g.Wait()
		return fmt.Errorf("materialize project: %w", err)
	}

	return g.Wait()
}

func (a *app) fileGen(ctx context.Context, path string) error {
	res, err := a.sendInstructions(ctx, filler, []any{0, 1, 2, 3, path})
	if err != nil {
		return fmt.Errorf("generate file %w", err)
	}

	if err := a.materializer.HandleFile(path, res); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	a.logger.Info("success", "path", path)
	return nil
}

func (a *app) sendInstructions(ctx context.Context, promptName string, args []any, opts ...gen.Option) ([]byte, error) {
	args, err := a.injectNavContent(args...)
	if err != nil {
		return nil, fmt.Errorf("inject nav content for %q: %w", promptName, err)
	}

	sys, hum, err := a.sysHumanPrompts(promptName, args...)
	if err != nil {
		return nil, fmt.Errorf("system and human prompts: %w", err)
	}

	resp, err := a.gen.GenerateWithSys(ctx, sys, hum, opts...)
	if err != nil {
		return nil, fmt.Errorf("generate with sys: %w", err)
	}

	return []byte(resp), nil
}

func (a *app) sysHumanPrompts(p string, args ...any) (string, string, error) {
	sys, err := a.gen.BuildPrompt(filepath.Join(p, "system.txt"))
	if err != nil {
		return "", "", fmt.Errorf("load system template: %w", err)
	}

	hum, err := a.gen.BuildPrompt(filepath.Join(p, "human.txt"), args...)
	if err != nil {
		return "", "", fmt.Errorf("load human template: %w", err)
	}

	return sys, hum, nil
}

func (a *app) injectNavContent(args ...any) ([]any, error) {
	for i := range args {
		if idx, ok := args[i].(int); ok {
			r, err := a.navigator.NavContent(defaultNavNames[idx])
			if err != nil {
				return nil, fmt.Errorf("retrieve nav %q: %w", defaultNavNames[idx], err)
			}
			args[i] = r
		}
	}
	return args, nil
}
