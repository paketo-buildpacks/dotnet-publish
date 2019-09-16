package publish_test

import (
	"github.com/cloudfoundry/dotnet-core-build-cnb/publish"
	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

//go:generate mockgen -source=publish.go -destination=mocks_test.go -package=publish_test

func TestUnitPublish(t *testing.T) {
	spec.Run(t, "Detect", testPublish, spec.Report(report.Terminal{}))
}

func testPublish(t *testing.T, when spec.G, it spec.S) {
	var (
		factory    *test.BuildFactory
		buildLayer layers.Layer
		mockRunner *MockRunner
		mockCtrl   *gomock.Controller
	)

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewBuildFactory(t)
		buildLayer = factory.Build.Layers.Layer("build")
		mockCtrl = gomock.NewController(t)
		mockRunner = NewMockRunner(mockCtrl)
	})

	it.After(func() {
		mockCtrl.Finish()
	})

	when("publish.NewContributor", func() {
		it("return true if a build plan exists", func() {
			factory.AddPlan(buildpackplan.Plan{Name: publish.Publish})

			publishContributor, willContribute, err := publish.NewContributor(factory.Build, mockRunner)
			Expect(err).ToNot(HaveOccurred())
			Expect(willContribute).To(BeTrue())
			Expect(publishContributor).ToNot(Equal(publish.Contributor{}))
		})

		it("return false if a build plan does not exists", func() {
			publishContributor, willContribute, err := publish.NewContributor(factory.Build, mockRunner)
			Expect(err).ToNot(HaveOccurred())
			Expect(willContribute).To(BeFalse())
			Expect(publishContributor).To(Equal(publish.Contributor{}))
		})
	})

	when("Contribute", func() {
		var (
			dotnetRoot string
			sdkLayer   string
		)
		it.Before(func() {
			var err error

			dotnetRoot, err = ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())
			sdkLayer, err = ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())

			symlinkTarget := filepath.Join(dotnetRoot, "symlink-target")
			Expect(os.MkdirAll(symlinkTarget, os.ModePerm)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(dotnetRoot, "shared"), os.ModePerm)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(dotnetRoot, "host"), os.ModePerm)).To(Succeed())
			Expect(os.Symlink(symlinkTarget, filepath.Join(dotnetRoot, "shared", "dir1"))).To(Succeed())
			Expect(os.Symlink(symlinkTarget, filepath.Join(dotnetRoot, "shared", "dir2"))).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(dotnetRoot, "dotnet"), []byte("dotnet executable"), 0644))

			Expect(os.MkdirAll(filepath.Join(sdkLayer, "sdk"), os.ModePerm)).To(Succeed())

			Expect(os.Setenv("DOTNET_ROOT", dotnetRoot)).To(Succeed())
			Expect(os.Setenv("SDK_LOCATION", sdkLayer)).To(Succeed())
		})

		it.After(func() {
			_ = os.RemoveAll(dotnetRoot)
			_ = os.RemoveAll(sdkLayer)
			_ = os.Unsetenv("DOTNET_ROOT")
			_ = os.Unsetenv("SDK_LOCATION")
		})

		it("symlinks shared frameworks and sdk, copies the dotnet driver", func() {
			factory.AddPlan(buildpackplan.Plan{Name: publish.Publish})

			mockRunner.EXPECT().Run(
				"dotnet",
				factory.Build.Application.Root,
				false,
				"publish",
				"-c", "Release",
				"-r", "ubuntu.18.04-x64",
				"--self-contained", "false",
				"-o", factory.Build.Application.Root,
			).Return(nil)

			publishContributor, _, err := publish.NewContributor(factory.Build, mockRunner)
			Expect(err).ToNot(HaveOccurred())

			Expect(publishContributor.Contribute()).To(Succeed())

			ExpectSymlink(filepath.Join(buildLayer.Root, "shared", "dir1"), t)
			ExpectSymlink(filepath.Join(buildLayer.Root, "shared", "dir2"), t)

			ExpectSymlink(filepath.Join(buildLayer.Root, "host"), t)

			Expect(filepath.Join(buildLayer.Root, "dotnet")).To(BeARegularFile())

			ExpectSymlink(filepath.Join(buildLayer.Root, "sdk"), t)

			Expect(os.Getenv("PATH")).To(HavePrefix(buildLayer.Root))
		})
	})
}

func ExpectSymlink(path string, t *testing.T) {
	t.Helper()
	hostFileInfo, err := os.Stat(path)
	Expect(err).ToNot(HaveOccurred())
	Expect(hostFileInfo.Mode() & os.ModeSymlink).ToNot(Equal(0))
}
