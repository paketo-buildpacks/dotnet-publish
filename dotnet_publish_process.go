package dotnetpublish

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) error
}

type DotnetPublishProcess struct {
	executable Executable
	logger     scribe.Emitter
	clock      chronos.Clock
}

func NewDotnetPublishProcess(executable Executable, logger scribe.Emitter, clock chronos.Clock) DotnetPublishProcess {
	return DotnetPublishProcess{
		executable: executable,
		logger:     logger,
		clock:      clock,
	}
}

func (p DotnetPublishProcess) Execute(workingDir, root, nugetCachePath, projectPath, outputPath string, debug bool, flags []string) error {
	args := []string{
		"publish", filepath.Join(workingDir, projectPath), // change to workingDir plus project path
	}

	if !containsFlag(flags, "--configuration") && !containsFlag(flags, "-c") {
		if debug {
			args = append(args, "--configuration", "Debug")
		} else {
			args = append(args, "--configuration", "Release")
		}
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

	return nil
}

func containsFlag(flags []string, match string) bool {
	for _, flag := range flags {
		if strings.HasPrefix(flag, match) {
			return true
		}
	}
	return false
}
