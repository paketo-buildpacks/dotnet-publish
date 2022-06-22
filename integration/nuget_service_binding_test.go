package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"net/http"
	"net/http/httptest"
	"net/url"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/occam"
	. "github.com/paketo-buildpacks/occam/matchers"
	"github.com/sclevine/spec"
)

func testNugetConfig(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		pack   occam.Pack
	)

	it.Before(func() {
		pack = occam.NewPack()
	})

	context("when building a .NET Core app", func() {
		var (
			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		context("when a nuget.config is provided via service binding", func() {
			var (
				server    *httptest.Server
				binding   string
				packBuild occam.PackBuild
				serverURI string
			)
			it.Before(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					switch req.URL.Path {
					case "/FindPackagesById()":
						w.WriteHeader(http.StatusOK)
					default:
						t.Fatalf("unknown path: %s", req.URL.Path)
					}
				}))

				uri, err := url.Parse(server.URL)
				Expect(err).NotTo(HaveOccurred())

				switch os := runtime.GOOS; os {
				case "darwin":
					serverURI = fmt.Sprintf("http://host.docker.internal:%s", uri.Port())
				case "windows":
					serverURI = fmt.Sprintf("http://host.docker.internal:%s", uri.Port())
				case "linux":
					// host.docker.internal is not supported on linux; use host's address
					// and build WithNetwork("host")
					serverURI = fmt.Sprintf("http://127.0.0.1:%s", uri.Port())
				default:
					t.Fatal("unrecognized runtime.GOOS: " + runtime.GOOS)
				}
				binding, err = os.MkdirTemp("", "bindingdir")
				Expect(err).NotTo(HaveOccurred())
				Expect(os.Chmod(binding, os.ModePerm)).To(Succeed())

				Expect(os.WriteFile(filepath.Join(binding, "type"), []byte("nugetconfig"), os.ModePerm)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(binding, "nuget.config"), []byte(fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <clear />
    <add key="nuget" value="%s" />
  </packageSources>
</configuration>`, serverURI)), os.ModePerm)).To(Succeed())

				packBuild = pack.Build.
					WithBuildpacks(
						icuBuildpack,
						dotnetCoreRuntimeBuildpack,
						dotnetCoreAspNetBuildpack,
						dotnetCoreSDKBuildpack,
						buildpack,
						dotnetExecuteBuildpack,
					).
					WithEnv(map[string]string{
						"SERVICE_BINDING_ROOT":    "/bindings",
						"BP_DOTNET_PUBLISH_FLAGS": "--verbosity=normal",
					}).
					WithVolumes(fmt.Sprintf("%s:/bindings/nugetconfig", binding))

				if runtime.GOOS == "linux" {
					packBuild = packBuild.WithNetwork("host") // this allows the container to reach 127.0.0.1 on the host
				}
			})

			it.After(func() {
				server.Close()
				os.RemoveAll(binding)
			})

			it("fails due to invalid package, but the nuget.config package source is used", func() {
				var err error
				source, err := occam.Source(filepath.Join("testdata", "source-3.1-aspnet-with-public-nuget"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				_, logs, err = packBuild.Execute(name, source)

				Expect(logs).To(ContainLines(
					"  Loading nuget service binding",
					"  Executing build process",
				))

				// Logs indicate the right registry was used, but the build failed because proper package content wasn't returned.
				Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("GET %s/FindPackagesById()?id='Swashbuckle.AspNetCore'&semVerLevel=2.0.0", serverURI))))
				Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("OK %s/FindPackagesById()?id='Swashbuckle.AspNetCore'&semVerLevel=2.0.0", serverURI))))
				Expect(err).To(MatchError(ContainSubstring("error : Failed to retrieve information about 'Swashbuckle.AspNetCore' from remote source")))
			})
		})
	})
}
