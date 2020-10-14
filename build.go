package dotnetpublish

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

//go:generate faux --interface RootManager --output fakes/root_manager.go
type RootManager interface {
	Setup(root, existingRoot, sdkLocation string) error
}

//go:generate faux --interface PublishProcess --output fakes/publish_process.go
type PublishProcess interface {
	Execute(workingDir, rootDir, projectPath string) error
}

func Build(
	rootManager RootManager,
	publishProcess PublishProcess,
	buildpackYMLParser BuildpackYMLParser,
	logger scribe.Logger,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		rootDir := filepath.Join(context.WorkingDir, ".dotnet-root")
		err := rootManager.Setup(rootDir, os.Getenv("DOTNET_ROOT"), os.Getenv("SDK_LOCATION"))
		if err != nil {
			return packit.BuildResult{}, err
		}

		var projectPath string
		projectPath, err = buildpackYMLParser.ParseProjectPath(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Process("Executing build process")
		err = publishProcess.Execute(context.WorkingDir, rootDir, projectPath)
		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{}, nil
	}
}
