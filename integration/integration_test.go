package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/dagger"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	bpDir, aspnetURI, runtimeURI, sdkURI, buildURI string
)

var suite = spec.New("Integration", spec.Report(report.Terminal{}))

func init() {
	suite("Integration", testIntegration)
}

func TestIntegration(t *testing.T) {
	var err error
	Expect := NewWithT(t).Expect
	bpDir, err = dagger.FindBPRoot()
	Expect(err).NotTo(HaveOccurred())

	buildURI, err = dagger.PackageBuildpack(bpDir)
	Expect(err).ToNot(HaveOccurred())
	defer dagger.DeleteBuildpack(buildURI)

	sdkURI, err = dagger.GetLatestBuildpack("dotnet-core-sdk-cnb")
	Expect(err).ToNot(HaveOccurred())
	defer dagger.DeleteBuildpack(sdkURI)

	aspnetURI, err = dagger.GetLatestBuildpack("dotnet-core-aspnet-cnb")
	Expect(err).ToNot(HaveOccurred())
	defer dagger.DeleteBuildpack(aspnetURI)

	runtimeURI, err = dagger.GetLatestBuildpack("dotnet-core-runtime-cnb")
	Expect(err).ToNot(HaveOccurred())
	defer dagger.DeleteBuildpack(runtimeURI)

	suite.Run(t)
}

func testIntegration(t *testing.T, _ spec.G, it spec.S) {
	var (
		Expect     func(interface{}, ...interface{}) Assertion
		Eventually func(interface{}, ...interface{}) AsyncAssertion
		app        *dagger.App
	)

	it.Before(func() {
		Expect = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually
	})

	it.After(func() {
		if app != nil {
			_ = app.Destroy()
		}
	})

	it("should build a working OCI image for a simple 2.2 app with aspnet dependencies", func() {
		app, err := dagger.PackBuild(filepath.Join("testdata", "source-2.2-aspnet"), runtimeURI, aspnetURI, sdkURI, buildURI)
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("./source-2.2-aspnet --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("Welcome"))
	})

	it("should build a working OCI image for a simple 2.1 app with aspnet dependencies", func() {
		app, err := dagger.PackBuild(filepath.Join("testdata", "source-2.1-aspnet"), runtimeURI, aspnetURI, sdkURI, buildURI)
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet source-2.1-aspnet.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("source_2._1_aspnet"))
	})


	it("should build a working OCI image for a simple 2.2 webapi with swagger dependency", func() {
		app, err := dagger.PackBuild(filepath.Join("testdata", "source-2.2-aspnet-with-public-nuget"), runtimeURI, aspnetURI, sdkURI, buildURI)
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet source-2.2-aspnet-with-public-nuget.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/swagger/index.html")
			return body
		}).Should(ContainSubstring("SourceWithNuget"))
	})
}
