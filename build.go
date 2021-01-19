package dotnetpublish

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/fs"
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

		tempDir, err := ioutil.TempDir("", "dotnet-publish-output")
		if err != nil {
			panic(err)
		}

		logger.Process("Executing build process")
		err = publishProcess.Execute(context.WorkingDir, rootDir, projectPath, tempDir)
		if err != nil {
			return packit.BuildResult{}, err
		}

		// fs.Move(rootDir, filepath.Join(tempDir, filepath.Base(rootDir)))

		workspaceFiles, err := filepath.Glob(filepath.Join(context.WorkingDir, "*"))
		if err != nil {
			panic(err)
		}

		for _, file := range workspaceFiles {
			if filepath.Base(file) != filepath.Base(rootDir) && filepath.Base(file) != ".dotnet_root" {
				err = os.RemoveAll(file)
				if err != nil {
					panic(err)
				}
			}
		}

		generatedFiles, err := filepath.Glob(filepath.Join(tempDir, "*"))
		if err != nil {
			panic(err)
		}
		for _, file := range generatedFiles {
			fs.Move(file, filepath.Join(context.WorkingDir, filepath.Base(file)))
		}

		err = os.RemoveAll(tempDir)
		if err != nil {
			panic(err)
		}

		return packit.BuildResult{}, nil
	}
}
