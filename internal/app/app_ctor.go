package app

import (
	"log/slog"

	"github.com/baudii/ada-ai/internal/core/project"
)

// ConfigFile is the default application configuration filename.
const ConfigFile = "ada.json"

type app struct {
	parallelism    int
	generator      generator
	materializer   materializer
	projectContext project.Context
	options        Options
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
		a.generator = gen
	}
}

// WithOptions sets the application options.
func WithOptions(opts Options) option {
	return func(a *app) {
		a.options = opts
	}
}

func WithMaterializer(m materializer) option {
	return func(a *app) {
		a.materializer = m
	}
}

// WithProjectData sets the project data for the application.
func WithProjectData(data project.Context) option {
	return func(a *app) {
		a.projectContext = data
	}
}

// WithDegree sets the maximum degree of concurrency for the application.
func WithDegree(deg int) option {
	return func(a *app) {
		a.parallelism = deg
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
		parallelism: 1,
	}
	for _, o := range opts {
		o(a)
	}
	return a
}
