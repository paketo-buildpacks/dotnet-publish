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

		Expect(app.StartWithCommand("./source-2.2-aspnet --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("Welcome"))
	})

	it("should build a working OCI image for a simple 2.1 app with aspnet dependencies", func() {
		app, err := dagger.PackBuild(filepath.Join("testdata", "source-2.1-aspnet"), runtimeURI, aspnetURI, sdkURI, buildURI)
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet source-2.1-aspnet.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("source_2._1_aspnet"))
	})

	it("should build a working OCI image for a simple 2.2 webapi with swagger dependency", func() {
		app, err := dagger.PackBuild(filepath.Join("testdata", "source-2.2-aspnet-with-public-nuget"), runtimeURI, aspnetURI, sdkURI, buildURI)
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet source-2.2-aspnet-with-public-nuget.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/swagger/index.html")
			return body
		}).Should(ContainSubstring("SourceWithNuget"))
	})

	// Todo - Copied from V2 to be fixed for V3
	it("should build a working OCI image for a angular dotnet 2.1 application", func() {
		// app, err := dagger.PackBuild(filepath.Join("testdata", "angular_msbuild_dotnet_2.1"), runtimeURI, aspnetURI, sdkURI, buildURI)
		// Expect(err).ToNot(HaveOccurred())

		// Expect(app.StartWithCommand("dotnet angular_msbuild_dotnet_2.1.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		// Eventually(func() string {
		// 	body, _, _ := app.HTTPGet("/")
		// 	return body
		// }).Should(ContainSubstring("Hello, world from Dotnet Core 2.1"))
	})

	// Todo - Copied from V2 to be fixed for V3
	it("should build a working OCI image for a console app", func() {
		// app, err := dagger.PackBuild(filepath.Join("testdata", "console_app"), runtimeURI, aspnetURI, sdkURI, buildURI)
		// Expect(err).ToNot(HaveOccurred())

		// Expect(app.StartWithCommand("dotnet console_app.dll")).To(Succeed())

		// Eventually(func() string {
		// 	body, _, _ := app.Logs()
		// 	return body
		// }).Should(ContainSubstring("Hello World!"))
	})

	// Todo - Copied from V2 to be fixed for V3
	it("should build a working OCI image for msbuild self contained application", func() {
		// app, err := dagger.PackBuild(filepath.Join("testdata", "fake_supply_dotnet_app"), runtimeURI, aspnetURI, sdkURI, buildURI)
		// Expect(err).ToNot(HaveOccurred())

		// Expect(app.StartWithCommand("dotnet msbuild_self_contained.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		// Eventually(func() string {
		// 	body, _, _ := app.Logs()
		// 	return body
		// }).Should(ContainSubstring("SUPPLYING BOSH2"))

		// Eventually(func() string {
		// 	body, _, _ := app.HTTPGet("/")
		// 	return body
		// }).Should(ContainSubstring("bosh2: version 2.0.1-74fad57"))
	})


	// Todo - Copied from V2 to be fixed for V3
	it("should build a working OCI image for a static file application", func() {
		// BeforeEach(func() {
		// 	if ok, err := cutlass.ApiGreaterThan("2.65.1"); err != nil || !ok {
		// 		Skip("API version does not have multi-buildpack support")
		// 	}

		// 	app = cutlass.New(filepath.Join(bpDir, "fixtures", "fake_supply_staticfile_app_with_no_csproj_file"))
		// 	app.Buildpacks = []string{
		// 		"dotnet_core_buildpack",
		// 		"https://github.com/cloudfoundry/staticfile-buildpack/#master",
		// 	}
		// 	app.Disk = "1G"
		// })

		// It("finds the supplied dependency in the runtime container", func() {
		// 	PushAppAndConfirm(app)
		// 	Expect(app.Stdout.String()).To(ContainSubstring("Supplying Dotnet Core"))
		// 	Expect(app.GetBody("/")).To(ContainSubstring("This is an example app for Cloud Foundry that is only static HTML/JS/CSS assets."))
		// })
	})

	// Test fixture moved from V2 and succeeds
	it("should build a working OCI image for a fsharp application", func() {
		app, err := dagger.PackBuild(filepath.Join("testdata", "fsharp_msbuild"), runtimeURI, aspnetURI, sdkURI, buildURI)
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet fsharp_msbuild.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("Hello World from F#!"))
	})

	// Todo - Copied from V2 to be fixed for V3
	// Renamed console_app to class_lib.
	// In the .Net world console apps are standalone cli exes. Usually they are not referenced from web apps as projects.
	// Class libraries are the ones which are referenced from 
	//	1. Other Classlibs (or)
	//	2. Console Apps (or)
	//	3. Web apps
	// The Build CNB should either be able to support 
	//	1. *.sln files for build (or)
	//	2. Build all csproj refereneced as a project in the root or entry point csproj
	it("should build a working OCI image for a fsharp application", func() {
		// app, err := dagger.PackBuild(filepath.Join("testdata", "multiple_projects_msbuild/src/asp_web_app"), runtimeURI, aspnetURI, sdkURI, buildURI)
		// Expect(err).ToNot(HaveOccurred())

		// Expect(app.StartWithCommand("dotnet asp_web_app.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		// Eventually(func() string {
		// 	body, _, _ := app.HTTPGet("/")
		// 	return body
		// }).Should(ContainSubstring("Hello, I'm a string!"))
	})

	// Todo - Copied from V2 to be fixed for V3
	it("should build a working OCI image for a nancy kestrel msbuild application", func() {
		// app, err := dagger.PackBuild(filepath.Join("testdata", "nancy_kestrel_msbuild_dotnet2"), runtimeURI, aspnetURI, sdkURI, buildURI)
		// Expect(err).ToNot(HaveOccurred())

		// Expect(app.StartWithCommand("dotnet nancy_kestrel_msbuild_dotnet2.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		// Eventually(func() string {
		// 	body, _, _ := app.HTTPGet("/")
		// 	return body
		// }).Should(ContainSubstring("Hello from Nancy running on CoreCLR"))
	})

	// Todo - Copied from V2 to be fixed for V3
	// Target platform is Linux (Ubuntu)
	it("should build a working OCI image for a self contained msbuild application", func() {
		// app, err := dagger.PackBuild(filepath.Join("testdata", "self_contained_msbuild"), runtimeURI, aspnetURI, sdkURI, buildURI)
		// Expect(err).ToNot(HaveOccurred())

		// Expect(app.StartWithCommand("dotnet self_contained_msbuild.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		// Eventually(func() string {
		// 	body, _, _ := app.HTTPGet("/")
		// 	return body
		// }).Should(ContainSubstring("Hello from Nancy running on CoreCLR"))
	})

	// Test fixture moved from V2 and succeeds
	it("should build a working OCI image for a simple 2.2 application", func() {
		app, err := dagger.PackBuild(filepath.Join("testdata", "simple_2.2_source"), runtimeURI, aspnetURI, sdkURI, buildURI)
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet simple_2.2_source.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("Hello World!"))
	})

	// Todo - Copied from V2 to be fixed for V3
	it("should build a working OCI image for a simple brats application", func() {
		// app, err := dagger.PackBuild(filepath.Join("testdata", "simple_brats"), runtimeURI, aspnetURI, sdkURI, buildURI)
		// Expect(err).ToNot(HaveOccurred())

		// Expect(app.StartWithCommand("dotnet simple_brats.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		// Eventually(func() string {
		// 	body, _, _ := app.HTTPGet("/")
		// 	return body
		// }).Should(ContainSubstring("Hello World!"))
	})

	// Todo - Copied from V2 to be fixed for V3
	it("should build a working OCI image for a simple fsharp brats application", func() {
		// app, err := dagger.PackBuild(filepath.Join("testdata", "simple_fsharp_brats"), runtimeURI, aspnetURI, sdkURI, buildURI)
		// Expect(err).ToNot(HaveOccurred())

		// Expect(app.StartWithCommand("dotnet simple_fsharp_brats.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		// Eventually(func() string {
		// 	body, _, _ := app.HTTPGet("/")
		// 	return body
		// }).Should(ContainSubstring("Hello World!"))
	})

	// Todo - Copied from V2 to be fixed for V3
	it("should build a working OCI image for a source 2.0 application", func() {
		// app, err := dagger.PackBuild(filepath.Join("testdata", "source_2.0"), runtimeURI, aspnetURI, sdkURI, buildURI)
		// Expect(err).ToNot(HaveOccurred())

		// Expect(app.StartWithCommand("dotnet source_2.0.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		// Eventually(func() string {
		// 	body, _, _ := app.HTTPGet("/")
		// 	return body
		// }).Should(ContainSubstring("Hello World!"))
	})

	// Todo - Copied from V2 to be fixed for V3
	it("should build a working OCI image for a source_2.1_explicit_runtime_templated application", func() {
		// app, err := dagger.PackBuild(filepath.Join("testdata", "source_2.1_explicit_runtime_templated"), runtimeURI, aspnetURI, sdkURI, buildURI)
		// Expect(err).ToNot(HaveOccurred())

		// Expect(app.StartWithCommand("dotnet netcoreapp2.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		// Eventually(func() string {
		// 	body, _, _ := app.HTTPGet("/")
		// 	return body
		// }).Should(ContainSubstring("Hello World!"))
	})

	// Test fixture moved from V2 and succeeds
	it("should build a working OCI image for a source_2.1_float_runtime application", func() {
		app, err := dagger.PackBuild(filepath.Join("testdata", "source_2.1_float_runtime"), runtimeURI, aspnetURI, sdkURI, buildURI)
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet netcoreapp2.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("netcoreapp2"))
	})

	// Todo - Copied from V2 to be fixed for V3
	it("should build a working OCI image for a source 2.1 with templated global json application", func() {
		// app, err := dagger.PackBuild(filepath.Join("testdata", "source_2.1_global_json_templated"), runtimeURI, aspnetURI, sdkURI, buildURI)
		// Expect(err).ToNot(HaveOccurred())

		// Expect(app.StartWithCommand("dotnet simple_brats.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		// Eventually(func() string {
		// 	body, _, _ := app.HTTPGet("/")
		// 	return body
		// }).Should(ContainSubstring("simple_brats"))
	})

	// Todo - Copied from V2 to be fixed for V3
	it("should build a working OCI image for a source_3_0_app application", func() {
		// app, err := dagger.PackBuild(filepath.Join("testdata", "source_3_0_app"), runtimeURI, aspnetURI, sdkURI, buildURI)
		// Expect(err).ToNot(HaveOccurred())

		// Expect(app.StartWithCommand("dotnet source_3_0_app.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		// Eventually(func() string {
		// 	body, _, _ := app.HTTPGet("/")
		// 	return body
		// }).Should(ContainSubstring("source_3_0_app"))
	})

	// Test fixture moved from V2 and succeeds
	it("should build a working OCI image for a source_aspnetcore_all_2.1 application", func() {
		app, err := dagger.PackBuild(filepath.Join("testdata", "source_aspnetcore_all_2.1"), runtimeURI, aspnetURI, sdkURI, buildURI)
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("dotnet netcoreapp2.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("netcoreapp2"))
	})

	// Todo - Copied from V2 to be fixed for V3
	it("should build a working OCI image for a source_aspnetcore_app_2.1 application", func() {
		// app, err := dagger.PackBuild(filepath.Join("testdata", "source_aspnetcore_app_2.1"), runtimeURI, aspnetURI, sdkURI, buildURI)
		// Expect(err).ToNot(HaveOccurred())

		// Expect(app.StartWithCommand("dotnet source_aspnetcore_app_2.1.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		// Eventually(func() string {
		// 	body, _, _ := app.HTTPGet("/")
		// 	return body
		// }).Should(ContainSubstring("source_aspnetcore_app_2.1"))
	})

	// Todo - Copied from V2 to be fixed for V3
	it("should build a working OCI image for a source_prerender_node application", func() {
		// app, err := dagger.PackBuild(filepath.Join("testdata", "source_prerender_node"), runtimeURI, aspnetURI, sdkURI, buildURI)
		// Expect(err).ToNot(HaveOccurred())

		// Expect(app.StartWithCommand("dotnet asp_prerender_node.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		// Eventually(func() string {
		// 	body, _, _ := app.HTTPGet("/")
		// 	return body
		// }).Should(ContainSubstring("asp_prerender_node"))
	})

	// Todo - Copied from V2 to be fixed for V3
	it("should build a working OCI image for a uses_libgdiplus application", func() {
		// app, err := dagger.PackBuild(filepath.Join("testdata", "uses_libgdiplus"), runtimeURI, aspnetURI, sdkURI, buildURI)
		// Expect(err).ToNot(HaveOccurred())

		// Expect(app.StartWithCommand("dotnet uses_libgdiplus.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		// Eventually(func() string {
		// 	body, _, _ := app.HTTPGet("/")
		// 	return body
		// }).Should(ContainSubstring("uses_libgdiplus"))
	})

	// Todo - Copied from V2 to be fixed for V3
	it("should build a working OCI image for a with_buildpack_yml_templated application", func() {
		// app, err := dagger.PackBuild(filepath.Join("testdata", "with_buildpack_yml_templated"), runtimeURI, aspnetURI, sdkURI, buildURI)
		// Expect(err).ToNot(HaveOccurred())

		// Expect(app.StartWithCommand("dotnet simple_brats.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		// Eventually(func() string {
		// 	body, _, _ := app.HTTPGet("/")
		// 	return body
		// }).Should(ContainSubstring("simple_brats"))
	})

	// Todo - Copied from V2 to be fixed for V3
	it("should build a working OCI image for a with_dot_in_name application", func() {
		// app, err := dagger.PackBuild(filepath.Join("testdata", "with_dot_in_name"), runtimeURI, aspnetURI, sdkURI, buildURI)
		// Expect(err).ToNot(HaveOccurred())

		// Expect(app.StartWithCommand("dotnet with_dot_in_name.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		// Eventually(func() string {
		// 	body, _, _ := app.HTTPGet("/")
		// 	return body
		// }).Should(ContainSubstring("with_dot_in_name"))
	})

}
