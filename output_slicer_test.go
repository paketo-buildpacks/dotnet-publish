package dotnetpublish_test

import (
	"os"
	"path/filepath"
	"testing"

	dotnetpublish "github.com/paketo-buildpacks/dotnet-publish"
	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testOutputSlicer(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		slicer    dotnetpublish.OutputSlicer
		assetsDir string
	)

	it.Before(func() {
		var err error
		assetsDir, err = occam.Source("testdata")
		Expect(err).NotTo(HaveOccurred())

		slicer = dotnetpublish.NewOutputSlicer()
	})

	it.After(func() {
		Expect(os.RemoveAll(assetsDir)).To(Succeed())
	})

	it("extracts the base file name of packages' runtime and runtimeTargets", func() {
		pkgs, _, _, err := slicer.Slice(filepath.Join(assetsDir, "packages.project.assets.json"))
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs.Paths).To(HaveLen(5))
		// Must use ContainElements, not Equal(), because unpacking JSON map into array
		// produces non-deterministic ordering
		Expect(pkgs.Paths).To(ContainElements([]string{
			"Microsoft.OpenApi.dll",
			"Swashbuckle.AspNetCore.SwaggerGen.dll",
			"Grpc.Core.dll",
			"libgrpc_csharp_ext.arm64.so",
			"grpc_csharp_ext.x86.dll",
		}))
		// Ignore the blanked out file name for the CSharp dependency
		Expect(pkgs.Paths).NotTo(ContainElement("_._"))
	})

	it("extracts the base file name of projects' runtime dlls", func() {
		pkgs, earlyPkgs, projects, err := slicer.Slice(filepath.Join(assetsDir, "projects.project.assets.json"))
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs.Paths).To(HaveLen(2))
		Expect(earlyPkgs.Paths).To(HaveLen(0))
		Expect(projects.Paths).To(HaveLen(2))
		Expect(projects.Paths).To(ContainElements([]string{
			"Migrations.dll",
			"Squidex.Domain.Apps.Core.Model.dll",
		}))
	})

	it("distinguishes between packages and early packages", func() {
		pkgs, earlyPkgs, projects, err := slicer.Slice(filepath.Join(assetsDir, "packages.project.assets.json"))
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs.Paths).To(HaveLen(5))
		Expect(earlyPkgs.Paths).To(HaveLen(1))
		Expect(projects.Paths).To(HaveLen(0))
	})

	context("failure cases", func() {
		context("assets file cannot be opened", func() {
			it.Before(func() {
				Expect(os.Chmod(filepath.Join(assetsDir, "packages.project.assets.json"), 0000)).To(Succeed())
			})
			it.After(func() {
				Expect(os.Chmod(filepath.Join(assetsDir, "packages.project.assets.json"), os.ModePerm)).To(Succeed())
			})
			it("returns an error", func() {
				_, _, _, err := slicer.Slice(filepath.Join(assetsDir, "packages.project.assets.json"))
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
		context("assets file JSON cannot be decoded", func() {
			// TODO Discuss: Should the slicer actually fail here? Or is slicing optional behaviour?
			it("returns an error", func() {
				_, _, _, err := slicer.Slice(filepath.Join(assetsDir, "malformed.project.assets.json"))
				Expect(err).To(MatchError(ContainSubstring("invalid character 's' looking for beginning of value")))
			})
		})
	})
}
