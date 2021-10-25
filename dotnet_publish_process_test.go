package dotnetpublish_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	dotnetpublish "github.com/paketo-buildpacks/dotnet-publish"
	"github.com/paketo-buildpacks/packit/cargo/fakes"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDotnetPublishProcess(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path       string
		executable *fakes.Executable
		process    dotnetpublish.DotnetPublishProcess

		buffer *bytes.Buffer
	)

	it.Before(func() {
		path = os.Getenv("PATH")
		Expect(os.Setenv("PATH", "some-path")).To(Succeed())

		executable = &fakes.Executable{}

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewLogger(buffer)

		now := time.Now()
		times := []time.Time{now, now.Add(1 * time.Second)}

		clock := chronos.NewClock(func() time.Time {
			if len(times) == 0 {
				return time.Now()
			}

			t := times[0]
			times = times[1:]
			return t
		})

		process = dotnetpublish.NewDotnetPublishProcess(executable, logger, clock)
	})

	it.After(func() {
		Expect(os.Setenv("PATH", path)).To(Succeed())
	})

	it("executes the dotnet publish process", func() {
		err := process.Execute("some-working-dir", "some-dotnet-root-dir", "some/project/path", "some-publish-output-dir", []string{"--flag", "value"})
		Expect(err).NotTo(HaveOccurred())

		args := []string{
			"publish", "some-working-dir/some/project/path",
			"--configuration", "Release",
			"--runtime", "ubuntu.18.04-x64",
			"--self-contained", "false",
			"--output", "some-publish-output-dir",
			"--flag", "value",
		}

		Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal(args))

		Expect(executable.ExecuteCall.Receives.Execution.Dir).To(Equal("some-working-dir"))
		Expect(executable.ExecuteCall.Receives.Execution.Env).To(ContainElement("PATH=some-dotnet-root-dir:some-path"))

		Expect(buffer.String()).To(ContainSubstring(fmt.Sprintf("Running 'dotnet %s'", strings.Join(args, " "))))
		Expect(buffer.String()).To(ContainSubstring("Completed in 1s"))
	})

	context("when the user passes flags that the buildpack sets by default", func() {
		it("overrides the default value with the user-provided one", func() {
			err := process.Execute("some-working-dir", "some-dotnet-root-dir", "some/project/path", "some-publish-output-dir",
				[]string{"--runtime", "user-value",
					"--self-contained=true",
					"--configuration", "UserConfiguration",
					"--output", "some-user-output-dir",
				})
			Expect(err).NotTo(HaveOccurred())

			args := []string{
				"publish", "some-working-dir/some/project/path",
				"--runtime", "user-value",
				"--self-contained=true",
				"--configuration", "UserConfiguration",
				"--output", "some-user-output-dir",
			}

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal(args))
		})
	})

	context("when the user passes --no-self-contained, equivalent to --self-contained=false", func() {
		it("overrides the buildpack's value for self-contained with the user-provided one", func() {
			err := process.Execute("some-working-dir", "some-dotnet-root-dir", "some/project/path", "some-publish-output-dir", []string{"--no-self-contained"})
			Expect(err).NotTo(HaveOccurred())

			args := []string{
				"publish", "some-working-dir/some/project/path",
				"--configuration", "Release",
				"--runtime", "ubuntu.18.04-x64",
				"--output", "some-publish-output-dir",
				"--no-self-contained",
			}

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal(args))
		})
	})

	context("failure cases", func() {
		context("when the dotnet publish executable errors", func() {
			it.Before(func() {
				executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
					fmt.Fprintln(execution.Stdout, "stdout-output")
					fmt.Fprintln(execution.Stderr, "stderr-output")

					return errors.New("execution error")
				}
			})

			it("returns an error", func() {
				err := process.Execute("some-working-dir", "some-dotnet-root-dir", "", "some-output-dir", []string{})
				Expect(err).To(MatchError("failed to execute 'dotnet publish': execution error"))
			})

			it("logs the command output", func() {
				err := process.Execute("some-working-dir", "some-dotnet-root-dir", "", "some-output-dir", []string{})
				Expect(err).To(HaveOccurred())

				Expect(buffer.String()).To(ContainSubstring("      Failed after 1s"))
				Expect(buffer.String()).To(ContainSubstring("        stdout-output"))
				Expect(buffer.String()).To(ContainSubstring("        stderr-output"))
			})
		})
	})
}
