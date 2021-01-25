package dotnetpublish

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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
	Execute(workingDir, rootDir, projectPath, outputPath string) error
}

func Build(
	sourceRemover SourceRemover,
	publishProcess PublishProcess,
	buildpackYMLParser BuildpackYMLParser,
	clock chronos.Clock,
	logger scribe.Logger,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		projectPath, err := buildpackYMLParser.ParseProjectPath(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			return packit.BuildResult{}, err
		}

		tempDir, err := ioutil.TempDir("", "dotnet-publish-output")
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("could not create temp directory: %w", err)
		}

		logger.Process("Executing build process")
		err = publishProcess.Execute(context.WorkingDir, os.Getenv("DOTNET_ROOT"), projectPath, tempDir)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Process("Removing source code")
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
