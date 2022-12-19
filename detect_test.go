package dotnetpublish_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	dotnetpublish "github.com/paketo-buildpacks/dotnet-publish"
	"github.com/paketo-buildpacks/dotnet-publish/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		projectParser      *fakes.ProjectParser
		buildpackYMLParser *fakes.BuildpackYMLParser
		workingDir         string
		detect             packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		projectParser = &fakes.ProjectParser{}
		projectParser.FindProjectFileCall.Returns.String = filepath.Join(workingDir, "app.csproj")
		projectParser.ParseVersionCall.Returns.String = "6.0.0"

		buildpackYMLParser = &fakes.BuildpackYMLParser{}

		detect = dotnetpublish.Detect(
			dotnetpublish.Configuration{},
			projectParser,
			buildpackYMLParser,
		)
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("returns a build plan", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: workingDir,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "dotnet-application"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "dotnet-sdk",
						Metadata: dotnetpublish.BuildPlanMetadata{
							Version:       "6.0.*",
							VersionSource: "app.csproj",
							Build:         true,
						},
					},
					{
						Name: "icu",
						Metadata: dotnetpublish.BuildPlanMetadata{
							Build: true,
						},
					},
				},
			},
		}))

		Expect(buildpackYMLParser.ParseProjectPathCall.Receives.Path).To(Equal(filepath.Join(workingDir, "buildpack.yml")))
		Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(workingDir))
		Expect(projectParser.ParseVersionCall.Receives.Path).To(Equal(filepath.Join(workingDir, "app.csproj")))
		Expect(projectParser.NodeIsRequiredCall.Receives.Path).To(Equal(filepath.Join(workingDir, "app.csproj")))
		Expect(projectParser.NPMIsRequiredCall.Receives.Path).To(Equal(filepath.Join(workingDir, "app.csproj")))
	})

	context("when node is required", func() {
		it.Before(func() {
			projectParser.NodeIsRequiredCall.Returns.Bool = true
		})

		it("requires node in the build plan", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: "dotnet-application"},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: "dotnet-sdk",
							Metadata: dotnetpublish.BuildPlanMetadata{
								Version:       "6.0.*",
								VersionSource: "app.csproj",
								Build:         true,
							},
						},
						{
							Name: "icu",
							Metadata: dotnetpublish.BuildPlanMetadata{
								Build: true,
							},
						},
						{
							Name: "node",
							Metadata: dotnetpublish.BuildPlanMetadata{
								Build: true,
							},
						},
					},
				},
			}))

			Expect(buildpackYMLParser.ParseProjectPathCall.Receives.Path).To(Equal(filepath.Join(workingDir, "buildpack.yml")))
			Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(workingDir))
			Expect(projectParser.ParseVersionCall.Receives.Path).To(Equal(filepath.Join(workingDir, "app.csproj")))
			Expect(projectParser.NodeIsRequiredCall.Receives.Path).To(Equal(filepath.Join(workingDir, "app.csproj")))
			Expect(projectParser.NPMIsRequiredCall.Receives.Path).To(Equal(filepath.Join(workingDir, "app.csproj")))
		})
	})

	context("when npm is required", func() {
		it.Before(func() {
			projectParser.NodeIsRequiredCall.Returns.Bool = true
			projectParser.NPMIsRequiredCall.Returns.Bool = true
		})

		it("requires node in the build plan", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: "dotnet-application"},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: "dotnet-sdk",
							Metadata: dotnetpublish.BuildPlanMetadata{
								Version:       "6.0.*",
								VersionSource: "app.csproj",
								Build:         true,
							},
						},
						{
							Name: "icu",
							Metadata: dotnetpublish.BuildPlanMetadata{
								Build: true,
							},
						},
						{
							Name: "node",
							Metadata: dotnetpublish.BuildPlanMetadata{
								Build: true,
							},
						},
						{
							Name: "npm",
							Metadata: dotnetpublish.BuildPlanMetadata{
								Build: true,
							},
						},
					},
				},
			}))

			Expect(buildpackYMLParser.ParseProjectPathCall.Receives.Path).To(Equal(filepath.Join(workingDir, "buildpack.yml")))
			Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(workingDir))
			Expect(projectParser.ParseVersionCall.Receives.Path).To(Equal(filepath.Join(workingDir, "app.csproj")))
			Expect(projectParser.NodeIsRequiredCall.Receives.Path).To(Equal(filepath.Join(workingDir, "app.csproj")))
			Expect(projectParser.NPMIsRequiredCall.Receives.Path).To(Equal(filepath.Join(workingDir, "app.csproj")))
		})
	})

	context("when the .csproj file is not at the base of the directory and project_path is set via $BP_DOTNET_PROJECT_PATH", func() {
		it.Before(func() {
			projectParser.FindProjectFileCall.Returns.String = filepath.Join(workingDir, "src/proj1", "app.csproj")
			detect = dotnetpublish.Detect(
				dotnetpublish.Configuration{ProjectPath: "src/proj1"},
				projectParser,
				buildpackYMLParser,
			)
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		it("finds the projfile and passes detection", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: "dotnet-application"},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: "dotnet-sdk",
							Metadata: dotnetpublish.BuildPlanMetadata{
								Version:       "6.0.*",
								VersionSource: "app.csproj",
								Build:         true,
							},
						},
						{
							Name: "icu",
							Metadata: dotnetpublish.BuildPlanMetadata{
								Build: true,
							},
						},
					},
				},
			}))

			Expect(buildpackYMLParser.ParseProjectPathCall.CallCount).To(Equal(0))
			Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(filepath.Join(workingDir, "src/proj1")))
			Expect(projectParser.ParseVersionCall.Receives.Path).To(Equal(filepath.Join(workingDir, "src/proj1", "app.csproj")))
			Expect(projectParser.NodeIsRequiredCall.Receives.Path).To(Equal(filepath.Join(workingDir, "src/proj1", "app.csproj")))
			Expect(projectParser.NPMIsRequiredCall.Receives.Path).To(Equal(filepath.Join(workingDir, "src/proj1", "app.csproj")))
		})
	})
	context("when the .csproj file is not at the base of the directory and project_path is set in buildpack.yml", func() {
		it.Before(func() {
			buildpackYMLParser.ParseProjectPathCall.Returns.ProjectFilePath = "src/proj1"
			projectParser.FindProjectFileCall.Returns.String = filepath.Join(workingDir, "src/proj1", "app.csproj")
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		it("finds the projfile and passes detection", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: "dotnet-application"},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: "dotnet-sdk",
							Metadata: dotnetpublish.BuildPlanMetadata{
								Version:       "6.0.*",
								VersionSource: "app.csproj",
								Build:         true,
							},
						},
						{
							Name: "icu",
							Metadata: dotnetpublish.BuildPlanMetadata{
								Build: true,
							},
						},
					},
				},
			}))

			Expect(buildpackYMLParser.ParseProjectPathCall.Receives.Path).To(Equal(filepath.Join(workingDir, "buildpack.yml")))
			Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(filepath.Join(workingDir, "src/proj1")))
			Expect(projectParser.ParseVersionCall.Receives.Path).To(Equal(filepath.Join(workingDir, "src/proj1", "app.csproj")))
			Expect(projectParser.NodeIsRequiredCall.Receives.Path).To(Equal(filepath.Join(workingDir, "src/proj1", "app.csproj")))
			Expect(projectParser.NPMIsRequiredCall.Receives.Path).To(Equal(filepath.Join(workingDir, "src/proj1", "app.csproj")))
		})
	})

	context("failure cases", func() {
		context("when buildpack.yml cannot be parsed", func() {
			it.Before(func() {
				buildpackYMLParser.ParseProjectPathCall.Returns.Err = fmt.Errorf("parsing error")
			})

			it("fails detection", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError("failed to parse buildpack.yml: parsing error"))
			})
		})

		context("when finding project file returns an error", func() {
			it.Before(func() {
				projectParser.FindProjectFileCall.Returns.Error = errors.New("some project file error")
			})

			it("fails detection", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError("some project file error"))
			})
		})

		context("when a project file cannot be found", func() {
			it.Before(func() {
				projectParser.FindProjectFileCall.Returns.String = ""
			})

			it("fails detection", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(packit.Fail.WithMessage("no project file found")))
			})
		})

		context("when parsing for SDK version errors", func() {
			it.Before(func() {
				projectParser.ParseVersionCall.Returns.Error = errors.New("parsing-version-error")
			})

			it("errors", func() {
				_, err := detect(packit.DetectContext{WorkingDir: workingDir})
				Expect(err).To(MatchError("parsing-version-error"))
			})
		})

		context("when parsing for Node errors", func() {
			it.Before(func() {
				projectParser.NodeIsRequiredCall.Returns.Error = errors.New("parsing-error")
			})

			it("errors", func() {
				_, err := detect(packit.DetectContext{WorkingDir: workingDir})
				Expect(err).To(MatchError("parsing-error"))
			})
		})

		context("when parsing for NPM errors", func() {
			it.Before(func() {
				projectParser.NPMIsRequiredCall.Returns.Error = errors.New("parsing-error")
			})

			it("errors", func() {
				_, err := detect(packit.DetectContext{WorkingDir: workingDir})
				Expect(err).To(MatchError("parsing-error"))
			})
		})
	})
}
