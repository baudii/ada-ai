package app

import (
	"fmt"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/ada"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/internal/project"
	"github.com/baudii/ada-ai/internal/project/folder"
	"github.com/baudii/ada-ai/pkg/jsonx"
)

const ConfigFile = "ada.json"

var defaultNavNames = [4]string{business, technical, scope, project.Structure}

type app struct {
	deg          int
	mode         folder.Mode
	gen          generator
	folderer     folderProvider
	materializer materializer
	navigator    navigator
	projectData  ProjectData
	opts         Options
	navNames     []string
	// TODO: inject logger
}

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
func WithGen(gen generator) option {
	return func(a *app) {
		a.gen = gen
	}
}

// WithNavigator sets the NavHandler for the application to manage navigation files.
func WithNavigator(n navigator) option {
	return func(a *app) {
		a.navigator = n
	}
}

// WithDirProvider sets the DirProvider for the application to provide project folder names.
func WithDirProvider(dp folderProvider) option {
	return func(a *app) {
		a.folderer = dp
	}
}

// WithOptions sets the application options.
func WithOptions(opts Options) option {
	return func(a *app) {
		a.opts = opts
	}
}

// WithProject sets the project manager for the application.
func WithProject(proj materializer) option {
	return func(a *app) {
		a.materializer = proj
	}
}

// WithMode sets whether to create a new project folder or reuse an existing one.
// Default is false (reuse existing).
func WithMode(mode folder.Mode) option {
	return func(a *app) {
		a.mode = mode
	}
}

// WithProjectData sets the project data for the application.
func WithProjectData(data ProjectData) option {
	return func(a *app) {
		a.projectData = data
	}
}

// WithMaterializer sets the materializer for the application to handle project materialization.
func WithMaterializer(m materializer) option {
	return func(a *app) {
		a.materializer = m
	}
}

// WithNavNames sets the names of the navigation files to be used in the application.
func WithNavNames(names []string) option {
	return func(a *app) {
		a.navNames = names
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
		deg:      1,
		navNames: defaultNavNames[:],
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

// ParseAppOptions parses the application options from a JSON file located
// in the specified path.
func ParseAppOptions(path string) (Options, error) {
	optsPath := filepath.Join(path, ConfigFile)
	opts, err := jsonx.LoadWithLocal[Options](optsPath)
	if err != nil {
		return Options{}, fmt.Errorf("parse ada options %q: %w", optsPath, err)
	}
	if opts.PromptsRoot == "" {
		opts.PromptsRoot = common.PromptsPath
	}
	if opts.ProjectsRoot == "" {
		opts.ProjectsRoot = common.ProjectsPath
	}
	return opts, nil
}
