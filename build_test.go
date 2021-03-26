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

		sourceRemover      *fakes.SourceRemover
		publishProcess     *fakes.PublishProcess
		buildpackYMLParser *fakes.BuildpackYMLParser

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(ioutil.WriteFile(filepath.Join(workingDir, "buildpack.yml"), nil, 0600)).To(Succeed())

		sourceRemover = &fakes.SourceRemover{}
		publishProcess = &fakes.PublishProcess{}

		buildpackYMLParser = &fakes.BuildpackYMLParser{}
		buildpackYMLParser.ParseProjectPathCall.Returns.ProjectFilePath = "some/project/path"

		os.Setenv("DOTNET_ROOT", "some-existing-root-dir")
		os.Setenv("BP_DOTNET_PROJECT_PATH", "")

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewLogger(buffer)

		timestamp = time.Now()
		clock := chronos.NewClock(func() time.Time {
			return timestamp
		})

		build = dotnetpublish.Build(sourceRemover, publishProcess, buildpackYMLParser, clock, logger)
	})

	it.After(func() {
		os.Unsetenv("DOTNET_ROOT")
		os.Unsetenv("BP_DOTNET_PROJECT_PATH")

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

		Expect(publishProcess.ExecuteCall.Receives.WorkingDir).To(Equal(workingDir))
		Expect(publishProcess.ExecuteCall.Receives.RootDir).To(Equal("some-existing-root-dir"))
		Expect(publishProcess.ExecuteCall.Receives.ProjectPath).To(Equal("some/project/path"))
		Expect(publishProcess.ExecuteCall.Receives.OutputPath).To(MatchRegexp(`dotnet-publish-output\d+`))

		Expect(buffer.String()).To(ContainSubstring("Some Buildpack 0.0.1"))
		Expect(buffer.String()).To(ContainSubstring("Executing build process"))
		Expect(buffer.String()).To(ContainSubstring("WARNING: Setting the project path through buildpack.yml will be deprecated soon in Dotnet Publish Buildpack v1.0.0"))
		Expect(buffer.String()).To(ContainSubstring("Please specify the project path through the $BP_DOTNET_PROJECT_PATH environment variable instead. See README.md or the documentation on paketo.io for more information."))
	})

	context("when project path is set via BP_DOTNET_PROJECT_PATH", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_DOTNET_PROJECT_PATH", "some/project/path"))
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

			Expect(publishProcess.ExecuteCall.Receives.WorkingDir).To(Equal(workingDir))
			Expect(publishProcess.ExecuteCall.Receives.RootDir).To(Equal("some-existing-root-dir"))
			Expect(publishProcess.ExecuteCall.Receives.ProjectPath).To(Equal("some/project/path"))
			Expect(publishProcess.ExecuteCall.Receives.OutputPath).To(MatchRegexp(`dotnet-publish-output\d+`))

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

		context("when the publish process fails", func() {
			it.Before(func() {
				publishProcess.ExecuteCall.Returns.Error = errors.New("some-error")
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
