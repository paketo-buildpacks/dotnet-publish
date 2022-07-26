package dotnetpublish

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver"
	"github.com/mattn/go-shellwords"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/paketo-buildpacks/packit/v2/servicebindings"
)

//go:generate faux --interface SymlinkManager --output fakes/symlink_manager.go
type SymlinkManager interface {
	Link(oldname, newname string) error
	Unlink(path string) error
}

//go:generate faux --interface SourceRemover --output fakes/source_remover.go
type SourceRemover interface {
	Remove(workingDir, publishOutputDir string, excludedFiles ...string) error
}

//go:generate faux --interface PublishProcess --output fakes/publish_process.go
type PublishProcess interface {
	Execute(workingDir, rootDir, nugetCachePath, projectPath, outputPath string, flags []string) error
}

//go:generate faux --interface BindingResolver --output fakes/binding_resolver.go
type BindingResolver interface {
	Resolve(typ, provider, platformDir string) ([]servicebindings.Binding, error)
}

//go:generate faux --interface Slicer --output fakes/slicer.go
type Slicer interface {
	Slice(assetsFile string) (pkgs, earlyPkgs, projects packit.Slice, err error)
}

type Configuration struct {
	DisableOutputSlicing bool   `env:"BP_DOTNET_DISABLE_BUILDPACK_OUTPUT_SLICING"`
	ProjectPath          string `env:"BP_DOTNET_PROJECT_PATH"`
	PublishFlags         []string
	RawPublishFlags      string `env:"BP_DOTNET_PUBLISH_FLAGS"`
}

func Build(
	config Configuration,
	sourceRemover SourceRemover,
	bindingResolver BindingResolver,
	homeDir string,
	symlinker SymlinkManager,
	publishProcess PublishProcess,
	slicer Slicer,
	buildpackYMLParser BuildpackYMLParser,
	clock chronos.Clock,
	logger scribe.Emitter,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)
		if config.ProjectPath == "" {
			var err error
			config.ProjectPath, err = buildpackYMLParser.ParseProjectPath(filepath.Join(context.WorkingDir, "buildpack.yml"))
			if err != nil {
				return packit.BuildResult{}, err
			}

			if config.ProjectPath != "" {
				nextMajorVersion := semver.MustParse(context.BuildpackInfo.Version).IncMajor()
				logger.Subprocess("WARNING: Setting the project path through buildpack.yml will be deprecated soon in Dotnet Publish Buildpack v%s", nextMajorVersion.String())
				logger.Subprocess("Please specify the project path through the $BP_DOTNET_PROJECT_PATH environment variable instead. See README.md or the documentation on paketo.io for more information.")
			}
		}

		tempDir, err := os.MkdirTemp("", "dotnet-publish-output")
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("could not create temp directory: %w", err)
		}

		shellwordsParser := shellwords.NewParser()
		shellwordsParser.ParseEnv = true

		config.PublishFlags, err = shellwordsParser.Parse(config.RawPublishFlags)
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to parse flags for dotnet publish: %w", err)
		}

		globalNugetPath, err := getBinding("nugetconfig", "", context.Platform.Path, "nuget.config", bindingResolver, logger)
		if err != nil {
			return packit.BuildResult{}, err
		}

		if globalNugetPath != "" {
			err = symlinker.Link(globalNugetPath, filepath.Join(homeDir, ".nuget", "NuGet", "NuGet.Config"))
			if err != nil {
				return packit.BuildResult{}, err
			}
		}

		nugetCache, err := context.Layers.Get("nuget-cache")
		if err != nil {
			return packit.BuildResult{}, err
		}

		nugetCache.Cache = true

		logger.Process("Executing build process")
		err = publishProcess.Execute(context.WorkingDir, os.Getenv("DOTNET_ROOT"), nugetCache.Path, config.ProjectPath, tempDir, config.PublishFlags)
		if err != nil {
			return packit.BuildResult{}, err
		}

		slices := []packit.Slice{
			{Paths: []string{".dotnet_root"}},
		}

		if !config.DisableOutputSlicing {
			logger.Process("Dividing build output into layers to optimize cache reuse")

			pkg, early, project, err := slicer.Slice(filepath.Join(context.WorkingDir, config.ProjectPath, "obj", "project.assets.json"))
			if err != nil {
				return packit.BuildResult{}, err
			}

			for _, slice := range []packit.Slice{pkg, early, project} {
				if len(slice.Paths) > 0 {
					slices = append(slices, slice)
				}
			}
			logger.Break()
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

		if globalNugetPath != "" {
			err = symlinker.Unlink(filepath.Join(homeDir, ".nuget", "NuGet", "NuGet.Config"))
			if err != nil {
				return packit.BuildResult{}, err
			}
		}

		var layers []packit.Layer
		exists, err := fs.Exists(nugetCache.Path)
		if exists {
			if !fs.IsEmptyDir(nugetCache.Path) {
				layers = append(layers, nugetCache)
			}
		}
		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{
			Layers: layers,
			Launch: packit.LaunchMetadata{
				Slices: slices,
			},
		}, nil
	}
}

func getBinding(typ, provider, bindingsRoot, entry string, bindingResolver BindingResolver, logger scribe.Emitter) (string, error) {
	bindings, err := bindingResolver.Resolve(typ, provider, bindingsRoot)
	if err != nil {
		return "", err
	}

	if len(bindings) > 1 {
		return "", errors.New("binding resolver found more than one binding of type 'nugetconfig'")
	}

	if len(bindings) == 1 {
		logger.Process("Loading nuget service binding")

		if _, ok := bindings[0].Entries[entry]; !ok {
			return "", fmt.Errorf("binding of type %s does not contain required entry %s", typ, entry)
		}
		return filepath.Join(bindings[0].Path, entry), nil
	}
	return "", nil
}
