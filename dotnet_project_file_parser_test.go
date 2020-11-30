package dotnetpublish_test

import (
	"io/ioutil"
	"os"
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

	context("ASPNetIsRequired", func() {
		var path string

		it.Before(func() {
			file, err := ioutil.TempFile("", "app.csproj")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			path = file.Name()
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		context("when project SDK is Microsoft.NET.Sdk.Web", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte(`<Project Sdk="Microsoft.NET.Sdk.Web"></Project>`), 0600)).To(Succeed())
			})

			it("returns true", func() {
				needsAspnt, err := parser.ASPNetIsRequired(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(needsAspnt).To(BeTrue())
			})
		})

		context("when project PackageReference is Microsoft.AspNetCore.App", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte(`
<Project Sdk="Microsoft.NET.Sdk">
<ItemGroup>
	<PackageReference Include="Microsoft.AspNetCore.App"/>
</ItemGroup>
</Project>
`), 0600)).To(Succeed())
			})

			it("returns true", func() {
				needsAspnt, err := parser.ASPNetIsRequired(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(needsAspnt).To(BeTrue())
			})
		})

		context("when project PackageReference is Microsoft.AspNetCore.All", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte(`
<Project Sdk="Microsoft.NET.Sdk">
<ItemGroup>
	<PackageReference Include="Microsoft.AspNetCore.All"/>
</ItemGroup>
</Project>
`), 0600)).To(Succeed())
			})

			it("returns true", func() {
				needsAspnt, err := parser.ASPNetIsRequired(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(needsAspnt).To(BeTrue())
			})
		})

		context("failure cases", func() {
			context("when the file can not be opened", func() {
				it.Before(func() {
					Expect(os.RemoveAll(path)).To(Succeed())
				})

				it("errors", func() {
					_, err := parser.ASPNetIsRequired(path)
					Expect(err.Error()).To(ContainSubstring("failed to open"))
				})
			})

			context("when the file can not be decoded", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(path, []byte("%%%"), 0644)).To(Succeed())
				})

				it("errors", func() {
					_, err := parser.ASPNetIsRequired(path)
					Expect(err.Error()).To(ContainSubstring("failed to decode"))
				})
			})
		})
	})

	context("NodeIsRequired", func() {
		var path string

		it.Before(func() {
			file, err := ioutil.TempFile("", "app.csproj")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			path = file.Name()
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		context("when project includes target commands that invoke node", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte(`
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
				Expect(ioutil.WriteFile(path, []byte(`
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
				Expect(ioutil.WriteFile(path, []byte(`
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
					Expect(ioutil.WriteFile(path, []byte("%%%"), 0644)).To(Succeed())
				})

				it("errors", func() {
					_, err := parser.NodeIsRequired(path)
					Expect(err.Error()).To(ContainSubstring("failed to decode"))
				})
			})

		})
	})

	context("NPMIsRequired", func() {
		var path string

		it.Before(func() {
			file, err := ioutil.TempFile("", "app.csproj")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			path = file.Name()
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		context("when project includes target commands that invoke npm", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte(`
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
				Expect(ioutil.WriteFile(path, []byte(`
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
	})
}
