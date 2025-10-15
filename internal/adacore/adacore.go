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

type Ada struct {
	ai       llms.Model
	Opts     Options
	Projdata ProjectData
	Proj     project
}

// Options is the configuration for Ada AI workflow.
type Options struct {
	Timeout      string           `json:"requestTimeout"`
	ProjectsRoot string           `json:"projectsRoot"`
	PromptsRoot  string           `json:"promptsRoot"`
	Reflection   ReflectConfig    `json:"reflection"`
	ModelCall    llms.CallOptions `json:"call-options"`
}

// SessionOption is a function that modifies the Ada session configuration.
type Option func(Options) Options

// ProjectData is the context of the current project.
type ProjectData struct {
	UserName string `json:"userName"`
	ProjName string `json:"projName"`
	Language string `json:"language"`
	Summary  string `json:"summary"`
}

type project interface {
	Materialize() error
	Structure() map[string]any
}

var options = Options{
	Timeout: "3m",
	Reflection: ReflectConfig{
		Depth:      3,
		Threshhold: 0.95,
	},
}

// WithOptions sets the entire options struct.
func WithOptions(opts Options) Option {
	return func(o Options) Options {
		return opts
	}
}

// New creates a new Ada AI workflow instance with the provided LLM model and options.
func New(ai llms.Model, opts ...Option) *Ada {
	for _, o := range opts {
		options = o(options)
	}

	return &Ada{ai: ai, Opts: options}
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
func (ada *Ada) GenerateWithSys(sys string, user string) (*llms.ContentResponse, error) {
	msgs := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeSystem, sys)}
	return ada.GenerateContent(user, msgs)
}

// GenerateContent sends a prompt along with a series of messages to the LLM
// and expects a response. It uses the options and timeout specified in the
// Ada configuration. If the timeout is invalid, it defaults to 3 minutes.
func (ada *Ada) GenerateContent(prompt string, msgs []llms.MessageContent) (*llms.ContentResponse, error) {
	dur, err := time.ParseDuration(ada.Opts.Timeout)
	if err != nil {
		dur = time.Minute * 3
		slog.Warn("failed to parse duration from config: using default", "duration", ada.Opts.Timeout, "default", dur)
	}

	msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, prompt))
	ctx, cf := context.WithTimeout(context.Background(), dur)
	res, err := ada.ai.GenerateContent(ctx, msgs, llms.WithOptions(ada.Opts.ModelCall))
	cf()
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
