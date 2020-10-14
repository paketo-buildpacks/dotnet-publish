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

func (p DotnetPublishProcess) Execute(workingDir, root, projectPath string) error {
	buffer := bytes.NewBuffer(nil)
	args := []string{
		"publish", filepath.Join(workingDir, projectPath), // change to workingDir plus project path
		"--configuration", "Release",
		"--runtime", "ubuntu.18.04-x64",
		"--self-contained", "false",
		"--output", workingDir,
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
