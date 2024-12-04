package integration_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testDefaultApps(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually
		pack       occam.Pack
		docker     occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()
	})

	context("when building a .NET Core app", func() {
		var (
			image      occam.Image
			images     map[string]string
			container  occam.Container
			containers map[string]string
			name       string
			source     string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			containers = make(map[string]string)
			images = make(map[string]string)
		})

		it.After(func() {
			for id := range containers {
				Expect(docker.Container.Remove.Execute(id)).To(Succeed())
			}
			for id := range images {
				Expect(docker.Image.Remove.Execute(id)).To(Succeed())
			}
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		context("given a source application with .NET Core 8", func() {
			it("should build (and rebuild) a working OCI image", func() {
				var err error
				source, err := occam.Source(filepath.Join("testdata", "source_8"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(
						icuBuildpack,
						vsdbgBuildpack,
						dotnetCoreSDKBuildpack,
						buildpack,
						dotnetCoreAspNetRuntimeBuildpack,
						dotnetExecuteBuildpack,
					).
					WithEnv(map[string]string{
						"BP_DOTNET_PUBLISH_FLAGS": "--verbosity=normal",
						"BP_DEBUG_ENABLED":        "true",
					}).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())
				images[image.ID] = ""

				Expect(logs).To(ContainLines(
					MatchRegexp(`    Running 'dotnet publish .* \-\-configuration Debug .* \-\-verbosity=normal'`),
				))

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())
				containers[container.ID] = ""

				Eventually(container).Should(Serve(ContainSubstring("source_8")).OnPort(8080))
			})
		})

		context("given a source application with .NET 9", func() {
			it("should build a working OCI image", func() {
				var err error
				source, err := occam.Source(filepath.Join("testdata", "source_9"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(
						icuBuildpack,
						dotnetCoreSDKBuildpack,
						buildpack,
						dotnetCoreAspNetRuntimeBuildpack,
						dotnetExecuteBuildpack,
					).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())
				images[image.ID] = ""

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())
				containers[container.ID] = ""

				Eventually(container).Should(Serve(ContainSubstring("source_9")).OnPort(8080))
			})
		})

		context("given a steeltoe application", func() {
			it("should build a working OCI image", func() {
				var err error
				source, err := occam.Source(filepath.Join("testdata", "source_steeltoe"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(
						icuBuildpack,
						dotnetCoreSDKBuildpack,
						buildpack,
						dotnetCoreAspNetRuntimeBuildpack,
						dotnetExecuteBuildpack,
					).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())
				images[image.ID] = ""

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())
				containers[container.ID] = ""

				Eventually(container).Should(Serve(ContainSubstring("Chilly")).WithEndpoint("/weatherforecast").OnPort(8080))
			})
		})

		context("given a simple webapi app with swagger dependency", func() {
			it("should build a working OCI image", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "source_aspnet_nuget_configuration"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(
						icuBuildpack,
						dotnetCoreSDKBuildpack,
						buildpack,
						dotnetCoreAspNetRuntimeBuildpack,
						dotnetExecuteBuildpack,
					).
					WithEnv(map[string]string{
						"BP_LOG_LEVEL": "DEBUG",
						"BP_DOTNET_DISABLE_BUILDPACK_OUTPUT_SLICING": "true",
					}).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())
				images[image.ID] = ""

				Expect(logs).To(ContainLines(
					"  Build configuration:",
					// General matcher since env vars are extracted from map in different order each time
					MatchRegexp("    BP_.*: .*"),
				))

				Expect(logs).To(ContainLines(
					"  Skipping output slicing",
					"",
				))

				Expect(logs).To(ContainLines(
					"  Setting up layer 'nuget-cache'",
					"    Available at launch: false",
					"    Available to other buildpacks: false",
					"    Cached for rebuilds: true",
					"",
				))

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())
				containers[container.ID] = ""

				Eventually(container).Should(Serve(ContainSubstring("Chilly")).WithEndpoint("/weatherforecast").OnPort(8080))
			})
		})

		context("when app source changes, NuGet packages are unchanged", func() {
			it("does not reuse cached app layer", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "source_aspnet_nuget_configuration"))
				Expect(err).NotTo(HaveOccurred())

				build := pack.WithNoColor().Build.
					WithBuildpacks(
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

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())
				containers[container.ID] = ""

				Eventually(container).Should(Serve(ContainSubstring("Chilly")).WithEndpoint("/weatherforecast").OnPort(8080))

				file, err := os.Open(filepath.Join(source, "Program.cs"))
				Expect(err).NotTo(HaveOccurred())

				contents, err := io.ReadAll(file)
				Expect(err).NotTo(HaveOccurred())

				contents = bytes.Replace(contents, []byte("Chilly"), []byte("Replacement"), 1)

				Expect(os.WriteFile(filepath.Join(source, "Program.cs"), contents, os.ModePerm)).To(Succeed())
				file.Close()

				image, logs, err = build.Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())
				images[image.ID] = ""

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())
				containers[container.ID] = ""

				Eventually(container).Should(Serve(ContainSubstring("Replacement")).WithEndpoint("/weatherforecast").OnPort(8080))
			})
		})

		context("given a .NET Core angular application", func() {
			var sbomDir string

			it.Before(func() {
				var err error
				sbomDir, err = os.MkdirTemp("", "sbom")
				Expect(err).NotTo(HaveOccurred())
			})

			it.After(func() {
				Expect(os.RemoveAll(sbomDir)).To(Succeed())
			})

			it("should build a working OCI image", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "angular_msbuild"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(
						nodeEngineBuildpack,
						icuBuildpack,
						dotnetCoreSDKBuildpack,
						buildpack,
						dotnetCoreAspNetRuntimeBuildpack,
						dotnetExecuteBuildpack,
					).
					WithSBOMOutputDir(sbomDir).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())
				images[image.ID] = ""

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())
				containers[container.ID] = ""

				Eventually(container).Should(Serve(ContainSubstring("Loading...")).OnPort(8080))

				// check that all expected SBOM files are present
				Expect(filepath.Join(sbomDir, "sbom", "build", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "sbom.cdx.json")).To(BeARegularFile())
				Expect(filepath.Join(sbomDir, "sbom", "build", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "sbom.spdx.json")).To(BeARegularFile())
				Expect(filepath.Join(sbomDir, "sbom", "build", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "sbom.syft.json")).To(BeARegularFile())

				// check an SBOM file to make sure it has an entry for an app node module
				contents, err := os.ReadFile(filepath.Join(sbomDir, "sbom", "build", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "sbom.cdx.json"))
				Expect(err).NotTo(HaveOccurred())

				Expect(string(contents)).To(ContainSubstring(`"name": "yaml"`))
			})
		})
	})
}
