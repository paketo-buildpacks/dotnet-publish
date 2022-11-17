package dotnetpublish

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver"
	"github.com/Netflix/go-env"
	"github.com/mattn/go-shellwords"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/sbom"
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
	Execute(workingDir, rootDir, nugetCachePath, projectPath, outputPath string, debug bool, flags []string) error
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
	LogLevel             string `env:"BP_LOG_LEVEL"`
	DebugEnabled         bool   `env:"BP_DEBUG_ENABLED"`
	DisableOutputSlicing bool   `env:"BP_DOTNET_DISABLE_BUILDPACK_OUTPUT_SLICING"`
	ProjectPath          string `env:"BP_DOTNET_PROJECT_PATH"`
	PublishFlags         []string
	RawPublishFlags      string `env:"BP_DOTNET_PUBLISH_FLAGS"`
}

//go:generate faux --interface SBOMGenerator --output fakes/sbom_generator.go
type SBOMGenerator interface {
	Generate(dir string) (sbom.SBOM, error)
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
	sbomGenerator SBOMGenerator,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)
		logger.Debug.Process("Build configuration:")
		es, err := env.Marshal(&config)
		if err != nil {
			// not tested
			return packit.BuildResult{}, fmt.Errorf("parsing build configuration: %w", err)
		}
		for envVar := range es {
			logger.Debug.Subprocess("%s: %s", envVar, es[envVar])
		}
		logger.Debug.Break()

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
			logger.Debug.Process("Setting up NuGet.Config from service binding")
			err = symlinker.Link(globalNugetPath, filepath.Join(homeDir, ".nuget", "NuGet", "NuGet.Config"))
			if err != nil {
				return packit.BuildResult{}, err
			}
			logger.Debug.Break()
		}

		nugetCache, err := context.Layers.Get("nuget-cache")
		if err != nil {
			return packit.BuildResult{}, err
		}

		stack, ok := nugetCache.Metadata["stack"].(string)
		if ok && stack != context.Stack {
			nugetCache, err = nugetCache.Reset()
			if err != nil {
				return packit.BuildResult{}, err
			}
		}

		if nugetCache.Metadata == nil {
			nugetCache.Metadata = make(map[string]interface{})
		}
		nugetCache.Metadata["stack"] = context.Stack
		nugetCache.Cache = true

		logger.Process("Executing build process")
		err = publishProcess.Execute(context.WorkingDir, os.Getenv("DOTNET_ROOT"), nugetCache.Path, config.ProjectPath, tempDir, config.DebugEnabled, config.PublishFlags)
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
		} else {
			logger.Debug.Process("Skipping output slicing")
			logger.Debug.Break()
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

		sbomLayer, err := context.Layers.Get("publish")
		if err != nil {
			return packit.BuildResult{}, err
		}
		sbomLayer.Build = true

		logger.GeneratingSBOM(context.WorkingDir)

		var sbomContent sbom.SBOM
		duration, err := clock.Measure(func() error {
			sbomContent, err = sbomGenerator.Generate(context.WorkingDir)
			return err
		})
		if err != nil {
			return packit.BuildResult{}, err
		}
		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.FormattingSBOM(context.BuildpackInfo.SBOMFormats...)

		sbomLayer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		layers = append(layers, sbomLayer)

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

		for _, layer := range layers {
			logger.Debug.Process("Setting up layer '%s'", layer.Name)
			logger.Debug.Subprocess("Available at launch: %t", layer.Launch)
			logger.Debug.Subprocess("Available to other buildpacks: %t", layer.Build)
			logger.Debug.Subprocess("Cached for rebuilds: %t", layer.Cache)
			logger.Debug.Break()
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
