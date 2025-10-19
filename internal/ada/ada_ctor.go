package ada

import "github.com/tmc/langchaingo/llms"

// Option is a function that configures Ada AI workflow.
type Option func(*Ada)

// WithTimeout sets the timeout duration for Ada AI requests.
func WithTimeout(timeout string) Option {
	return func(ada *Ada) {
		ada.Timeout = timeout
	}
}

// WithPromptsRoot sets the root directory for prompt templates.
func WithPromptsRoot(root string) Option {
	return func(ada *Ada) {
		ada.PromptsRoot = root
	}
}

// WithReflection sets the reflection configuration for Ada.
func WithReflection(reflect ReflectConfig) Option {
	return func(ada *Ada) {
		ada.Reflection = reflect
	}
}

// New creates a new Ada AI workflow instance with the provided LLM model
// and optional configurations. It initializes the Ada struct and applies any
// provided options to customize its behavior. The main LLM model is stored
// under the "main" key in the AIs map.
func New(ai llms.Model, opts ...Option) *Ada {
	ada := &Ada{
		ai:      ai,
		Timeout: "3m",
		Reflection: ReflectConfig{
			Depth:      3,
			Threshhold: 0.95,
		},
	}

	for _, o := range opts {
		o(ada)
	}

	return ada
}
