package integration_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
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
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when building a .NET Core app", func() {
		var (
			image     occam.Image
			container occam.Container
			name      string
			source    string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		context("given a source application with .NET Core 6", func() {
			it("should build a working OCI image", func() {
				var err error
				source, err := occam.Source(filepath.Join("testdata", "source_6_app"))
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
					WithEnv(map[string]string{
						"BP_DOTNET_PUBLISH_FLAGS": "--verbosity=normal",
					}).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())

				Expect(logs).To(ContainLines(
					MatchRegexp(`    Running 'dotnet publish .* --verbosity=normal'`),
				))

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())

				response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort("8080")))
				Expect(err).NotTo(HaveOccurred())
				defer response.Body.Close()

				Expect(response.StatusCode).To(Equal(http.StatusOK))

				content, err := ioutil.ReadAll(response.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("source_6_app"))
			})
		})

		context("given a source application with .NET Core 5", func() {
			it("should build a working OCI image", func() {
				var err error
				source, err := occam.Source(filepath.Join("testdata", "source_5_app"))
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
					WithEnv(map[string]string{
						"BP_DOTNET_PUBLISH_FLAGS": "--verbosity=normal",
					}).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())

				Expect(logs).To(ContainLines(
					MatchRegexp(`    Running 'dotnet publish .* --verbosity=normal'`),
				))

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())

				response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort("8080")))
				Expect(err).NotTo(HaveOccurred())
				defer response.Body.Close()

				Expect(response.StatusCode).To(Equal(http.StatusOK))

				content, err := ioutil.ReadAll(response.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("source_5_app"))
			})
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

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())

				response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort("8080")))
				Expect(err).NotTo(HaveOccurred())
				defer response.Body.Close()

				Expect(response.StatusCode).To(Equal(http.StatusOK))

				content, err := ioutil.ReadAll(response.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("simple_3_0_app"))
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

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())

				response, err := http.Get(fmt.Sprintf("http://localhost:%s/api/values/6", container.HostPort("8080")))
				Expect(err).NotTo(HaveOccurred())
				defer response.Body.Close()

				Expect(response.StatusCode).To(Equal(http.StatusOK))

				content, err := ioutil.ReadAll(response.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("value"))
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

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() int {
					response, _ := http.Get(fmt.Sprintf("http://localhost:%s/swagger/index.html", container.HostPort("8080")))
					if response != nil {
						defer response.Body.Close()
						return response.StatusCode
					}
					return -1
				}).Should(Equal(http.StatusOK))

				response, err := http.Get(fmt.Sprintf("http://localhost:%s/swagger/index.html", container.HostPort("8080")))
				Expect(err).NotTo(HaveOccurred())
				defer response.Body.Close()

				content, err := ioutil.ReadAll(response.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("<title>Swagger UI</title>"))
			})
		})

		context("given a .NET Core angular application", func() {
			it("should build a working OCI image", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "angular_msbuild_dotnet_5"))
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

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())

				response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort("8080")))
				Expect(err).NotTo(HaveOccurred())
				defer response.Body.Close()

				Expect(response.StatusCode).To(Equal(http.StatusOK))

				content, err := ioutil.ReadAll(response.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("Loading..."))
			})
		})
	})
}
