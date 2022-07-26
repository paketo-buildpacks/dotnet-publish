package dotnetpublish_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	dotnetpublish "github.com/paketo-buildpacks/dotnet-publish"
	"github.com/paketo-buildpacks/dotnet-publish/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/paketo-buildpacks/packit/v2/servicebindings"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer     *bytes.Buffer
		workingDir string
		homeDir    string
		layersDir  string

		bindingResolver    *fakes.BindingResolver
		buildpackYMLParser *fakes.BuildpackYMLParser
		publishProcess     *fakes.PublishProcess
		slicer             *fakes.Slicer
		sourceRemover      *fakes.SourceRemover
		symlinker          *fakes.SymlinkManager
		logger             scribe.Emitter

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers-dir")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		homeDir, err = os.MkdirTemp("", "home-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.WriteFile(filepath.Join(workingDir, "buildpack.yml"), nil, 0600)).To(Succeed())

		symlinker = &fakes.SymlinkManager{}
		sourceRemover = &fakes.SourceRemover{}
		publishProcess = &fakes.PublishProcess{}
		bindingResolver = &fakes.BindingResolver{}
		slicer = &fakes.Slicer{}

		buildpackYMLParser = &fakes.BuildpackYMLParser{}
		buildpackYMLParser.ParseProjectPathCall.Returns.ProjectFilePath = "some/project/path"

		slicer.SliceCall.Returns.Pkgs = packit.Slice{Paths: []string{"some-package.dll"}}
		slicer.SliceCall.Returns.EarlyPkgs = packit.Slice{Paths: []string{"some-release-candidate-package.dll"}}
		slicer.SliceCall.Returns.Projects = packit.Slice{Paths: []string{"some-project.dll"}}

		Expect(os.MkdirAll(filepath.Join(layersDir, "nuget-cache"), os.ModePerm)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(layersDir, "nuget-cache", "some-cache"), []byte{}, 0600)).To(Succeed())

		os.Setenv("DOTNET_ROOT", "some-existing-root-dir")

		buffer = bytes.NewBuffer(nil)
		logger = scribe.NewEmitter(buffer)

		build = dotnetpublish.Build(
			dotnetpublish.Configuration{RawPublishFlags: "--publishflag value"},
			sourceRemover,
			bindingResolver,
			homeDir,
			symlinker,
			publishProcess,
			slicer,
			buildpackYMLParser,
			chronos.DefaultClock,
			logger)
	})

	it.After(func() {
		os.Unsetenv("DOTNET_ROOT")
		Expect(os.RemoveAll(workingDir)).To(Succeed())
		Expect(os.RemoveAll(homeDir)).To(Succeed())
		Expect(os.RemoveAll(layersDir)).To(Succeed())
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
			Layers: packit.Layers{Path: layersDir},
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		layer := result.Layers[0]

		Expect(layer.Name).To(Equal("nuget-cache"))
		Expect(layer.Path).To(Equal(filepath.Join(layersDir, "nuget-cache")))
		Expect(layer.Cache).To(BeTrue())

		Expect(result.Launch.Slices).To(Equal([]packit.Slice{
			{Paths: []string{".dotnet_root"}},
			{Paths: []string{"some-package.dll"}},
			{Paths: []string{"some-release-candidate-package.dll"}},
			{Paths: []string{"some-project.dll"}},
		}))

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

		Expect(slicer.SliceCall.Receives.AssetsFile).To(Equal(filepath.Join(workingDir, "some/project/path", "obj", "project.assets.json")))

		Expect(buffer.String()).To(ContainSubstring("Some Buildpack 0.0.1"))
		Expect(buffer.String()).To(ContainSubstring("Executing build process"))
		Expect(buffer.String()).To(ContainSubstring("WARNING: Setting the project path through buildpack.yml will be deprecated soon in Dotnet Publish Buildpack v1.0.0"))
		Expect(buffer.String()).To(ContainSubstring("Please specify the project path through the $BP_DOTNET_PROJECT_PATH environment variable instead. See README.md or the documentation on paketo.io for more information."))
	})

	context("the cache layer is empty", func() {
		it.Before(func() {
			Expect(os.Remove(filepath.Join(layersDir, "nuget-cache", "some-cache"))).To(Succeed())
		})

		it("does not return a layer", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "0.0.1",
				},
				Platform: packit.Platform{
					Path: "some-platform-path",
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(0))
		})
	})

	context("when project path is set via BP_DOTNET_PROJECT_PATH", func() {
		it.Before(func() {
			build = dotnetpublish.Build(
				dotnetpublish.Configuration{
					RawPublishFlags: "--publishflag value",
					ProjectPath:     "some/project/path",
				},
				sourceRemover,
				bindingResolver,
				homeDir,
				symlinker,
				publishProcess,
				slicer,
				buildpackYMLParser,
				chronos.DefaultClock,
				logger)
		})

		it("returns a build result", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Name).To(Equal("nuget-cache"))
			Expect(layer.Path).To(Equal(filepath.Join(layersDir, "nuget-cache")))
			Expect(layer.Cache).To(BeTrue())

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
			Expect(buffer.String()).To(ContainSubstring("Dividing build output into layers to optimize cache reuse"))
			Expect(buffer.String()).To(ContainSubstring("Removing source code"))
		})
	})

	context("when a NuGet.Config is provided via service binding", func() {
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

	context("when output slicer produces an empty slice", func() {
		it.Before(func() {
			slicer.SliceCall.Returns.Pkgs = packit.Slice{Paths: []string{}}
		})
		it("does not attach empty slices to the build result", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "0.0.1",
				},
				Platform: packit.Platform{
					Path: "some-platform-path",
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Launch.Slices).To(Equal([]packit.Slice{
				{Paths: []string{".dotnet_root"}},
				{Paths: []string{"some-release-candidate-package.dll"}},
				{Paths: []string{"some-project.dll"}},
			}))
		})
	})
	context("when output slicing is turned off via BP_DOTNET_DISABLE_BUILDPACK_OUTPUT_SLICING", func() {
		it.Before(func() {
			build = dotnetpublish.Build(
				dotnetpublish.Configuration{DisableOutputSlicing: true},
				sourceRemover,
				bindingResolver,
				homeDir,
				symlinker,
				publishProcess,
				slicer,
				buildpackYMLParser,
				chronos.DefaultClock,
				logger)
		})

		it("returns a build result", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "0.0.1",
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(slicer.SliceCall.CallCount).To(BeZero())
			Expect(result.Launch.Slices).To(Equal([]packit.Slice{
				{Paths: []string{".dotnet_root"}},
			}))

			Expect(buffer.String()).NotTo(ContainSubstring("Dividing build output into layers to optimize cache reuse"))
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

		context("dotnet publish flags cannot be parsed", func() {
			it.Before(func() {
				build = dotnetpublish.Build(
					dotnetpublish.Configuration{RawPublishFlags: "\""},
					sourceRemover,
					bindingResolver,
					homeDir,
					symlinker,
					publishProcess,
					slicer,
					buildpackYMLParser,
					chronos.DefaultClock,
					logger)
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					BuildpackInfo: packit.BuildpackInfo{
						Version: "0.0.1",
					},
				})
				Expect(err).To(MatchError("failed to parse flags for dotnet publish: invalid command line string"))
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

		context("when output slicing fails", func() {
			it.Before(func() {
				slicer.SliceCall.Returns.Err = errors.New("some-error")
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
