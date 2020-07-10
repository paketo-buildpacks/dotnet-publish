package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/semver"
	"github.com/cloudfoundry/dagger"
	"github.com/sclevine/agouti"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	bpDir, buildURI, nodeURI, runtimeURI, builder string
	bpList, bpNoASPList                           []string
)

const (
	testBuildpack    = "test-buildpack"
	aspnetBuildpack  = "dotnet-core-aspnet-cnb"
	runtimeBuildpack = "dotnet-core-runtime-cnb"
)

func Package(root, version string, cached bool) (string, error) {
	var cmd *exec.Cmd

	bpPath := filepath.Join(root, "artifact")
	if cached {
		cmd = exec.Command(".bin/packager", "--archive", "--version", version, fmt.Sprintf("%s-cached", bpPath))
	} else {
		cmd = exec.Command(".bin/packager", "--archive", "--uncached", "--version", version, bpPath)
	}

	cmd.Env = append(os.Environ(), fmt.Sprintf("PACKAGE_DIR=%s", bpPath))
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if cached {
		return fmt.Sprintf("%s-cached.tgz", bpPath), err
	}

	return fmt.Sprintf("%s.tgz", bpPath), err
}

func BeforeSuite() {
	var err error
	var config dagger.TestConfig

	bpDir, err = filepath.Abs("./..")
	Expect(err).NotTo(HaveOccurred())

	buildURI, err = Package(bpDir, "1.2.3", false)
	Expect(err).ToNot(HaveOccurred())

	config, err = dagger.ParseConfig("config.json")
	Expect(err).ToNot(HaveOccurred())

	builder = config.Builder

	for _, bp := range config.BuildpackOrder[builder] {
		var bpURI string
		if bp == testBuildpack {
			bpURI = buildURI
		} else {
			bpURI, err = dagger.GetLatestBuildpack(bp)
			Expect(err).ToNot(HaveOccurred())
		}

		if bp == runtimeBuildpack {
			runtimeURI = bpURI
		}

		bpList = append(bpList, bpURI)
		if bp != aspnetBuildpack {
			bpNoASPList = append(bpNoASPList, bpURI)
		}
	}

	nodeURI, err = dagger.GetLatestBuildpack("node-engine-cnb")
	Expect(err).ToNot(HaveOccurred())
}

func AfterSuite() {
	Expect(dagger.DeleteBuildpack(buildURI)).To(Succeed())
	Expect(dagger.DeleteBuildpack(nodeURI)).To(Succeed())
	for _, bp := range bpList {
		Expect(dagger.DeleteBuildpack(bp)).To(Succeed())
	}
}

func TestIntegration(t *testing.T) {
	var suite = spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Integration", testIntegration)

	RegisterTestingT(t)
	BeforeSuite()
	suite.Run(t)
	AfterSuite()
}

func testIntegration(t *testing.T, _ spec.G, it spec.S) {
	var (
		app *dagger.App
		err error
	)

	it.After(func() {
		Expect(app.Destroy()).To(Succeed())
	})

	it("should build a working OCI image for a simple 2.1 app with aspnet dependencies", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "source-2.1-aspnet"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(bpList...),
			dagger.SetBuilder(builder),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		if builder == "bionic" {
			app.SetHealthCheck("stat /workspace", "2s", "15s")
		}

		Expect(app.StartWithCommand("dotnet source-2.1-aspnet.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("source_2._1_aspnet"))
	})

	it("should build a working OCI image for a simple 2.1 webapi with swagger dependency", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "source-2.1-aspnet-with-public-nuget"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(bpList...),
			dagger.SetBuilder(builder),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		if builder == "bionic" {
			app.SetHealthCheck("stat /workspace", "2s", "15s")
		}

		Expect(app.StartWithCommand("dotnet source-2.1-aspnet-with-public-nuget.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/swagger/index.html")
			return body
		}).Should(ContainSubstring("SourceWithNuget"))
	})

	it("should build a working OCI image for a angular dotnet 2.1 application", func() {
		//This test pulls in a node module that relies on python, which is not present in bionic
		if builder != "bionic" {
			nodeOrder := append([]string{nodeURI}, bpList...)
			app, err = dagger.NewPack(
				filepath.Join("testdata", "angular_msbuild_dotnet_2.1"),
				dagger.RandomImage(),
				dagger.SetBuildpacks(nodeOrder...),
				dagger.SetBuilder(builder),
				dagger.SetVerbose(),
			).Build()
			Expect(err).ToNot(HaveOccurred())

			Expect(app.StartWithCommand("dotnet angular_msbuild.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

			url := app.GetBaseURL()

			browser := agouti.ChromeDriver(agouti.ChromeOptions("args", []string{"--headless", "--disable-gpu", "--no-sandbox"}))
			err = browser.Start()
			Expect(err).NotTo(HaveOccurred())

			page, err := browser.NewPage()
			Expect(err).NotTo(HaveOccurred())
			Expect(page.Navigate(url)).To(Succeed())
			Eventually(page.HTML, 30*time.Second).Should(ContainSubstring("Hello, world!"))
		}
	})

	it("should build a working OCI image for an app that specifies it should be self contained", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "self_contained_msbuild"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(bpNoASPList...),
			dagger.SetBuilder(builder),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		if builder == "bionic" {
			app.SetHealthCheck("stat /workspace", "2s", "15s")
		}

		Expect(app.Start()).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("Hello World!"))
	})

	it("should build a working OCI image for a console app", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "console_app"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(bpNoASPList...),
			dagger.SetBuilder(builder),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		app.SetHealthCheck("stat /workspace/console.dll", "2s", "15s")
		Expect(app.Start()).To(Succeed())

		Eventually(func() string {
			body, _ := app.Logs()
			return body
		}).Should(ContainSubstring("Hello World!"))
	})

	it("should build a working OCI image for a fsharp application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "fsharp_msbuild"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(bpList...),
			dagger.SetBuilder(builder),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		if builder == "bionic" {
			app.SetHealthCheck("stat /workspace", "2s", "15s")
		}

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
	it("should build a working OCI image for an with multiple project files", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "multiple_projects_msbuild"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(bpNoASPList...),
			dagger.SetBuilder(builder),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		if builder == "bionic" {
			app.SetHealthCheck("stat /workspace", "2s", "15s")
		}

		Expect(app.StartWithCommand("dotnet asp_web_app.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("Hello, I'm a string!"))
	})

	it("should build a working OCI image for a nancy kestrel msbuild application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "nancy_kestrel_msbuild_dotnet2"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(bpNoASPList...),
			dagger.SetBuilder(builder),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		if builder == "bionic" {
			app.SetHealthCheck("stat /workspace", "2s", "15s")
		}

		app.Env["PORT"] = "8080"
		app.Env["ASPNETCORE_URLS"] = "http://0.0.0.0:8080"
		Expect(app.StartWithCommand("dotnet nancy_kestrel_msbuild_dotnet2.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("Hello from Nancy running on CoreCLR"))
	})

	it("should build a working OCI image for a source_2.1_explicit_runtime_templated application", func() {
		majorMinor := "2.1"
		version, err := GetLowestRuntimeVersionInMajorMinor(majorMinor, filepath.Join(runtimeURI, "buildpack.toml"))
		Expect(err).ToNot(HaveOccurred())
		bpYml := fmt.Sprintf(`<Project Sdk="Microsoft.NET.Sdk.Web">

  <PropertyGroup>
    <TargetFramework>netcoreapp%s</TargetFramework>
    <RuntimeFrameworkVersion>%s</RuntimeFrameworkVersion>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="Microsoft.AspNetCore.All" Version="2.1.0" />
  </ItemGroup>

  <ItemGroup>
    <DotNetCliToolReference Include="Microsoft.VisualStudio.Web.CodeGeneration.Tools" Version="2.0.0" />
  </ItemGroup>

</Project>
`, majorMinor, version)

		bpYmlPath := filepath.Join("testdata", "source_2.1_explicit_runtime_templated", "netcoreapp2.csproj")
		Expect(ioutil.WriteFile(bpYmlPath, []byte(bpYml), 0644)).To(Succeed())

		app, err = dagger.NewPack(
			filepath.Join("testdata", "source_2.1_explicit_runtime_templated"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(bpList...),
			dagger.SetBuilder(builder),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		if builder == "bionic" {
			app.SetHealthCheck("stat /workspace", "2s", "15s")
		}

		Expect(app.StartWithCommand("dotnet netcoreapp2.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("netcoreapp2"))
	})

	it("should build a working OCI image for a source_2.1_float_runtime application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "source_2.1_float_runtime"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(bpList...),
			dagger.SetBuilder(builder),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		if builder == "bionic" {
			app.SetHealthCheck("stat /workspace", "2s", "15s")
		}

		Expect(app.StartWithCommand("dotnet netcoreapp2.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())
		Expect(app.BuildLogs()).To(ContainSubstring("dotnet-runtime_2.1"))
		Expect(app.BuildLogs()).To(ContainSubstring("dotnet-aspnetcore_2.1"))
		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("netcoreapp2"))
	})

	it("should build a working OCI image for a source_3_1_app application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "source_3_1_app"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(bpList...),
			dagger.SetBuilder(builder),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		if builder == "bionic" {
			app.SetHealthCheck("stat /workspace", "2s", "15s")
		}

		Expect(app.StartWithCommand("dotnet simple_3_0_app.dll --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("simple_3_0_app"))
	})

	it("should build a working OCI image for a source_aspnetcore_all_2.1 application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "source_aspnetcore_all_2.1"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(bpList...),
			dagger.SetBuilder(builder),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		if builder == "bionic" {
			app.SetHealthCheck("stat /workspace", "2s", "15s")
		}

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
			dagger.SetBuildpacks(bpList...),
			dagger.SetBuilder(builder),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		if builder == "bionic" {
			app.SetHealthCheck("stat /workspace", "2s", "15s")
		}

		err = app.StartWithCommand("dotnet source_aspnetcore_2.1.dll --urls http://0.0.0.0:${PORT}")
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("Hello World!"))
	})

	it("should build a working OCI image for a source_steeltoe_3.1 application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "source_steeltoe_3.1"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(bpList...),
			dagger.SetBuilder(builder),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		app.SetHealthCheck("stat /workspace", "2s", "15s")

		// watch out this PORT is hard-coded to 8080 in the application
		err = app.StartWithCommand("./Steeltoe.Demo --urls http://0.0.0.0:8080")
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/api/values/6")
			return body
		}).Should(ContainSubstring("value"))
	})

	// TODO: Requires node should be moved
	it.Pend("should build a working OCI image for a source_prerender_node application", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "source_prerender_node"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(bpList...),
			dagger.SetBuilder(builder),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		if builder == "bionic" {
			app.SetHealthCheck("stat /workspace", "2s", "15s")
		}

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
			dagger.SetBuildpacks(bpList...),
			dagger.SetBuilder(builder),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		if builder == "bionic" {
			app.SetHealthCheck("stat /workspace", "2s", "15s")
		}

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
			dagger.SetBuildpacks(bpList...),
			dagger.SetBuilder(builder),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		if builder == "bionic" {
			app.SetHealthCheck("stat /workspace", "2s", "15s")
		}

		Expect(app.StartWithCommand("dotnet with_dot_in_name.dll --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("with_dot_in_name"))
	})
}

func GetLowestRuntimeVersionInMajorMinor(majorMinor, bpTomlPath string) (string, error) {
	type buildpackTomlVersion struct {
		Metadata struct {
			Dependencies []struct {
				Version string `toml:"version"`
			} `toml:"dependencies"`
		} `toml:"metadata"`
	}

	bpToml := buildpackTomlVersion{}
	output, err := ioutil.ReadFile(filepath.Join(bpTomlPath))
	if err != nil {
		return "", err
	}

	majorMinorConstraint, err := semver.NewConstraint(fmt.Sprintf("%s.*", majorMinor))
	if err != nil {
		return "", err
	}

	lowestVersion, err := semver.NewVersion(fmt.Sprintf("%s.9999", majorMinor))
	if err != nil {
		return "", err
	}

	_, err = toml.Decode(string(output), &bpToml)
	if err != nil {
		return "", err
	}

	for _, dep := range bpToml.Metadata.Dependencies {
		depVersion, err := semver.NewVersion(dep.Version)
		if err != nil {
			return "", err
		}
		if majorMinorConstraint.Check(depVersion) {
			if depVersion.LessThan(lowestVersion) {
				lowestVersion = depVersion
			}
		}
	}

	return lowestVersion.String(), nil
}
