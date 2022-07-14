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

		context("given a source application with .NET Core 6", func() {
			for _, b := range config.Builders {
				var builder string = b
				context(fmt.Sprintf("with %s builder", builder), func() {
					it("should build (and rebuild) a working OCI image", func() {
						var err error
						source, err := occam.Source(filepath.Join("testdata", "source_6_app"))
						Expect(err).NotTo(HaveOccurred())

						for i := 0; i < 2; i++ {
							var logs fmt.Stringer
							image, logs, err = pack.WithNoColor().Build.
								WithBuildpacks(
									icuBuildpack,
									dotnetCoreRuntimeBuildpack,
									dotnetCoreAspNetBuildpack,
									dotnetCoreSDKBuildpack,
									buildpack,
									dotnetExecuteBuildpack,
								).
								WithBuilder(builder).
								WithEnv(map[string]string{
									"BP_DOTNET_PUBLISH_FLAGS": "--verbosity=normal",
								}).
								Execute(name, source)
							Expect(err).NotTo(HaveOccurred(), logs.String())
							images[image.ID] = ""

							Expect(logs).To(ContainLines(
								MatchRegexp(`    Running 'dotnet publish .* --verbosity=normal'`),
							))

							container, err = docker.Container.Run.
								WithEnv(map[string]string{"PORT": "8080"}).
								WithPublish("8080").
								WithPublishAll().
								Execute(image.ID)
							Expect(err).NotTo(HaveOccurred())
							containers[container.ID] = ""

							Eventually(container).Should(Serve(ContainSubstring("source_6_app")).OnPort(8080))
						}
					})
				})
			}
		})

		context("given a source application with .NET Core 3.1", func() {
			it("should build a working OCI image", func() {
				var err error
				source, err := occam.Source(filepath.Join("testdata", "source_3_1_app"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(
						icuBuildpack,
						dotnetCoreRuntimeBuildpack,
						dotnetCoreAspNetBuildpack,
						dotnetCoreSDKBuildpack,
						buildpack,
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

				Eventually(container).Should(Serve(ContainSubstring("simple_3_0_app")).OnPort(8080))
			})
		})

		context("given a steeltoe application", func() {
			it("should build a working OCI image", func() {
				var err error
				source, err := occam.Source(filepath.Join("testdata", "source_steeltoe_3.1"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(
						icuBuildpack,
						dotnetCoreRuntimeBuildpack,
						dotnetCoreAspNetBuildpack,
						dotnetCoreSDKBuildpack,
						buildpack,
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

				Eventually(container).Should(Serve(ContainSubstring("value")).WithEndpoint("/api/values/6").OnPort(8080))
			})
		})

		context("given a simple webapi app with swagger dependency", func() {
			it("should build a working OCI image", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "source-3.1-aspnet-with-public-nuget"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(
						icuBuildpack,
						dotnetCoreRuntimeBuildpack,
						dotnetCoreAspNetBuildpack,
						dotnetCoreSDKBuildpack,
						buildpack,
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

				Eventually(container).Should(Serve(ContainSubstring("<title>Swagger UI</title>")).WithEndpoint("/swagger/index.html").OnPort(8080))
			})
		})

		context("when app source changes, NuGet packages are unchanged", func() {
			it("does not reuse cached app layer", func() {
				var err error
				source, err := occam.Source(filepath.Join("testdata", "source-3.1-aspnet-with-public-nuget"))
				Expect(err).NotTo(HaveOccurred())

				build := pack.WithNoColor().Build.
					WithBuildpacks(
						icuBuildpack,
						dotnetCoreRuntimeBuildpack,
						dotnetCoreAspNetBuildpack,
						dotnetCoreSDKBuildpack,
						buildpack,
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

				Eventually(container).Should(Serve(ContainSubstring("My API V1")).WithEndpoint("/swagger/index.html").OnPort(8080))
				file, err := os.Open(filepath.Join(source, "Startup.cs"))
				Expect(err).NotTo(HaveOccurred())

				contents, err := io.ReadAll(file)
				Expect(err).NotTo(HaveOccurred())

				contents = bytes.Replace(contents, []byte("My API V1"), []byte("My Cool V1 API"), 1)

				Expect(os.WriteFile(filepath.Join(source, "Startup.cs"), contents, os.ModePerm)).To(Succeed())
				file.Close()

				modified, err := os.Open(filepath.Join(source, "Startup.cs"))
				Expect(err).NotTo(HaveOccurred())

				contents, err = io.ReadAll(modified)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(ContainSubstring("My Cool V1 API"))
				modified.Close()

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

				Eventually(container).Should(Serve(ContainSubstring("My Cool V1 API")).WithEndpoint("/swagger/index.html").OnPort(8080))
			})
		})

		context("given a .NET Core angular application", func() {
			it("should build a working OCI image", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "angular_msbuild"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(
						nodeEngineBuildpack,
						icuBuildpack,
						dotnetCoreRuntimeBuildpack,
						dotnetCoreAspNetBuildpack,
						dotnetCoreSDKBuildpack,
						buildpack,
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

				Eventually(container).Should(Serve(ContainSubstring("Loading...")).OnPort(8080))
			})
		})
	})
}
