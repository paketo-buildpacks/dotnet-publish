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

func (p DotnetProcess) Restore(workingDir, root, projectPath string) error {
	buffer := bytes.NewBuffer(nil)
	args := []string{
		"restore", filepath.Join(workingDir, projectPath),
		"--runtime", "ubuntu.18.04-x64",
	}

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

func (p DotnetProcess) Publish(workingDir, root, projectPath, outputPath string) error {
	buffer := bytes.NewBuffer(nil)
	args := []string{
		"publish", filepath.Join(workingDir, projectPath), // change to workingDir plus project path
		"--no-restore",
		"--configuration", "Release",
		"--runtime", "ubuntu.18.04-x64",
		"--self-contained", "false",
		"--output", outputPath,
	}

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
