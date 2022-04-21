package dotnetpublish_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
		cacheDir   string

		workingDir string

		buffer *bytes.Buffer
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "workingDir")
		Expect(err).NotTo(HaveOccurred())

		cacheDir, err = os.MkdirTemp("", cacheDir)
		Expect(err).NotTo(HaveOccurred())

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

		executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
			fmt.Fprintln(execution.Stdout, "stdout-output")
			fmt.Fprintln(execution.Stderr, "stderr-output")

			return nil
		}

		process = dotnetpublish.NewDotnetPublishProcess(executable, logger, clock)
	})

	it.After(func() {
		Expect(os.Setenv("PATH", path)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
		Expect(os.RemoveAll(cacheDir)).To(Succeed())
	})

	it("executes the dotnet publish process", func() {
		err := process.Execute(workingDir, "some-dotnet-root-dir", "some/nuget/cache/path", cacheDir, "", "some-publish-output-dir", []string{"--flag", "value"})
		Expect(err).NotTo(HaveOccurred())

		args := []string{
			"publish",
			workingDir,
			"--configuration", "Release",
			"--runtime", "ubuntu.18.04-x64",
			"--self-contained", "false",
			"--output", "some-publish-output-dir",
			"--flag", "value",
		}

		Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal(args))

		Expect(executable.ExecuteCall.Receives.Execution.Dir).To(Equal(workingDir))
		Expect(executable.ExecuteCall.Receives.Execution.Env).To(ContainElement("PATH=some-dotnet-root-dir:some-path"))
		Expect(executable.ExecuteCall.Receives.Execution.Env).To(ContainElement("NUGET_PACKAGES=some/nuget/cache/path"))

		Expect(buffer.String()).To(ContainLines(
			fmt.Sprintf("    Running 'dotnet %s'", strings.Join(args, " ")),
			"      stdout-output",
			"      stderr-output",
			"      Completed in 1s",
		))
	})

	context("when there is a project specific path", func() {
		var projectPath string

		it.Before(func() {
			projectPath = filepath.Join("src", "project")

			Expect(os.MkdirAll(filepath.Join(workingDir, projectPath), os.ModePerm)).To(Succeed())
		})

		it("executes the dotnet publish process", func() {
			err := process.Execute(workingDir, "some-dotnet-root-dir", "some/nuget/cache/path", cacheDir, projectPath, "some-publish-output-dir", []string{"--flag", "value"})
			Expect(err).NotTo(HaveOccurred())

			args := []string{
				"publish",
				filepath.Join(workingDir, projectPath),
				"--configuration", "Release",
				"--runtime", "ubuntu.18.04-x64",
				"--self-contained", "false",
				"--output", "some-publish-output-dir",
				"--flag", "value",
			}

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal(args))

			Expect(executable.ExecuteCall.Receives.Execution.Dir).To(Equal(workingDir))
			Expect(executable.ExecuteCall.Receives.Execution.Env).To(ContainElement("PATH=some-dotnet-root-dir:some-path"))
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
			err := process.Execute(workingDir, "some-dotnet-root-dir", "some/nuget/cache/path", cacheDir, "", "some-publish-output-dir",
				[]string{"--runtime", "user-value",
					"--self-contained=true",
					"--configuration", "UserConfiguration",
					"--output", "some-user-output-dir",
				})
			Expect(err).NotTo(HaveOccurred())

			args := []string{
				"publish",
				workingDir,
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
			err := process.Execute(workingDir, "some-dotnet-root-dir", "some/nuget/cache/path", cacheDir, "", "some-publish-output-dir", []string{"--no-self-contained"})
			Expect(err).NotTo(HaveOccurred())

			args := []string{
				"publish",
				workingDir,
				"--configuration", "Release",
				"--runtime", "ubuntu.18.04-x64",
				"--output", "some-publish-output-dir",
				"--no-self-contained",
			}

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal(args))
		})
	})
	context("when the build generates an obj directory", func() {
		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(workingDir, "project-path"), os.ModePerm)).To(Succeed())
			executable.ExecuteCall.Stub = func(e pexec.Execution) error {
				return os.MkdirAll(filepath.Join(workingDir, "project-path", "obj"), os.ModePerm)
			}
			Expect(os.MkdirAll(filepath.Join(cacheDir, "obj"), os.ModePerm)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(cacheDir, "obj", "some-contents"), nil, os.ModePerm)).To(Succeed())
		})
		it("copies the obj directory into the cache layer", func() {
			err := process.Execute(workingDir, "some-dotnet-root-dir", "some/nuget/cache/path", cacheDir, "project-path", "some-publish-output-dir", []string{"--no-self-contained"})
			Expect(err).NotTo(HaveOccurred())

			Expect(filepath.Join(cacheDir, "obj")).To(BeADirectory())
		})
		context("and the cache layer directory doesn't exist yet", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(workingDir, "project-path"), os.ModePerm)).To(Succeed())
				executable.ExecuteCall.Stub = func(e pexec.Execution) error {
					return os.MkdirAll(filepath.Join(workingDir, "project-path", "obj"), os.ModePerm)
				}
				Expect(os.RemoveAll(cacheDir)).To(Succeed())
			})
			it("creates the layer dir and moves the cache contents into it", func() {
				err := process.Execute(workingDir, "some-dotnet-root-dir", "some/nuget/cache/path", cacheDir, "project-path", "some-publish-output-dir", []string{"--no-self-contained"})
				Expect(err).NotTo(HaveOccurred())

				Expect(filepath.Join(cacheDir, "obj")).To(BeADirectory())
			})
		})
	})

	context("when the cache has an obj directory from a previous build", func() {
		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(cacheDir, "obj"), os.ModePerm)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(workingDir, "project-path"), os.ModePerm)).To(Succeed())
		})
		it("copies the obj directory into the working directory", func() {
			err := process.Execute(workingDir, "some-dotnet-root-dir", "some/nuget/cache/path", cacheDir, "project-path", "some-publish-output-dir", []string{"--no-self-contained"})
			Expect(err).NotTo(HaveOccurred())

			Expect(filepath.Join(workingDir, "project-path", "obj")).To(BeADirectory())
		})
		context("and there's already an obj directory in the /workspace", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(cacheDir, "obj"), os.ModePerm)).To(Succeed())
				Expect(os.MkdirAll(filepath.Join(workingDir, "project-path", "obj"), os.ModePerm)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, "project-path", "obj", "some-contents"), nil, os.ModePerm)).To(Succeed())
			})
			it("overwrites the /workspace obj with the layer obj dir", func() {
				err := process.Execute(workingDir, "some-dotnet-root-dir", "some/nuget/cache/path", cacheDir, "project-path", "some-publish-output-dir", []string{"--no-self-contained"})
				Expect(err).NotTo(HaveOccurred())

				Expect(filepath.Join(workingDir, "project-path", "obj")).To(BeADirectory())
			})
		})
	})

	context("failure cases", func() {
		context("when check for existence of obj file in cache layer fails", func() {
			it.Before(func() {
				Expect(os.Chmod(cacheDir, 0000)).To(Succeed())
			})
			it.After(func() {
				Expect(os.Chmod(cacheDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				err := process.Execute(workingDir, "some-dotnet-root-dir", "some/nuget/cache/path", cacheDir, "", "some-output-dir", []string{})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
		context("when removing obj directory in working dir fails", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(cacheDir, "obj"), os.ModePerm))
				Expect(os.MkdirAll(filepath.Join(workingDir, "obj"), 0500)).To(Succeed())
				Expect(os.Chmod(workingDir, 0500)).To(Succeed())
			})
			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				err := process.Execute(workingDir, "some-dotnet-root-dir", "some/nuget/cache/path", cacheDir, "", "some-output-dir", []string{})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
		context("when copying build cache into working directory fails", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(cacheDir, "obj"), 0200))
			})
			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				err := process.Execute(workingDir, "some-dotnet-root-dir", "some/nuget/cache/path", cacheDir, "", "some-output-dir", []string{})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when checking for obj dir generated during build fails", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0200)).To(Succeed())
			})
			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				err := process.Execute(workingDir, "some-dotnet-root-dir", "some/nuget/cache/path", cacheDir, "", "some-output-dir", []string{})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when copying obj dir into cache layer fails", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(workingDir, "obj"), 0000))
			})

			it("returns an error", func() {
				err := process.Execute(workingDir, "some-dotnet-root-dir", "some/nuget/cache/path", cacheDir, "", "some-output-dir", []string{})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the dotnet publish executable errors", func() {
			it.Before(func() {
				executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
					fmt.Fprintln(execution.Stdout, "stdout-output")
					fmt.Fprintln(execution.Stderr, "stderr-output")

					return errors.New("execution error")
				}
			})

			it("returns an error", func() {
				err := process.Execute(workingDir, "some-dotnet-root-dir", "some/nuget/cache/path", cacheDir, "", "some-output-dir", []string{})
				Expect(err).To(MatchError("failed to execute 'dotnet publish': execution error"))
			})

			it("logs the command output", func() {
				err := process.Execute(workingDir, "some-dotnet-root-dir", "some/nuget/cache/path", cacheDir, "", "some-output-dir", []string{})
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
