package dotnetpublish

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/scribe"
)

//go:generate faux --interface SourceRemover --output fakes/source_remover.go
type SourceRemover interface {
	Remove(workingDir, publishOutputDir string, excludedFiles ...string) error
}

//go:generate faux --interface PublishProcess --output fakes/publish_process.go
type PublishProcess interface {
	Execute(workingDir, rootDir, projectPath, outputPath string, flags []string) error
}

//go:generate faux --interface CommandConfigParser --output fakes/command_config_parser.go
type CommandConfigParser interface {
	ParseFlagsFromEnvVar(envVar string) ([]string, error)
}

func Build(
	sourceRemover SourceRemover,
	publishProcess PublishProcess,
	buildpackYMLParser BuildpackYMLParser,
	configParser CommandConfigParser,
	clock chronos.Clock,
	logger scribe.Logger,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)
		var projectPath string
		var ok bool
		var err error

		if projectPath, ok = os.LookupEnv("BP_DOTNET_PROJECT_PATH"); !ok {
			projectPath, err = buildpackYMLParser.ParseProjectPath(filepath.Join(context.WorkingDir, "buildpack.yml"))
			if err != nil {
				return packit.BuildResult{}, err
			}

			if projectPath != "" {
				nextMajorVersion := semver.MustParse(context.BuildpackInfo.Version).IncMajor()
				logger.Subprocess("WARNING: Setting the project path through buildpack.yml will be deprecated soon in Dotnet Publish Buildpack v%s", nextMajorVersion.String())
				logger.Subprocess("Please specify the project path through the $BP_DOTNET_PROJECT_PATH environment variable instead. See README.md or the documentation on paketo.io for more information.")
			}
		}

		tempDir, err := ioutil.TempDir("", "dotnet-publish-output")
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("could not create temp directory: %w", err)
		}

		flags, err := configParser.ParseFlagsFromEnvVar("BP_DOTNET_PUBLISH_FLAGS")
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Process("Executing build process")
		err = publishProcess.Execute(context.WorkingDir, os.Getenv("DOTNET_ROOT"), projectPath, tempDir, flags)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Process("Removing source code")
		logger.Break()
		err = sourceRemover.Remove(context.WorkingDir, tempDir, ".dotnet_root")
		if err != nil {
			return packit.BuildResult{}, err
		}

		err = os.RemoveAll(tempDir)
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("could not remove temp directory: %w", err)
		}

		return packit.BuildResult{}, nil
	}
}
