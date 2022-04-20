package dotnetpublish

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) error
}

type DotnetPublishProcess struct {
	executable Executable
	logger     scribe.Logger
	clock      chronos.Clock
}

func NewDotnetPublishProcess(executable Executable, logger scribe.Logger, clock chronos.Clock) DotnetPublishProcess {
	return DotnetPublishProcess{
		executable: executable,
		logger:     logger,
		clock:      clock,
	}
}

func (p DotnetPublishProcess) Execute(workingDir, root, nugetCachePath, intermediateBuildCachePath, projectPath, outputPath string, flags []string) error {
	err := loadBuildCache(workingDir, projectPath, intermediateBuildCachePath)
	if err != nil {
		return fmt.Errorf("failed to load build cache: %w", err)
	}

	args := []string{
		"publish",
		filepath.Join(workingDir, projectPath),
	}

	if !containsFlag(flags, "--configuration") && !containsFlag(flags, "-c") {
		args = append(args, "--configuration", "Release")
	}

	if !containsFlag(flags, "--runtime") && !containsFlag(flags, "-r") {
		args = append(args, "--runtime", "ubuntu.18.04-x64")
	}

	if !containsFlag(flags, "--self-contained") && !containsFlag(flags, "--no-self-contained") {
		args = append(args, "--self-contained", "false")
	}

	if !containsFlag(flags, "--output") && !containsFlag(flags, "-o") {
		args = append(args, "--output", outputPath)
	}

	args = append(args, flags...)

	p.logger.Subprocess("Running 'dotnet %s'", strings.Join(args, " "))

	duration, err := p.clock.Measure(func() error {
		return p.executable.Execute(pexec.Execution{
			Args:   args,
			Dir:    workingDir,
			Env:    append(os.Environ(), fmt.Sprintf("PATH=%s:%s", root, os.Getenv("PATH")), fmt.Sprintf("NUGET_PACKAGES=%s", nugetCachePath)),
			Stdout: p.logger.ActionWriter,
			Stderr: p.logger.ActionWriter,
		})
	})

	if err != nil {
		p.logger.Action("Failed after %s", duration)
		return fmt.Errorf("failed to execute 'dotnet publish': %w", err)
	}

	p.logger.Action("Completed in %s", duration)
	p.logger.Break()

	err = recreateBuildCache(workingDir, projectPath, intermediateBuildCachePath)
	if err != nil {
		return err
	}

	return nil
}

func loadBuildCache(workingDir, projectPath, cachePath string) error {
	objExists, err := fs.Exists(filepath.Join(cachePath, "obj"))
	if err != nil {
		return err
	}

	if objExists {
		err = fs.Copy(filepath.Join(cachePath, "obj"), filepath.Join(workingDir, projectPath, "obj"))
		if err != nil {
			return err
		}
	}
	return nil
}

func recreateBuildCache(workingDir, projectPath, cachePath string) error {
	obj, err := fs.Exists(filepath.Join(workingDir, projectPath, "obj"))
	if err != nil {
		return fmt.Errorf("failed to locate build cache: %w", err)
	}

	if obj {
		err = os.RemoveAll(filepath.Join(cachePath, "obj"))
		if err != nil {
			// not tested
			return fmt.Errorf("failed to reset build cache: %w", err)
		}
		err = os.MkdirAll(filepath.Join(cachePath, "obj"), os.ModePerm)
		if err != nil {
			// not tested
			return fmt.Errorf("failed to reset build cache: %w", err)
		}
		err = fs.Copy(filepath.Join(workingDir, projectPath, "obj"), filepath.Join(cachePath, "obj"))
		if err != nil {
			return fmt.Errorf("failed to store build cache: %w", err)
		}
	}
	return nil
}
