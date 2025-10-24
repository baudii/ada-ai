package app

import (
	"log/slog"

	"github.com/baudii/ada-ai/internal/core/project"
	"github.com/baudii/ada-ai/internal/core/seqdir"
)

// ConfigFile is the default application configuration filename.
const ConfigFile = "ada.json"

var defaultNavNames = []string{business, technical, scope, project.Structure}

type app struct {
	deg          int
	mode         seqdir.Mode
	gen          generator
	materializer materializer
	navigator    navigator
	projectData  project.Context
	opts         Options
	navNames     []string
	logger       *slog.Logger
}

// Options is the configuration for Ada AI workflow.
type Options struct {
	Timeout      string `json:"requestTimeout"`
	ProjectsRoot string `json:"projectsRoot"`
	PromptsRoot  string `json:"promptsRoot"`
}

type option func(*app)

// WithGenerator sets the Generator instance for the application to generate content.
func WithGenerator(gen generator) option {
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
func WithMode(mode seqdir.Mode) option {
	return func(a *app) {
		a.mode = mode
	}
}

// WithProjectData sets the project data for the application.
func WithProjectData(data project.Context) option {
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

// WithLogger sets the logger for the application.
func WithLogger(logger *slog.Logger) option {
	return func(a *app) {
		a.logger = logger
	}
}

// New creates a new application instance with the provided options.
func New(opts ...option) *app {
	a := &app{
		deg:      1,
		navNames: defaultNavNames,
	}
	for _, o := range opts {
		o(a)
	}
	return a
}
