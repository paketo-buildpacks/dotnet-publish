package dotnetpublish_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	dotnetpublish "github.com/paketo-buildpacks/dotnet-publish"
	"github.com/paketo-buildpacks/dotnet-publish/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		timestamp  time.Time
		buffer     *bytes.Buffer
		workingDir string

		sourceRemover       *fakes.SourceRemover
		dotnetProcess       *fakes.Dotnet
		buildpackYMLParser  *fakes.BuildpackYMLParser
		commandConfigParser *fakes.CommandConfigParser

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(ioutil.WriteFile(filepath.Join(workingDir, "buildpack.yml"), nil, 0600)).To(Succeed())

		sourceRemover = &fakes.SourceRemover{}
		dotnetProcess = &fakes.Dotnet{}

		buildpackYMLParser = &fakes.BuildpackYMLParser{}
		buildpackYMLParser.ParseProjectPathCall.Returns.ProjectFilePath = "some/project/path"

		commandConfigParser = &fakes.CommandConfigParser{}
		commandConfigParser.ParseFlagsFromEnvVarCall.Stub = func(envVar string) ([]string, error) {
			if envVar == "BP_DOTNET_RESTORE_FLAGS" {
				return []string{"--restoreflag", "value"}, nil
			}
			if envVar == "BP_DOTNET_PUBLISH_FLAGS" {
				return []string{"--publishflag", "value"}, nil
			}
			return nil, nil
		}

		os.Setenv("DOTNET_ROOT", "some-existing-root-dir")

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewLogger(buffer)

		timestamp = time.Now()
		clock := chronos.NewClock(func() time.Time {
			return timestamp
		})

		build = dotnetpublish.Build(sourceRemover, dotnetProcess, buildpackYMLParser, commandConfigParser, clock, logger)
	})

	it.After(func() {
		os.Unsetenv("DOTNET_ROOT")
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("returns a build result", func() {
		result, err := build(packit.BuildContext{
			WorkingDir: workingDir,
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "0.0.1",
			},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(packit.BuildResult{}))

		Expect(sourceRemover.RemoveCall.Receives.WorkingDir).To(Equal(workingDir))
		Expect(sourceRemover.RemoveCall.Receives.PublishOutputDir).To(MatchRegexp(`dotnet-publish-output\d+`))
		Expect(sourceRemover.RemoveCall.Receives.ExcludedFiles).To(ConsistOf([]string{".dotnet_root"}))

		Expect(dotnetProcess.RestoreCall.Receives.WorkingDir).To(Equal(workingDir))
		Expect(dotnetProcess.RestoreCall.Receives.RootDir).To(Equal("some-existing-root-dir"))
		Expect(dotnetProcess.RestoreCall.Receives.ProjectPath).To(Equal("some/project/path"))
		Expect(dotnetProcess.RestoreCall.Receives.Flags).To(Equal([]string{"--restoreflag", "value"}))

		Expect(dotnetProcess.PublishCall.Receives.WorkingDir).To(Equal(workingDir))
		Expect(dotnetProcess.PublishCall.Receives.RootDir).To(Equal("some-existing-root-dir"))
		Expect(dotnetProcess.PublishCall.Receives.ProjectPath).To(Equal("some/project/path"))
		Expect(dotnetProcess.PublishCall.Receives.OutputPath).To(MatchRegexp(`dotnet-publish-output\d+`))
		Expect(dotnetProcess.PublishCall.Receives.Flags).To(Equal([]string{"--publishflag", "value"}))

		Expect(buffer.String()).To(ContainSubstring("Some Buildpack 0.0.1"))
		Expect(buffer.String()).To(ContainSubstring("Executing package restore process"))
		Expect(buffer.String()).To(ContainSubstring("Executing build process"))
		Expect(buffer.String()).To(ContainSubstring("WARNING: Setting the project path through buildpack.yml will be deprecated soon in Dotnet Publish Buildpack v1.0.0"))
		Expect(buffer.String()).To(ContainSubstring("Please specify the project path through the $BP_DOTNET_PROJECT_PATH environment variable instead. See README.md or the documentation on paketo.io for more information."))
	})

	context("when project path is set via BP_DOTNET_PROJECT_PATH", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_DOTNET_PROJECT_PATH", "some/project/path"))
		})

		it.After(func() {
			os.Unsetenv("BP_DOTNET_PROJECT_PATH")
		})

		it("returns a build result", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.BuildResult{}))

			Expect(sourceRemover.RemoveCall.Receives.WorkingDir).To(Equal(workingDir))
			Expect(sourceRemover.RemoveCall.Receives.PublishOutputDir).To(MatchRegexp(`dotnet-publish-output\d+`))
			Expect(sourceRemover.RemoveCall.Receives.ExcludedFiles).To(ConsistOf([]string{".dotnet_root"}))

			Expect(dotnetProcess.RestoreCall.Receives.WorkingDir).To(Equal(workingDir))
			Expect(dotnetProcess.RestoreCall.Receives.RootDir).To(Equal("some-existing-root-dir"))
			Expect(dotnetProcess.RestoreCall.Receives.ProjectPath).To(Equal("some/project/path"))

			Expect(dotnetProcess.PublishCall.Receives.WorkingDir).To(Equal(workingDir))
			Expect(dotnetProcess.PublishCall.Receives.RootDir).To(Equal("some-existing-root-dir"))
			Expect(dotnetProcess.PublishCall.Receives.ProjectPath).To(Equal("some/project/path"))
			Expect(dotnetProcess.PublishCall.Receives.OutputPath).To(MatchRegexp(`dotnet-publish-output\d+`))

			Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
			Expect(buffer.String()).To(ContainSubstring("Executing build process"))
		})
	})

	context("failure cases", func() {
		context("when the source code cannot be removed", func() {
			it.Before(func() {
				sourceRemover.RemoveCall.Returns.Error = errors.New("some-error")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					BuildpackInfo: packit.BuildpackInfo{
						Version: "0.0.1",
					},
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		context("when the buildpack.yml can not be parsed", func() {
			it.Before(func() {
				buildpackYMLParser.ParseProjectPathCall.Returns.Err = errors.New("some-error")
			})
			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					BuildpackInfo: packit.BuildpackInfo{
						Version: "0.0.1",
					},
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		context("BP_DOTNET_RESTORE_FLAGS cannot be parsed", func() {
			it.Before(func() {
				commandConfigParser.ParseFlagsFromEnvVarCall.Stub = func(envVar string) ([]string, error) {
					if envVar == "BP_DOTNET_RESTORE_FLAGS" {
						return nil, errors.New("some restore parsing error")
					}
					return nil, nil
				}
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					BuildpackInfo: packit.BuildpackInfo{
						Version: "0.0.1",
					},
				})
				Expect(err).To(MatchError("some restore parsing error"))
			})
		})

		context("when the restore process fails", func() {
			it.Before(func() {
				dotnetProcess.RestoreCall.Returns.Error = errors.New("some-error")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					BuildpackInfo: packit.BuildpackInfo{
						Version: "0.0.1",
					},
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		context("BP_DOTNET_PUBLISH_FLAGS cannot be parsed", func() {
			it.Before(func() {
				commandConfigParser.ParseFlagsFromEnvVarCall.Stub = func(envVar string) ([]string, error) {
					if envVar == "BP_DOTNET_PUBLISH_FLAGS" {
						return nil, errors.New("some publish parsing error")
					}
					return nil, nil
				}
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					BuildpackInfo: packit.BuildpackInfo{
						Version: "0.0.1",
					},
				})
				Expect(err).To(MatchError("some publish parsing error"))
			})
		})

		context("when the publish process fails", func() {
			it.Before(func() {
				dotnetProcess.PublishCall.Returns.Error = errors.New("some-error")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					BuildpackInfo: packit.BuildpackInfo{
						Version: "0.0.1",
					},
				})
				Expect(err).To(MatchError("some-error"))
			})
		})
	})
}
