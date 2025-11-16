package app

import (
	"log/slog"

	"github.com/baudii/ada-ai/internal/core/project"
)

// ConfigFile is the default application configuration filename.
const ConfigFile = "ada.json"

var defaultNavNames = []string{business, technical, scope, project.Structure}

type app struct {
	deg            int
	gen            generator
	projectContext project.Context
	opts           Options
	navNames       []string
	logger         *slog.Logger
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

// WithOptions sets the application options.
func WithOptions(opts Options) option {
	return func(a *app) {
		a.opts = opts
	}
}

// WithProjectData sets the project data for the application.
func WithProjectData(data project.Context) option {
	return func(a *app) {
		a.projectContext = data
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
