package dotnetpublish

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type DotnetBuildpackYMLParser struct{}

func NewDotnetBuildpackYMLParser() BuildpackYMLParser {
	return DotnetBuildpackYMLParser{}
}

func (p DotnetBuildpackYMLParser) ParseProjectPath(path string) (string, error) {
	var buildpack struct {
		Config struct {
			ProjectPath string `yaml:"project-path"`
		} `yaml:"dotnet-build"`
	}

	file, err := os.Open(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	defer file.Close()

	if !os.IsNotExist(err) {
		err = yaml.NewDecoder(file).Decode(&buildpack)
		if err != nil {
			return "", fmt.Errorf("invalid buildpack.yml: %w", err)
		}
	}

	return buildpack.Config.ProjectPath, nil
}
