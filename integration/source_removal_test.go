package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testSourceRemoval(t *testing.T, context spec.G, it spec.S) {
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

	context("when building an app", func() {
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

		it("removes source code from /workspace and replaces it with the built app", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "source_3_1_app"))
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

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Executing build process",
				MatchRegexp(`    Running 'dotnet publish \/workspace --configuration Release --runtime ubuntu\.18\.04-x64 --self-contained false --output \/tmp\/dotnet-publish-output\d+`),
			))

			Expect(logs).To(ContainLines(
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				"",
				"  Dividing build output into layers to optimize cache reuse",
				"",
				"  Generating SBOM for /workspace",
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				"",
				"  Removing source code",
				"",
			))

			container, err = docker.Container.Run.
				WithEntrypoint("launcher").
				WithCommand("ls -al /workspace").
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(container.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(
				And(
					MatchRegexp(`-rw-r--r-- +\d+ +\w+ +cnb.*\.runtimeconfig.json`),
					Not(MatchRegexp(`-rw-r--r-- +\d+ +\w+ +cnb.*Program.cs`)),
				),
			)
		})
	})
}
