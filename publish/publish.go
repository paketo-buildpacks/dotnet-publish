package publish

import (
	"fmt"
	"github.com/cloudfoundry/dotnet-core-conf-cnb/utils"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"os"
	"path/filepath"
)

const Publish = "build"

type Runner interface {
	Run(bin, dir string, quiet bool, args ...string) error
}

type MetadataInterface interface {
	Identity() (name string, version string)
}

type Metadata struct {
	Name string
	Hash string
}

func (m Metadata) Identity() (name string, version string) {
	return m.Name, m.Hash
}

type Contributor struct {
	context         build.Build
	buildLayer      layers.Layer
	buildMetadata   MetadataInterface
	publishMetadata MetadataInterface
	runner          Runner
}

func NewContributor(context build.Build, runner Runner) (Contributor, bool, error) {
	_, wantDependency, err := context.Plans.GetShallowMerged(Publish)
	if err != nil {
		return Contributor{}, false, err
	}
	if !wantDependency {
		return Contributor{}, false, nil
	}

	return Contributor{
		context:    context,
		buildLayer: context.Layers.Layer("build"),
		runner:     runner,
	}, true, nil
}

func (c Contributor) Contribute() error {
	err := c.buildLayer.Contribute(c.buildMetadata, c.contributeBuildLayer, layers.Build)
	if err != nil {
		return err
	}

	err = c.buildLayer.Contribute(c.publishMetadata, c.contributePublish, layers.Build)
	if err != nil {
		return err
	}

	return nil
}

func (c Contributor) contributeBuildLayer(layer layers.Layer) error {
	layer.Logger.Body("Symlinking runtime libraries")
	pathToRuntime := os.Getenv("DOTNET_ROOT")
	runtimeFiles, err := filepath.Glob(filepath.Join(pathToRuntime, "shared", "*"))
	if err != nil {
		return err
	}
	for _, file := range runtimeFiles {
		if err := helper.CopySymlink(file, filepath.Join(layer.Root, "shared", filepath.Base(file))); err != nil {
			return err
		}
	}

	hostDir := filepath.Join(pathToRuntime, "host")
	if err := utils.CreateValidSymlink(hostDir, filepath.Join(layer.Root, filepath.Base(hostDir))); err != nil {
		return err
	}

	layer.Logger.Body("Moving dotnet driver from %s", pathToRuntime)
	if err := helper.CopyFile(filepath.Join(pathToRuntime, "dotnet"), filepath.Join(layer.Root, "dotnet")); err != nil {
		return err
	}

	sdkLocation := os.Getenv("SDK_LOCATION")
	layer.Logger.Body("Symlinking the SDK from %s", sdkLocation)
	if err := utils.CreateValidSymlink(filepath.Join(sdkLocation, "sdk"), filepath.Join(layer.Root, "sdk")); err != nil {
		return err
	}

	if err := os.Setenv("PATH", fmt.Sprintf("%s:%s", layer.Root, os.Getenv("PATH"))); err != nil {
		return err
	}

	return nil
}

func (c Contributor) contributePublish(layer layers.Layer) error {
	layer.Logger.Body("Publishing source code")
	args := []string{
		"publish",
		"-c", "Release",
		"-r", "ubuntu.18.04-x64",
		"--self-contained", "false",
		"-o", c.context.Application.Root,
	}
	return c.runner.Run("dotnet", c.context.Application.Root, false, args...)
}
