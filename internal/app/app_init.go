package app

import (
	"fmt"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/ada"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/internal/project"
	"github.com/baudii/ada-ai/pkg/utils"
)

const ConfigFile = "ada.json"

// Options is the configuration for Ada AI workflow.
type Options struct {
	Timeout      string            `json:"requestTimeout"`
	ProjectsRoot string            `json:"projectsRoot"`
	PromptsRoot  string            `json:"promptsRoot"`
	Reflection   ada.ReflectConfig `json:"reflection"`
}

// ProjectData is the context of the current project.
type ProjectData struct {
	UserName string `json:"userName"`
	ProjName string `json:"projName"`
	Language string `json:"language"`
	Summary  string `json:"summary"`
}

type option func(*app)

// WithGen sets the Generator instance for the application to generate content.
func WithGen(gen Generator) option {
	return func(a *app) {
		a.gen = gen
	}
}

// WithDirProvider sets the DirProvider for the application to provide project folder names.
func WithDirProvider(dp DirProvider) option {
	return func(a *app) {
		a.dp = dp
	}
}

// WithOptions sets the application options.
func WithOptions(opts *Options) option {
	return func(a *app) {
		a.opts = opts
	}
}

// WithProject sets the project manager for the application.
func WithProject(proj project.Manager) option {
	return func(a *app) {
		a.proj = proj
	}
}

// WithMode sets whether to create a new project folder or reuse an existing one.
// Default is false (reuse existing).
func WithMode(mode project.Mode) option {
	return func(a *app) {
		a.mode = mode
	}
}

// WithProjectData sets the project data for the application.
func WithProjectData(data ProjectData) option {
	return func(a *app) {
		a.projData = data
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
		deg: 1,
	}
	for _, o := range opts {
		o(a)
	}

	return a
}

// ParseAppOptions parses the application options from a JSON file located
// in the specified path.
func ParseAppOptions(path string) (*Options, error) {
	optsPath := filepath.Join(path, ConfigFile)
	opts, err := utils.ParseJSONConfigWithLocal[Options](optsPath)
	if err != nil {
		return nil, fmt.Errorf("parse ada options %q: %w", optsPath, err)
	}
	if opts.PromptsRoot == "" {
		opts.PromptsRoot = common.PromptsPath
	}
	if opts.ProjectsRoot == "" {
		opts.ProjectsRoot = common.ProjectsPath
	}
	return opts, nil
}
