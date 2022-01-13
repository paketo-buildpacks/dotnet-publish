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
	"github.com/paketo-buildpacks/packit/servicebindings"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		timestamp  time.Time
		buffer     *bytes.Buffer
		workingDir string
		homeDir    string

		symlinker           *fakes.SymlinkManager
		sourceRemover       *fakes.SourceRemover
		publishProcess      *fakes.PublishProcess
		bindingResolver     *fakes.BindingResolver
		buildpackYMLParser  *fakes.BuildpackYMLParser
		commandConfigParser *fakes.CommandConfigParser

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		homeDir, err = ioutil.TempDir("", "home-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(ioutil.WriteFile(filepath.Join(workingDir, "buildpack.yml"), nil, 0600)).To(Succeed())

		symlinker = &fakes.SymlinkManager{}
		sourceRemover = &fakes.SourceRemover{}
		publishProcess = &fakes.PublishProcess{}
		bindingResolver = &fakes.BindingResolver{}

		buildpackYMLParser = &fakes.BuildpackYMLParser{}
		buildpackYMLParser.ParseProjectPathCall.Returns.ProjectFilePath = "some/project/path"

		commandConfigParser = &fakes.CommandConfigParser{}
		commandConfigParser.ParseFlagsFromEnvVarCall.Returns.StringSlice = []string{"--publishflag", "value"}

		os.Setenv("DOTNET_ROOT", "some-existing-root-dir")

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewLogger(buffer)

		timestamp = time.Now()
		clock := chronos.NewClock(func() time.Time {
			return timestamp
		})

		build = dotnetpublish.Build(sourceRemover, bindingResolver, homeDir, symlinker, publishProcess, buildpackYMLParser, commandConfigParser, clock, logger)
	})

	it.After(func() {
		os.Unsetenv("DOTNET_ROOT")
		Expect(os.RemoveAll(workingDir)).To(Succeed())
		Expect(os.RemoveAll(homeDir)).To(Succeed())
	})

	it("returns a build result", func() {
		result, err := build(packit.BuildContext{
			WorkingDir: workingDir,
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "0.0.1",
			},
			Platform: packit.Platform{
				Path: "some-platform-path",
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(packit.BuildResult{}))

		Expect(sourceRemover.RemoveCall.Receives.WorkingDir).To(Equal(workingDir))
		Expect(sourceRemover.RemoveCall.Receives.PublishOutputDir).To(MatchRegexp(`dotnet-publish-output\d+`))
		Expect(sourceRemover.RemoveCall.Receives.ExcludedFiles).To(ConsistOf([]string{".dotnet_root"}))

		Expect(bindingResolver.ResolveCall.Receives.Typ).To(Equal("nugetconfig"))
		Expect(bindingResolver.ResolveCall.Receives.PlatformDir).To(Equal("some-platform-path"))
		Expect(symlinker.LinkCall.CallCount).To(Equal(0))
		Expect(symlinker.UnlinkCall.CallCount).To(Equal(0))

		Expect(publishProcess.ExecuteCall.Receives.WorkingDir).To(Equal(workingDir))
		Expect(publishProcess.ExecuteCall.Receives.RootDir).To(Equal("some-existing-root-dir"))
		Expect(publishProcess.ExecuteCall.Receives.ProjectPath).To(Equal("some/project/path"))
		Expect(publishProcess.ExecuteCall.Receives.OutputPath).To(MatchRegexp(`dotnet-publish-output\d+`))
		Expect(publishProcess.ExecuteCall.Receives.Flags).To(Equal([]string{"--publishflag", "value"}))

		Expect(buffer.String()).To(ContainSubstring("Some Buildpack 0.0.1"))
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

			Expect(publishProcess.ExecuteCall.Receives.WorkingDir).To(Equal(workingDir))
			Expect(publishProcess.ExecuteCall.Receives.RootDir).To(Equal("some-existing-root-dir"))
			Expect(publishProcess.ExecuteCall.Receives.ProjectPath).To(Equal("some/project/path"))
			Expect(publishProcess.ExecuteCall.Receives.OutputPath).To(MatchRegexp(`dotnet-publish-output\d+`))
			Expect(publishProcess.ExecuteCall.Receives.Flags).To(Equal([]string{"--publishflag", "value"}))

			Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
			Expect(buffer.String()).To(ContainSubstring("Executing build process"))
		})
	})

	context("when a NuGet.Config is provide via service binding", func() {
		it.Before(func() {
			bindingResolver.ResolveCall.Returns.BindingSlice = []servicebindings.Binding{
				servicebindings.Binding{
					Name: "some-binding",
					Path: "some-binding-path",
					Type: "nugetconfig",
					Entries: map[string]*servicebindings.Entry{
						"nuget.config": servicebindings.NewEntry("some-binding-path"),
					},
				},
			}
		})

		it("symlinks the provided config file into $HOME/.nuget/NuGet/Nuget.Config during build", func() {
			_, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "0.0.1",
				},
				Platform: packit.Platform{
					Path: "some-platform-path",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(bindingResolver.ResolveCall.Receives.Typ).To(Equal("nugetconfig"))
			Expect(bindingResolver.ResolveCall.Receives.PlatformDir).To(Equal("some-platform-path"))

			Expect(symlinker.LinkCall.Receives.Oldname).To(Equal(filepath.Join("some-binding-path", "nuget.config")))
			Expect(symlinker.LinkCall.Receives.Newname).To(Equal(filepath.Join(homeDir, ".nuget", "NuGet", "NuGet.Config")))
			Expect(symlinker.UnlinkCall.CallCount).To(Equal(1))
			Expect(symlinker.UnlinkCall.Receives.Path).To(Equal(filepath.Join(homeDir, ".nuget", "NuGet", "NuGet.Config")))
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

		context("BP_DOTNET_PUBLISH_FLAGS cannot be parsed", func() {
			it.Before(func() {
				commandConfigParser.ParseFlagsFromEnvVarCall.Returns.Error = errors.New("some publish parsing error")
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

		context("when the binding resolution fails", func() {
			it.Before(func() {
				bindingResolver.ResolveCall.Returns.Error = errors.New("some-error")
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

		context("when the more than one nuget.config binding is provided", func() {
			it.Before(func() {
				bindingResolver.ResolveCall.Returns.BindingSlice = []servicebindings.Binding{
					servicebindings.Binding{
						Name: "some-binding",
						Path: "some-binding-path",
						Type: "nugetconfig",
						Entries: map[string]*servicebindings.Entry{
							"nuget.config": servicebindings.NewEntry("some-binding-path"),
						},
					},
					servicebindings.Binding{
						Name: "some-binding-2",
						Path: "some-binding-path-2",
						Type: "nugetconfig",
						Entries: map[string]*servicebindings.Entry{
							"nuget.config": servicebindings.NewEntry("some-binding-path-2"),
						},
					},
				}
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					BuildpackInfo: packit.BuildpackInfo{
						Version: "0.0.1",
					},
				})
				Expect(err).To(MatchError("binding resolver found more than one binding of type 'nugetconfig'"))
			})
		})

		context("when the nuget.config service binding doens't contain a nuget.config file", func() {
			it.Before(func() {
				bindingResolver.ResolveCall.Returns.BindingSlice = []servicebindings.Binding{
					servicebindings.Binding{
						Name: "some-binding",
						Path: "some-binding-path",
						Type: "nugetconfig",
						Entries: map[string]*servicebindings.Entry{
							"random.config": servicebindings.NewEntry("some-binding-path"),
						},
					},
				}
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					BuildpackInfo: packit.BuildpackInfo{
						Version: "0.0.1",
					},
				})
				Expect(err).To(MatchError("binding of type nugetconfig does not contain required entry nuget.config"))
			})
		})

		context("when the $HOME/.nuget/NuGet path cannot be created for the nuget.config", func() {
			it.Before(func() {
				bindingResolver.ResolveCall.Returns.BindingSlice = []servicebindings.Binding{
					servicebindings.Binding{
						Name: "some-binding",
						Path: "some-binding-path",
						Type: "nugetconfig",
						Entries: map[string]*servicebindings.Entry{
							"nuget.config": servicebindings.NewEntry("some-binding-path"),
						},
					},
				}
				Expect(os.Mkdir(filepath.Join(homeDir, ".nuget"), 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.RemoveAll(filepath.Join(homeDir, ".nuget"))).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					BuildpackInfo: packit.BuildpackInfo{
						Version: "0.0.1",
					},
				})
				Expect(err).To(MatchError(ContainSubstring("failed to make directory for NuGet.Config")))
			})
		})

		context("when symlinking the nuget.config path to the binding path fails", func() {
			it.Before(func() {
				bindingResolver.ResolveCall.Returns.BindingSlice = []servicebindings.Binding{
					servicebindings.Binding{
						Name: "some-binding",
						Path: "some-binding-path",
						Type: "nugetconfig",
						Entries: map[string]*servicebindings.Entry{
							"nuget.config": servicebindings.NewEntry("some-binding-path"),
						},
					},
				}

				symlinker.LinkCall.Returns.Error = errors.New("failed to symlink")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					BuildpackInfo: packit.BuildpackInfo{
						Version: "0.0.1",
					},
				})
				Expect(err).To(MatchError(ContainSubstring("failed to symlink")))
			})
		})

		context("when removing the symlink between the nuget.config path and the binding path fails", func() {
			it.Before(func() {
				bindingResolver.ResolveCall.Returns.BindingSlice = []servicebindings.Binding{
					servicebindings.Binding{
						Name: "some-binding",
						Path: "some-binding-path",
						Type: "nugetconfig",
						Entries: map[string]*servicebindings.Entry{
							"nuget.config": servicebindings.NewEntry("some-binding-path"),
						},
					},
				}
				symlinker.UnlinkCall.Returns.Error = errors.New("failed to remove symlink")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					BuildpackInfo: packit.BuildpackInfo{
						Version: "0.0.1",
					},
				})
				Expect(err).To(MatchError(ContainSubstring("failed to remove symlink")))
			})
		})
	})
}
