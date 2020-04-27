package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/paketo-buildpacks/dotnet-core-build/publish"
)

func main() {
	context, err := build.DefaultBuild()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default build context: %s", err)
		os.Exit(101)
	}

	code, err := runBuild(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)

}

func runBuild(context build.Build) (int, error) {
	context.Logger.Title(context.Buildpack)

	dotnetBuildContributor, willContribute, err := publish.NewContributor(context, command{})
	if err != nil {
		return context.Failure(102), err
	}

	if willContribute {
		if err := dotnetBuildContributor.Contribute(); err != nil {
			return context.Failure(103), err
		}
	}

	return context.Success()
}

type command struct{}

func (c command) Run(bin, dir string, quiet bool, args ...string) error {
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	if quiet {
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}
