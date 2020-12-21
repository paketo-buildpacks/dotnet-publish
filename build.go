package dotnetpublish

import (
	"os"
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/scribe"
)

//go:generate faux --interface RootManager --output fakes/root_manager.go
type RootManager interface {
	Setup(root, existingRoot, sdkLocation string) error
}

//go:generate faux --interface PublishProcess --output fakes/publish_process.go
type PublishProcess interface {
	Execute(workingDir, rootDir, projectPath, outputPath string) error
}

func Build(
	rootManager RootManager,
	publishProcess PublishProcess,
	buildpackYMLParser BuildpackYMLParser,
	clock chronos.Clock,
	logger scribe.Logger,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		publishOutputLayer, err := context.Layers.Get("publish-output")
		if err != nil {
			return packit.BuildResult{}, err
		}

		publishOutputLayer.Launch = true
		publishOutputLayer.Build = true

		publishOutputLayer.BuildEnv.Override("PUBLISH_OUTPUT_LOCATION", publishOutputLayer.Path)
		logger.Process("Configuring environment")
		logger.Subprocess("%s", scribe.NewFormattedMapFromEnvironment(publishOutputLayer.BuildEnv))
		logger.Break()

		rootDir := filepath.Join(context.WorkingDir, ".dotnet-root")
		err = rootManager.Setup(rootDir, os.Getenv("DOTNET_ROOT"), os.Getenv("SDK_LOCATION"))
		if err != nil {
			return packit.BuildResult{}, err
		}

		var projectPath string
		projectPath, err = buildpackYMLParser.ParseProjectPath(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Process("Executing build process")
		err = publishProcess.Execute(context.WorkingDir, rootDir, projectPath, publishOutputLayer.Path)
		if err != nil {
			return packit.BuildResult{}, err
		}

		publishOutputLayer.Metadata = map[string]interface{}{
			"built_at": clock.Now().Format(time.RFC3339Nano),
		}

		return packit.BuildResult{
			Layers: []packit.Layer{publishOutputLayer},
		}, nil
	}
}
