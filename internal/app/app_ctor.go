package app

import (
	"log/slog"
)

// ConfigFile is the default application configuration filename.
const ConfigFile = "ada.json"

type App struct {
	parallelism int
	generator   Generator
	options     Options
	logger      *slog.Logger
}

// Options is the configuration for Ada AI workflow.
type Options struct {
	Timeout      string `json:"requestTimeout"`
	ProjectsRoot string `json:"projectsRoot"`
	PromptsRoot  string `json:"promptsRoot"`
}

type option func(*App)

// WithGenerator sets the Generator instance for the application to generate content.
func WithGenerator(gen Generator) option {
	return func(a *App) {
		a.generator = gen
	}
}

// WithOptions sets the application options.
func WithOptions(opts Options) option {
	return func(a *App) {
		a.options = opts
	}
}

// WithDegree sets the maximum degree of concurrency for the application.
func WithDegree(deg int) option {
	return func(a *App) {
		a.parallelism = deg
	}
}

// WithLogger sets the logger for the application.
func WithLogger(logger *slog.Logger) option {
	return func(a *App) {
		a.logger = logger
	}
}

// New creates a new application instance with the provided options.
func New(opts ...option) *App {
	a := &App{
		parallelism: 1,
	}
	for _, o := range opts {
		o(a)
	}
	return a
}
