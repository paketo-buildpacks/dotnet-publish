package dotnetpublish_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	dotnetpublish "github.com/paketo-buildpacks/dotnet-publish"
	"github.com/sclevine/spec"
)

func testProjectFileParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		parser dotnetpublish.ProjectFileParser
	)

	it.Before(func() {
		parser = dotnetpublish.NewProjectFileParser()
	})

	context("FindProjectFile", func() {
		var path string
		it.Before(func() {
			var err error
			path, err = os.MkdirTemp("", "workingDir")
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		it("returns an empty string and no error", func() {
			projectFilePath, err := parser.FindProjectFile(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(projectFilePath).To(Equal(""))
		})

		context("when there is a csproj", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(path, "app.csproj"), nil, 0600)).To(Succeed())
			})

			it("returns the path to it", func() {
				projectFilePath, err := parser.FindProjectFile(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(projectFilePath).To(Equal(filepath.Join(path, "app.csproj")))
			})
		})

		context("when there is an fsproj", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(path, "app.fsproj"), nil, 0600)).To(Succeed())
			})

			it("returns the path to it", func() {
				projectFilePath, err := parser.FindProjectFile(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(projectFilePath).To(Equal(filepath.Join(path, "app.fsproj")))
			})
		})

		context("when there is a vbproj", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(path, "app.vbproj"), nil, 0600)).To(Succeed())
			})

			it("returns the path to it", func() {
				projectFilePath, err := parser.FindProjectFile(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(projectFilePath).To(Equal(filepath.Join(path, "app.vbproj")))
			})
		})

		context("failure cases", func() {
			context("when file pattern matching fails", func() {
				it("returns the error", func() {
					_, err := parser.FindProjectFile(`\`)
					Expect(err).To(MatchError("syntax error in pattern"))
				})
			})

		})
	})

	context("ParseVersion", func() {
		var path string

		it.Before(func() {
			file, err := os.CreateTemp("", "app.csproj")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			path = file.Name()
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		context("when RuntimeFrameworkVersion is set", func() {
			it.Before(func() {
				Expect(os.WriteFile(path, []byte(`
					<Project>
					  <PropertyGroup>
              <RuntimeFrameworkVersion>1.2.3</RuntimeFrameworkVersion>
            </PropertyGroup>
					</Project>
				`), 0600)).To(Succeed())
			})

			it("returns the version", func() {
				version, err := parser.ParseVersion(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(version).To(Equal("1.2.3"))
			})
		})

		context("when TargetFramework is set to net<x>.<y>", func() {
			it.Before(func() {
				Expect(os.WriteFile(path, []byte(`
					<Project>
					  <PropertyGroup>
              <TargetFramework>net1.2</TargetFramework>
            </PropertyGroup>
					</Project>
				`), 0600)).To(Succeed())
			})

			it("returns the version", func() {
				version, err := parser.ParseVersion(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(version).To(Equal("1.2.0"))
			})
		})

		context("when TargetFramework is set to net<x>.<y>-platform", func() {
			it.Before(func() {
				Expect(os.WriteFile(path, []byte(`
					<Project>
					  <PropertyGroup>
              <TargetFramework>net1.2-windows</TargetFramework>
            </PropertyGroup>
					</Project>
				`), 0600)).To(Succeed())
			})

			it("returns the version", func() {
				version, err := parser.ParseVersion(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(version).To(Equal("1.2.0"))
			})
		})

		context("when TargetFramework is set to netcoreapp<x>.<y>", func() {
			it.Before(func() {
				Expect(os.WriteFile(path, []byte(`
					<Project>
					  <PropertyGroup>
              <TargetFramework>netcoreapp1.2</TargetFramework>
            </PropertyGroup>
					</Project>
				`), 0600)).To(Succeed())
			})

			it("returns the version", func() {
				version, err := parser.ParseVersion(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(version).To(Equal("1.2.0"))
			})
		})

		context("failure cases", func() {
			context("when the file can not be opened", func() {
				it.Before(func() {
					Expect(os.RemoveAll(path)).To(Succeed())
				})

				it("errors", func() {
					_, err := parser.ParseVersion(path)
					Expect(err.Error()).To(ContainSubstring("failed to read project file"))
				})
			})

			context("when the file can not be decoded", func() {
				it.Before(func() {
					Expect(os.WriteFile(path, []byte("%%%"), 0644)).To(Succeed())
				})

				it("errors", func() {
					_, err := parser.ParseVersion(path)
					Expect(err.Error()).To(ContainSubstring("failed to parse project file"))
				})
			})

			context("when the file can not be decoded", func() {
				it.Before(func() {
					Expect(os.WriteFile(path, []byte(`
					<Project>
					  <PropertyGroup>
              <TargetFramework>bad content</TargetFramework>
            </PropertyGroup>
					</Project>
				`), 0600)).To(Succeed())
				})

				it("errors", func() {
					_, err := parser.ParseVersion(path)
					Expect(err.Error()).To(ContainSubstring("failed to find version in project file: missing or invalid TargetFramework property"))
				})
			})
		})
	})

	context("NodeIsRequired", func() {
		var path string

		it.Before(func() {
			file, err := os.CreateTemp("", "app.csproj")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			path = file.Name()
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		context("when project includes target commands that invoke node", func() {
			it.Before(func() {
				Expect(os.WriteFile(path, []byte(`
					<Project>
						<Target Name="first-target">
							<Exec Command="echo hello" />
						</Target>
						<Target Name="second-target">
							<Exec Command="node --version" />
						</Target>
					</Project>
				`), 0600)).To(Succeed())
			})

			it("returns true", func() {
				needNode, err := parser.NodeIsRequired(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(needNode).To(BeTrue())
			})
		})

		context("when project does NOT include target commands that invoke node", func() {
			it.Before(func() {
				Expect(os.WriteFile(path, []byte(`
					<Project>
						<Target Name="first-target">
							<Exec Command="echo hello" />
						</Target>
						<Target Name="second-target">
							<Exec Command="echo goodbye" />
						</Target>
					</Project>
				`), 0600)).To(Succeed())
			})

			it("returns false", func() {
				needNode, err := parser.NodeIsRequired(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(needNode).To(BeFalse())
			})
		})

		context("when project includes target commands that invoke npm", func() {
			it.Before(func() {
				Expect(os.WriteFile(path, []byte(`
					<Project>
						<Target Name="first-target">
							<Exec Command="echo hello" />
						</Target>
						<Target Name="second-target">
							<Exec Command="npm install" />
						</Target>
					</Project>
				`), 0600)).To(Succeed())
			})

			it("returns true", func() {
				needNode, err := parser.NodeIsRequired(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(needNode).To(BeTrue())
			})
		})

		context("failure cases", func() {
			context("when the file can not be opened", func() {
				it.Before(func() {
					Expect(os.RemoveAll(path)).To(Succeed())
				})

				it("errors", func() {
					_, err := parser.NodeIsRequired(path)
					Expect(err.Error()).To(ContainSubstring("failed to open"))
				})
			})

			context("when the file can not be decoded", func() {
				it.Before(func() {
					Expect(os.WriteFile(path, []byte("%%%"), 0644)).To(Succeed())
				})

				it("errors", func() {
					_, err := parser.NodeIsRequired(path)
					Expect(err.Error()).To(ContainSubstring("failed to decode"))
				})
			})

		})
	})

	context("NPMIsRequired", func() {
		var path, importPath, targetPath string

		it.Before(func() {
			file, err := os.CreateTemp("", "app.csproj")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			importFile, err := os.CreateTemp("", "import.csproj")
			Expect(err).NotTo(HaveOccurred())
			defer importFile.Close()

			targetFile, err := os.CreateTemp("", "target.csproj")
			Expect(err).NotTo(HaveOccurred())
			defer targetFile.Close()

			importPath = importFile.Name()
			targetPath = targetFile.Name()
			path = file.Name()
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		context("when project includes target commands that invoke npm", func() {
			it.Before(func() {
				Expect(os.WriteFile(path, []byte(`
					<Project>
						<Target Name="first-target">
							<Exec Command="echo hello" />
						</Target>
						<Target Name="second-target">
							<Exec Command="npm install" />
						</Target>
					</Project>
				`), 0600)).To(Succeed())
			})

			it("returns true", func() {
				needNode, err := parser.NPMIsRequired(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(needNode).To(BeTrue())
			})
		})

		context("when project does NOT include target commands that invoke node", func() {
			it.Before(func() {
				Expect(os.WriteFile(path, []byte(`
					<Project>
						<Target Name="first-target">
							<Exec Command="echo hello" />
						</Target>
						<Target Name="second-target">
							<Exec Command="echo goodbye" />
						</Target>
					</Project>
				`), 0600)).To(Succeed())
			})

			it("returns false", func() {
				needNode, err := parser.NPMIsRequired(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(needNode).To(BeFalse())
			})
		})

		context("when project includes target commands that invoke npm in a different file", func() {
			it.Before(func() {
				Expect(os.WriteFile(targetPath, []byte(`
					<Project>
						<Target Name="first-target">
							<Exec Command="echo hello" />
						</Target>
						<Target Name="second-target">
							<Exec Command="npm install" />
						</Target>
					</Project>
				`), 0600)).To(Succeed())

				Expect(os.WriteFile(importPath, []byte(fmt.Sprintf(`
					<Project>
          				<Import Project="%s" />
					</Project>
				`, targetPath)), 0600)).To(Succeed())

				Expect(os.WriteFile(path, []byte(fmt.Sprintf(`
					<Project>
						<ItemGroup>
						<ProjectReference Include="%s" />
						</ItemGroup>
					</Project>
				`, importPath)), 0600)).To(Succeed())
			})

			it("returns true", func() {
				needNode, err := parser.NPMIsRequired(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(needNode).To(BeTrue())
			})
		})
	})
}
