package dotnetpublish

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/scribe"
)

type Executable interface {
	Execute(pexec.Execution) error
}

type DotnetProcess struct {
	executable Executable
	logger     scribe.Logger
	clock      chronos.Clock
}

func NewDotnetProcess(executable Executable, logger scribe.Logger, clock chronos.Clock) DotnetProcess {
	return DotnetProcess{
		executable: executable,
		logger:     logger,
		clock:      clock,
	}
}

func (p DotnetProcess) Restore(workingDir, root, projectPath string, flags []string) error {
	buffer := bytes.NewBuffer(nil)
	args := []string{
		"restore", filepath.Join(workingDir, projectPath),
	}

	if !containsFlag(flags, "--runtime") && !containsFlag(flags, "-r") {
		args = append(args, "--runtime", "ubuntu.18.04-x64")
	}

	args = append(args, flags...)

	p.logger.Subprocess("Running 'dotnet %s'", strings.Join(args, " "))

	duration, err := p.clock.Measure(func() error {
		return p.executable.Execute(pexec.Execution{
			Args:   args,
			Dir:    workingDir,
			Env:    append(os.Environ(), fmt.Sprintf("PATH=%s:%s", root, os.Getenv("PATH"))),
			Stdout: buffer,
			Stderr: buffer,
		})
	})
	if err != nil {
		p.logger.Action("Failed after %s", duration)
		p.logger.Detail(buffer.String())

		return fmt.Errorf("failed to execute 'dotnet restore': %w", err)
	}

	p.logger.Action("Completed in %s", duration)
	p.logger.Break()

	return nil
}

func (p DotnetProcess) Publish(workingDir, root, projectPath, outputPath string, flags []string) error {
	buffer := bytes.NewBuffer(nil)
	args := []string{
		"publish", filepath.Join(workingDir, projectPath), // change to workingDir plus project path
		"--no-restore",
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
			Env:    append(os.Environ(), fmt.Sprintf("PATH=%s:%s", root, os.Getenv("PATH"))),
			Stdout: buffer,
			Stderr: buffer,
		})
	})
	if err != nil {
		p.logger.Action("Failed after %s", duration)
		p.logger.Detail(buffer.String())

		return fmt.Errorf("failed to execute 'dotnet publish': %w", err)
	}

	p.logger.Action("Completed in %s", duration)
	p.logger.Break()

	return nil
}
