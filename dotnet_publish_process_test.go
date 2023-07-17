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
	"github.com/paketo-buildpacks/dotnet-publish/fakes"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
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
		logger := scribe.NewEmitter(buffer)

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

		executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
			fmt.Fprintln(execution.Stdout, "stdout-output")
			fmt.Fprintln(execution.Stderr, "stderr-output")

			return nil
		}

		process = dotnetpublish.NewDotnetPublishProcess(executable, logger, clock)
	})

	it.After(func() {
		Expect(os.Setenv("PATH", path)).To(Succeed())
	})

	it("executes the dotnet publish process", func() {
		err := process.Execute("some-working-dir", "some/nuget/cache/path", "some/project/path", "some-publish-output-dir", false, []string{"--flag", "value"})
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
		Expect(executable.ExecuteCall.Receives.Execution.Env).To(ContainElement("NUGET_PACKAGES=some/nuget/cache/path"))

		Expect(buffer.String()).To(ContainLines(
			fmt.Sprintf("    Running 'dotnet %s'", strings.Join(args, " ")),
			"      stdout-output",
			"      stderr-output",
			"      Completed in 1s",
		))
	})
	context("when debug mode is enabled", func() {
		it("adds Debug to the publish configuration", func() {
			err := process.Execute("some-working-dir", "some/nuget/cache/path", "some/project/path", "some-publish-output-dir", true, []string{"--flag", "value"})
			Expect(err).NotTo(HaveOccurred())

			args := []string{
				"publish", "some-working-dir/some/project/path",
				"--configuration", "Debug",
				"--runtime", "ubuntu.18.04-x64",
				"--self-contained", "false",
				"--output", "some-publish-output-dir",
				"--flag", "value",
			}

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal(args))

			Expect(executable.ExecuteCall.Receives.Execution.Dir).To(Equal("some-working-dir"))
			Expect(executable.ExecuteCall.Receives.Execution.Env).To(ContainElement("NUGET_PACKAGES=some/nuget/cache/path"))

			Expect(buffer.String()).To(ContainLines(
				fmt.Sprintf("    Running 'dotnet %s'", strings.Join(args, " ")),
				"      stdout-output",
				"      stderr-output",
				"      Completed in 1s",
			))
		})
	})

	context("when the user passes flags that the buildpack sets by default", func() {
		it("overrides the default value with the user-provided one", func() {
			err := process.Execute("some-working-dir", "some/nuget/cache/path", "some/project/path", "some-publish-output-dir",
				true,
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
			err := process.Execute("some-working-dir", "some/nuget/cache/path", "some/project/path", "some-publish-output-dir", false, []string{"--no-self-contained"})
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
				err := process.Execute("some-working-dir", "some/nuget/cache/path", "", "some-output-dir", false, []string{})
				Expect(err).To(MatchError("failed to execute 'dotnet publish': execution error"))
			})

			it("logs the command output", func() {
				err := process.Execute("some-working-dir", "some/nuget/cache/path", "", "some-output-dir", false, []string{})
				Expect(err).To(HaveOccurred())

				Expect(buffer.String()).To(ContainLines(
					"      stdout-output",
					"      stderr-output",
					"      Failed after 1s",
				))
			})
		})
	})
}
