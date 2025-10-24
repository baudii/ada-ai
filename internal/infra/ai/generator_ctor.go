package ai

import "github.com/tmc/langchaingo/llms"

// Option is a function that configures Ada AI workflow.
type Option func(*generator)

// WithTimeout sets the timeout duration for Ada AI requests.
func WithTimeout(timeout string) Option {
	return func(ada *generator) {
		ada.Timeout = timeout
	}
}

// WithPromptsRoot sets the root directory for prompt templates.
func WithPromptsRoot(root string) Option {
	return func(ada *generator) {
		ada.PromptsRoot = root
	}
}

// New creates a new Ada AI workflow instance with the provided LLM model
// and optional configurations. It initializes the Ada struct and applies any
// provided options to customize its behavior. The main LLM model is stored
// under the "main" key in the AIs map.
func NewGenerator(ai llms.Model, opts ...Option) *generator {
	ada := &generator{
		ai:      ai,
		Timeout: "3m",
	}

	for _, o := range opts {
		o(ada)
	}

	return ada
}
