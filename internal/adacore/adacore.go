package adacore

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/tmc/langchaingo/llms"
)

const (
	ProjectStructurePrompt = "project-structure-template.txt"

	reflectPrompt      = "reflect-template.txt"
	reflectShortPrompt = "reflect-template-short.txt"
	improvePrompt      = "improve-template.txt"

	promptsFolder        = "prompts"
	ProjectStrucutreFile = "project-structure.json" // TODO: fix typo
)

var ProjectDataFile = "user_data.json"

type Ada struct {
	ai      llms.Model
	Cfg     *Config
	Session *session
	Proj    Project
}

type Config struct {
	Timeout    string        `json:"requestTimeout"`
	ProjRoot   string        `json:"projRoot"`
	Reflection ReflectConfig `json:"reflection"`
}

type ProjectData struct {
	UserName string `json:"userName"`
	ProjName string `json:"projName"`
	// TODO: Maybe add additional project context like
	// tech stack, architecture, project summary etc.
}

type SessionOption func(*session)

type Project interface {
	Materialize() error
	Structure() map[string]any
}

type session struct {
	root    string
	Project ProjectData
}

var defaultCfg = Config{
	ProjRoot: ".projects",
	Timeout:  "3m",
	Reflection: ReflectConfig{
		Depth:      3,
		Threshhold: 0.95,
	},
}

// WithRoot sets the root path. It is used to resolve all paths during application
// execution including searching prompts, configurations and saving states and
// artifacts.
func WithRoot(path string) SessionOption {
	return func(s *session) {
		s.root = path
	}
}

// WithProjectData sets the project data in the current session.
func WithProjectData(p ProjectData) SessionOption {
	return func(s *session) {
		s.Project = p
	}
}

// New initializes a new Ada instance with the provided LLM model and configuration.
// If a user context is found in the user data file, it is loaded and associated with
// the Ada instance. Otherwise, the context remains nil and expected to be set later
// using AddProjCtx.
func New(ai llms.Model, config *Config, opts ...SessionOption) *Ada {
	if config == nil {
		config = &defaultCfg
	}
	session := &session{}
	for _, v := range opts {
		v(session)
	}
	return &Ada{ai: ai, Cfg: config, Session: session}
}

// ResolvePromptPath a path to the prompt from the prompts folder with
// given filename.
func (ada *Ada) ResolvePromptPath(filename string) string {
	return filepath.Join(ada.Session.root, promptsFolder, filename)
}

// ResolveProjectPath resolves a project root folder path, based using
// current project and user context.
func (ada *Ada) ResolveProjectPath() string {
	return filepath.Join(ada.Session.root, ada.Cfg.ProjRoot, ada.Session.Project.UserName, ada.Session.Project.ProjName)
}

// GenerateJSON sends a prompt along with a series of messages to the LLM
// and expects a JSON response. It uses the timeout specified in the Ada configuration.
// If the timeout is invalid, it defaults to 3 minutes.
func (ada *Ada) GenerateJSON(prompt string, msgs []llms.MessageContent) (*llms.ContentResponse, error) {
	dur, err := time.ParseDuration(ada.Cfg.Timeout)
	if err != nil {
		dur = time.Minute * 3
		slog.Warn("failed to parse duration from config: using default", "duration", ada.Cfg.Timeout, "default", dur)
	}

	msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeSystem, prompt))
	ctx, cf := context.WithTimeout(context.Background(), dur)
	res, err := ada.ai.GenerateContent(ctx, msgs, llms.WithJSONMode())
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
