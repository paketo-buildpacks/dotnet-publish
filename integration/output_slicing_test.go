package integration_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testOutputSlicing(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when rebuilding an app", func() {
		var (
			image  occam.Image
			images map[string]string
			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			images = make(map[string]string)
		})

		it.After(func() {
			for id := range images {
				Expect(docker.Image.Remove.Execute(id)).To(Succeed())
			}
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		context("when app source changes, NuGet packages are unchanged", func() {
			it("reuses package layers, adds a new app layer", func() {
				var err error
				source, err := occam.Source(filepath.Join("testdata", "source_6_aspnet_nuget"))
				Expect(err).NotTo(HaveOccurred())

				build := pack.WithNoColor().Build.
					WithBuildpacks(
						nodeEngineBuildpack,
						icuBuildpack,
						dotnetCoreSDKBuildpack,
						buildpack,
						dotnetCoreAspNetRuntimeBuildpack,
						dotnetExecuteBuildpack,
					)

				var logs fmt.Stringer
				image, logs, err = build.Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())
				images[image.ID] = ""

				file, err := os.Open(filepath.Join(source, "Program.cs"))
				Expect(err).NotTo(HaveOccurred())

				contents, err := io.ReadAll(file)
				Expect(err).NotTo(HaveOccurred())

				contents = bytes.Replace(contents, []byte("My API V1"), []byte("My Cool V1 API"), 1)

				Expect(os.WriteFile(filepath.Join(source, "Program.cs"), contents, os.ModePerm)).To(Succeed())
				file.Close()

				modified, err := os.Open(filepath.Join(source, "Program.cs"))
				Expect(err).NotTo(HaveOccurred())

				contents, err = io.ReadAll(modified)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(ContainSubstring("My Cool V1 API"))
				modified.Close()

				image, logs, err = build.Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())
				images[image.ID] = ""
				Expect(logs).To(ContainLines(
					"Reused 1/2 app layer(s)",
					"Added 1/2 app layer(s)",
				), logs.String())
			})
		})
	})
}
