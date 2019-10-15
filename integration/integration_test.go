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
		err        error
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
		app, err = dagger.NewPack(
			filepath.Join("testdata", "source-2.2-aspnet"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, aspnetURI, sdkURI, buildURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("./source-2.2-aspnet --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("Welcome"))
	})

	it("should build a working OCI image for a simple 2.1 app with aspnet dependencies", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "source-2.1-aspnet"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, aspnetURI, sdkURI, buildURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet source-2.1-aspnet.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("source_2._1_aspnet"))
	})

	it("should build a working OCI image for a simple 2.2 webapi with swagger dependency", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "source-2.2-aspnet-with-public-nuget"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, aspnetURI, sdkURI, buildURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet source-2.2-aspnet-with-public-nuget.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/swagger/index.html")
			return body
		}).Should(ContainSubstring("SourceWithNuget"))
	})

	// TODO - Needs NPM, maybe move to end to end intergrations
	it.Pend("should build a working OCI image for a angular dotnet 2.1 application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "angular_msbuild_dotnet_2.1"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, aspnetURI, sdkURI, buildURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet angular_msbuild_dotnet_2.1.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("Hello, world from Dotnet Core 2.1"))
	})

	// TODO - Requires change to SDK symlinking
	it.Pend("should build a working OCI image for a console app", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "console_app"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, sdkURI, buildURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet console_app.dll")).To(Succeed())

		Eventually(func() string {
			body, _ := app.Logs()
			return body
		}).Should(ContainSubstring("Hello World!"))
	})

	it("should build a working OCI image for a fsharp application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "fsharp_msbuild"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, aspnetURI, sdkURI, buildURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet fsharp_msbuild.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("Hello World from F#!"))
	})

	// Renamed console_app to class_lib.
	// In the .Net world console apps are standalone cli exes. Usually they are not referenced from web apps as projects.
	// Class libraries are the ones which are referenced from
	//	1. Other Classlibs (or)
	//	2. Console Apps (or)
	//	3. Web apps
	// The Build CNB should either be able to support
	//	1. *.sln files for build (or)
	//	2. Build all csproj refereneced as a project in the root or entry point csproj
	// TODO: Figure out supoported struture for multiple proj files
	it.Pend("should build a working OCI image for an with multiple project files", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "multiple_projects_msbuild/src/asp_web_app"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, aspnetURI, sdkURI, buildURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet asp_web_app.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("Hello, I'm a string!"))
	})

	// TODO: Make the app take command line arguments or set environ
	it.Pend("should build a working OCI image for a nancy kestrel msbuild application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "nancy_kestrel_msbuild_dotnet2"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, aspnetURI, sdkURI, buildURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet nancy_kestrel_msbuild_dotnet2.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("Hello from Nancy running on CoreCLR"))
	})

	it("should build a working OCI image for a simple 2.2 application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "simple_2.2_source"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, aspnetURI, sdkURI, buildURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet simple_2.2_source.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("Hello World!"))
	})

	// TODO: Templating needs to be added and logic to use RuntimeFrameworkVersion also needs to be added
	it.Pend("should build a working OCI image for a source_2.1_explicit_runtime_templated application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "source_2.1_explicit_runtime_templated"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, aspnetURI, sdkURI, buildURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet netcoreapp2.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("Hello World!"))
	})

	// TODO: This test comes for free when logic is added to use RuntimeFrameworkVersion
	it.Pend("should build a working OCI image for a source_2.1_float_runtime application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "source_2.1_float_runtime"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, aspnetURI, sdkURI, buildURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet netcoreapp2.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())
		Expect(app.BuildLogs()).To(ContainSubstring("dotnet-runtime.2.1"))
		Expect(app.BuildLogs()).To(ContainSubstring("dotnet-aspnetcore.2.1"))
		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("netcoreapp2"))
	})

	// TODO: dotnet 3.0 resource is currently malformed for cnb useage see https://www.pivotaltracker.com/story/show/169138134 for more details
	it.Pend("should build a working OCI image for a source_3_0_app application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "source_3_0_app"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, aspnetURI, sdkURI, buildURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet source_3_0_app.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("source_3_0_app"))
	})

	it("should build a working OCI image for a source_aspnetcore_all_2.1 application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "source_aspnetcore_all_2.1"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, aspnetURI, sdkURI, buildURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet netcoreapp2.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("netcoreapp2"))
	})

	it("should build a working OCI image for a source_aspnetcore_app_2.1 application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "source_aspnetcore_app_2.1"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, aspnetURI, sdkURI, buildURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		err = app.StartWithCommand("dotnet source_aspnetcore_2.1.dll --urls http://0.0.0.0:${PORT}")
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("Hello World!"))
	})

	// TODO: Requires node should be moved
	it.Pend("should build a working OCI image for a source_prerender_node application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "source_prerender_node"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, aspnetURI, sdkURI, buildURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet asp_prerender_node.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("asp_prerender_node"))
	})

	// TODO: Waiting on decision to be made on how libgdiplus will be added to the environment
	it.Pend("should build a working OCI image for a uses_libgdiplus application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "uses_libgdiplus"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, aspnetURI, sdkURI, buildURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet uses_libgdiplus.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("uses_libgdiplus"))
	})

	// TODO: Make sure that "." does not appear in final published app dll by scrubbing . from assembly name is csproj
	it.Pend("should build a working OCI image for a with_dot_in_name application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "with_dot_in_name"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, aspnetURI, sdkURI, buildURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet with_dot_in_name.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("with_dot_in_name"))
	})
}
