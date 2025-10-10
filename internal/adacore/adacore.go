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
	ProjectStructurePrompt = "project-structure-template.txt"

	reflectPrompt      = "reflect-template.txt"
	reflectShortPrompt = "reflect-template-short.txt"
	improvePrompt      = "improve-template.txt"
)

type ada struct {
	ai      llms.Model
	Session session
	Proj    project
}

// Config is the configuration for Ada AI workflow.
type Config struct {
	Timeout    string        `json:"requestTimeout"`
	Reflection ReflectConfig `json:"reflection"`
	CallOpts   CallOptConfig `json:"call-options"`
}

// ProjectData is the context of the current project.
type ProjectData struct {
	UserName string `json:"userName"`
	ProjName string `json:"projName"`
	// TODO: Maybe add additional project context like
	// tech stack, architecture, project summary etc.
}

type CallOptConfig struct {
	Temperature float64 `json:"temperature"`
	JSONMode    bool    `json:"jsonMode"`
}

type project interface {
	Materialize() error
	Structure() map[string]any
}

// SessionOptions is a set of options that can be used to configure
// the session. It allows to set the root folders for projects and prompts,
// as well as the configuration for the session.
type SessionOption func(*session)

type session struct {
	Cfg          *Config
	Project      ProjectData
	projectsRoot string
	promptsRoot  string
	callOpts     []llms.CallOption
}

var defaultCfg = Config{
	Timeout: "3m",
	Reflection: ReflectConfig{
		Depth:      3,
		Threshhold: 0.95,
	},
}

// WithProjectsRoot sets the root folder that will be used when resolving
// project-specific directories during the session lifecycle.
func WithProjectsRoot(path string) SessionOption {
	return func(s *session) {
		s.projectsRoot = path
	}
}

// WithPromptsRoot sets the root folder of where the ada should search
// for prompt templates.
func WithPromptsRoot(path string) SessionOption {
	return func(s *session) {
		s.promptsRoot = path
	}
}

// WithConfig sets the configuration for current session.
func WithConfig(cfg *Config) SessionOption {
	return func(s *session) {
		s.Cfg = cfg
		if cfg != nil {
			if cfg.CallOpts.Temperature != 0 {
				s.callOpts = append(s.callOpts, llms.WithTemperature(cfg.CallOpts.Temperature))
			}
			if cfg.CallOpts.JSONMode {
				s.callOpts = append(s.callOpts, llms.WithJSONMode())
			}
		}
	}
}

// WithProjectData sets the project data in the current session.
func WithProjectData(p ProjectData) SessionOption {
	return func(s *session) {
		s.Project = p
	}
}

// New initializes a new Ada instance with the provided LLM model and session
// options. It applies defaults for user, project, and configuration values when
// they are not supplied through the provided options.
func New(ai llms.Model, opts ...SessionOption) *ada {
	session := &session{}
	for _, v := range opts {
		v(session)
	}

	if session.Project.UserName == "" {
		session.Project.UserName = "unknown_user"
	}
	if session.Project.ProjName == "" {
		session.Project.ProjName = fmt.Sprintf("project_%s", uuid.NewString())
	}
	if session.Cfg == nil {
		session.Cfg = &defaultCfg
	}

	return &ada{ai: ai, Session: *session}
}

// ResolvePromptPath a path to the prompt from the prompts folder with
// given filename.
func (ada *ada) ResolvePromptPath(filename string) string {
	return filepath.Join(ada.Session.promptsRoot, filename)
}

// ResolveProjectPath resolves a project root folder path, based using
// current project and user context.
func (ada *ada) ResolveProjectPath() string {
	return filepath.Join(ada.Session.projectsRoot, ada.Session.Project.UserName, ada.Session.Project.ProjName)
}

// GenerateContent sends a prompt along with a series of messages to the LLM
// and expects a response. It uses the options and timeout specified in the
// Ada configuration. If the timeout is invalid, it defaults to 3 minutes.
func (ada *ada) GenerateContent(prompt string, msgs []llms.MessageContent) (*llms.ContentResponse, error) {
	dur, err := time.ParseDuration(ada.Session.Cfg.Timeout)
	if err != nil {
		dur = time.Minute * 3
		slog.Warn("failed to parse duration from config: using default", "duration", ada.Session.Cfg.Timeout, "default", dur)
	}

	msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeSystem, prompt))
	ctx, cf := context.WithTimeout(context.Background(), dur)
	res, err := ada.ai.GenerateContent(ctx, msgs, ada.Session.callOpts...)
	cf()
	return res, err
}

// PromptFromTemplate reads a prompt template file and formats it with the provided input.
// It returns the formatted prompt string or an error if the file cannot be read.
func (ada *ada) PromptFromTemplate(filename string, input ...any) (string, error) {
	file := ada.ResolvePromptPath(filename)
	template, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("read template: %w", err)
	}

	return fmt.Sprintf(string(template), input...), nil
}
