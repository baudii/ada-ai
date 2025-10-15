package app

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/baudii/ada-ai/internal/adacore"
	"github.com/baudii/ada-ai/internal/ai"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/pkg/utils"
)

var (
	prod = prompt{
		system: filepath.Join("product", "system.txt"),
		human:  filepath.Join("product", "human.txt"),
	}
)

type prompt struct {
	system  string
	human   string
	reflect string
}

var defaultCfg adacore.Options = adacore.Options{
	Timeout: "3m",
}

type Runner interface {
	Projdata(chan adacore.ProjectData)
}

// Run initializes and runs the CLI application. It sets up the AI model,
// configures the Ada AI workflow, and handles user input to generate and
// materialize a project based on the provided description.
//
// It also manages configuration loading and error handling throughout the process.
func Run(runner Runner) error {
	ai, err := ai.RegisterFromFile(common.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to register llm", "error", err)
	}

	var opts *adacore.Options
	cfgPath := filepath.Join(common.ConfigPath, "ada.json")
	opts, err = utils.ParseJSONConfigWithLocal[adacore.Options](cfgPath)
	if err != nil {
		slog.Error("failed to parse json configuration", "path", cfgPath, "error", err)
		opts = &defaultCfg
	} else {
		slog.Info("succesffully parsed json configuration", "cfg", opts)
	}

	if opts.PromptsRoot == "" {
		opts.PromptsRoot = common.PromptsPath
	}
	if opts.ProjectsRoot == "" {
		opts.ProjectsRoot = common.ProjectsPath
	}

	ada := adacore.New(ai, adacore.WithOptions(*opts))
	c := make(chan adacore.ProjectData)
	go runner.Projdata(c)
	data := <-c
	ada.AddProjectData(data)
	slog.Info("starting app session", "user", ada.Project.UserName, "project", ada.Project.ProjName)
	sys, err := ada.PromptFromTemplate(prod.system)
	if err != nil {
		return fmt.Errorf("load prod system template: %w", err)
	}

	hum, err := ada.PromptFromTemplate(prod.human, stringify(data))
	if err != nil {
		return fmt.Errorf("load prod human template: %w", err)
	}

	resp, err := ada.GenerateWithSys(sys, hum)
	if err != nil {
		return fmt.Errorf("generate with sys: %w", err)
	}

	slog.Info(resp.Choices[0].Content)
	return nil
}

func stringify(v adacore.ProjectData) string {
	val := reflect.ValueOf(v)
	typ := val.Type()

	var b strings.Builder
	for i := 0; i < val.NumField(); i++ {
		if i > 0 {
			b.WriteByte('\n')
		}
		fieldName := typ.Field(i).Name
		fieldValue := val.Field(i).Interface()
		if fieldValue != "" {
			fmt.Fprintf(&b, "- %s: %v", fieldName, fieldValue)
		}
	}
	return b.String()
}
