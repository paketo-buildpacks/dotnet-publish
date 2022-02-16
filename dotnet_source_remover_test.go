package dotnetpublish_test

import (
	"os"
	"path/filepath"
	"testing"

	dotnetpublish "github.com/paketo-buildpacks/dotnet-publish"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDotnetSourceRemover(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir       string
		publishOutputDir string
		sourceRemover    dotnetpublish.DotnetSourceRemover
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		publishOutputDir, err = os.MkdirTemp("", "publish-output-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(err).NotTo(HaveOccurred())

		Expect(os.WriteFile(filepath.Join(workingDir, "Program.cs"), nil, 0600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workingDir, "app.csproj"), nil, 0600)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(workingDir, ".dotnet_root"), 0700)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(workingDir, ".dotnet-root"), 0700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(publishOutputDir, "app.runtimeconfig.json"), nil, 0600)).To(Succeed())

		sourceRemover = dotnetpublish.NewDotnetSourceRemover()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
		Expect(os.RemoveAll(publishOutputDir)).To(Succeed())
	})

	context("Remove", func() {
		it("clears out all files from workingDir (except excluded ones) and replaces with contents of publishOutputDir", func() {
			err := sourceRemover.Remove(workingDir, publishOutputDir, ".dotnet-root", ".dotnet_root")
			Expect(err).NotTo(HaveOccurred())

			var files []string
			err = filepath.Walk(workingDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				files = append(files, path)

				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(files).To(ConsistOf([]string{
				filepath.Join(workingDir),
				filepath.Join(workingDir, ".dotnet_root"),
				filepath.Join(workingDir, ".dotnet-root"),
				filepath.Join(workingDir, "app.runtimeconfig.json"),
			}))
		})
	})

	context("failure cases", func() {
		context("when the paths can't be found in the working directory", func() {
			var badPatternDir string
			it.Before(func() {
				badPatternDir = "\\"
			})

			it("errrors", func() {
				err := sourceRemover.Remove(badPatternDir, publishOutputDir, ".dotnet-root", ".dotnet_root")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not glob"))
			})
		})
		context("when the paths can't be found in the output directory", func() {
			var badPatternDir string
			it.Before(func() {
				badPatternDir = "\\"
			})

			it("errrors", func() {
				err := sourceRemover.Remove(workingDir, badPatternDir, ".dotnet-root", ".dotnet_root")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not glob"))
			})
		})
		context("when a file can't be moved from output directory to working directory", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(publishOutputDir, "some-dir"), os.ModePerm)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(publishOutputDir, "some-dir", "some.dll"), nil, 0000)).To(Succeed())
			})

			it("errrors", func() {
				err := sourceRemover.Remove(workingDir, publishOutputDir, ".dotnet-root", ".dotnet_root")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to move"))
			})
		})
	})
}
