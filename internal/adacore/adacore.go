package adacore

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
)

const (
	reflectPrompt      = "reflect-template.txt"
	reflectShortPrompt = "reflect-template-short.txt"
	improvePrompt      = "improve-template.txt"
)

// Ada is the main struct for Ada AI workflow, encapsulating LLM models,
// configuration options, and project context.
type Ada struct {
	ai       llms.Model
	Opts     Options
	Projdata ProjectData
}

// Options is the configuration for Ada AI workflow.
type Options struct {
	Timeout      string        `json:"requestTimeout"`
	ProjectsRoot string        `json:"projectsRoot"`
	PromptsRoot  string        `json:"promptsRoot"`
	Reflection   ReflectConfig `json:"reflection"`
}

// Option is a function that configures Ada AI workflow.
type Option func(*Ada)

// ProjectData is the context of the current project.
type ProjectData struct {
	UserName string `json:"userName"`
	ProjName string `json:"projName"`
	Language string `json:"language"`
	Summary  string `json:"summary"`
}

var defaultOpts = Options{
	Timeout: "3m",
	Reflection: ReflectConfig{
		Depth:      3,
		Threshhold: 0.95,
	},
}

// WithOptions sets the entire options struct.
func WithOptions(opts Options) Option {
	return func(ada *Ada) {
		ada.Opts = opts
	}
}

// New creates a new Ada AI workflow instance with the provided LLM model
// and optional configurations. It initializes the Ada struct and applies any
// provided options to customize its behavior. The main LLM model is stored
// under the "main" key in the AIs map.
func New(ai llms.Model, opts ...Option) *Ada {
	ada := &Ada{
		ai:   ai,
		Opts: defaultOpts,
	}

	for _, o := range opts {
		o(ada)
	}

	return ada
}

// AddProjectData sets the current project data in the Ada session.
func (ada *Ada) AddProjectData(pd ProjectData) {
	if pd.UserName == "" {
		pd.UserName = "unknown_user"
	}
	if pd.ProjName == "" {
		pd.ProjName = fmt.Sprintf("project_%s", uuid.NewString())
	}
	ada.Projdata = pd
}

// ResolvePromptPath a path to the prompt from the prompts folder with
// given filename.
func (ada *Ada) ResolvePromptPath(filename string) string {
	return filepath.Join(ada.Opts.PromptsRoot, filename)
}

// ResolveProjectPath resolves a project root folder path, based using
// current project and user context.
func (ada *Ada) ResolveProjectPath() string {
	return filepath.Join(ada.Opts.ProjectsRoot, ada.Projdata.UserName, ada.Projdata.ProjName)
}

// GenerateWithSys sends a prompt with a system message to the LLM and expects a response.
// It calls GenerateContent with the provided system prompt.
func (ada *Ada) GenerateWithSys(ctx context.Context, sys, user string, callOptions ...llms.CallOption) (*llms.ContentResponse, error) {
	msgs := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeSystem, sys)}
	return ada.GenerateContent(ctx, user, msgs, callOptions...)
}

// GenerateContent sends a prompt along with a series of messages to the LLM
// and expects a response. It uses the options and timeout specified in the
// Ada configuration. If the timeout is invalid, it defaults to 3 minutes.
func (ada *Ada) GenerateContent(ctx context.Context, prompt string, msgs []llms.MessageContent, callOptions ...llms.CallOption) (*llms.ContentResponse, error) {
	dur, err := time.ParseDuration(ada.Opts.Timeout)
	if err != nil {
		dur = time.Minute * 3
		slog.Warn("failed to parse duration from config: using default", "duration", ada.Opts.Timeout, "default", dur)
	}

	msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, prompt))
	ctx, cancel := context.WithTimeout(ctx, dur)
	res, err := ada.ai.GenerateContent(ctx, msgs, callOptions...)
	cancel()
	return res, err
}

// PromptFromTemplate reads a prompt template file and formats it with the provided input.
// It returns the formatted prompt string or an error if the file cannot be read.
func (ada *Ada) PromptFromTemplate(filename string, input ...any) (string, error) {
	file := ada.ResolvePromptPath(filename)
	template, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("read template: %w", err)
	}

	return fmt.Sprintf(string(template), input...), nil
}
