package dotnetpublish_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	dotnetpublish "github.com/paketo-buildpacks/dotnet-publish"
	"github.com/sclevine/spec"
)

func testBuildpackYMLParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		path   string
		parser dotnetpublish.DotnetBuildpackYMLParser
	)

	it.Before(func() {
		file, err := ioutil.TempFile("", "buildpack.yml")
		Expect(err).NotTo(HaveOccurred())
		file.Close()

		path = file.Name()

		err = ioutil.WriteFile(path, []byte(`---
dotnet-build:
  project-path: "src/proj1"
`), os.ModePerm)

		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	it("parses the project path", func() {
		projectPath, err := parser.ParseProjectPath(path)

		Expect(err).NotTo(HaveOccurred())
		Expect(projectPath).To(Equal("src/proj1"))
	})

	context("when the buildpack.yml file does not exist", func() {
		it.Before(func() {
			Expect(os.Remove(path)).To(Succeed())
		})

		it("returns an empty path", func() {
			projectPath, err := parser.ParseProjectPath(path)

			Expect(err).NotTo(HaveOccurred())
			Expect(projectPath).To(BeEmpty())
		})
	})

	context("failure cases", func() {
		context("when the file cannot be opened", func() {
			it.Before(func() {
				Expect(os.Chmod(path, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(path, os.ModePerm)).To(Succeed())
			})

			it("returns the error", func() {
				_, err := parser.ParseProjectPath(path)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("permission denied"))
			})
		})
		context("when the file cannot be unmarshalled", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte("%%%"), 0644)).To(Succeed())
			})

			it("returns the error", func() {
				_, err := parser.ParseProjectPath(filepath.Join(path))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid buildpack.yml:"))
			})
		})
	})
}
