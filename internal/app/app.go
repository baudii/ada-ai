package app

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/baudii/ada-ai/internal/adacore"
	"github.com/baudii/ada-ai/internal/ai"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/pkg/utils"
)

const fold = "scheme"

var (
	busn = prompt{
		system: filepath.Join("business", "system.txt"),
		human:  filepath.Join("business", "human.txt"),
	}
	tech = prompt{
		system: filepath.Join("technical", "system.txt"),
		human:  filepath.Join("technical", "human.txt"),
	}
)

type prompt struct {
	system string
	human  string
	//reflect string
}

var defaultCfg adacore.Options = adacore.Options{
	Timeout: "3m",
}

type Runner interface {
	Projdata(string, chan adacore.ProjectData)
}

// Run initializes and runs the CLI application. It sets up the AI model,
// configures the Ada AI workflow, and handles user input to generate and
// materialize a project based on the provided description.
//
// It also manages configuration loading and error handling throughout the process.
func Run(runner Runner) error {
	udpath := filepath.Join(common.DataPath, "user_data.json")
	ada, err := initAda(runner, udpath)
	if err != nil {
		return fmt.Errorf("initialize ada: %w", err)
	}

	res, err := getDescription(ada, "business-description.json", busn, stringify(ada.Projdata))
	if err != nil {
		return fmt.Errorf("project description: %w", err)
	}

	res, err = getDescription(ada, "technical-description.json", tech, res)
	if err != nil {
		return fmt.Errorf("technical description: %w", err)
	}

	slog.Info(res, "type", "technical description")
	return nil
}

func initAda(runner Runner, udpath string) (*adacore.Ada, error) {
	ai, err := ai.RegisterFromFile(common.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("register llm: %w", err)
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
	go runner.Projdata(udpath, c)
	data := <-c
	ada.AddProjectData(data)
	slog.Info("starting app session", "user", ada.Projdata.UserName, "project", ada.Projdata.ProjName)
	return ada, nil
}

func getDescription(ada *adacore.Ada, name string, p prompt, args ...any) (string, error) {
	projpath := ada.ResolveProjectPath()
	path := filepath.Join(projpath, fold, name)
	_, err := os.Stat(path)
	if err == nil {
		slog.Info("project description already exists, skipping generation", "path", path)
		res, err := os.ReadFile(path)
		if err != nil {
			slog.Error("failed to unmarshal existing project description", "error", err)
		} else {
			return string(res), nil
		}
	}

	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("create project path: %w", err)
	}

	sys, err := ada.PromptFromTemplate(p.system)
	if err != nil {
		return "", fmt.Errorf("load prod system template: %w", err)
	}

	hum, err := ada.PromptFromTemplate(p.human, args...)
	if err != nil {
		return "", fmt.Errorf("load prod human template: %w", err)
	}

	resp, err := ada.GenerateWithSys(sys, hum)
	if err != nil {
		return "", fmt.Errorf("generate with sys: %w", err)
	}

	err = os.WriteFile(path, []byte(resp.Choices[0].Content), 0644)
	if err != nil {
		return "", fmt.Errorf("write project description: %w", err)
	}

	return resp.Choices[0].Content, nil
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
