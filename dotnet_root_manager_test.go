package dotnetpublish_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	dotnetpublish "github.com/paketo-buildpacks/dotnet-publish"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDotnetRootManager(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		existingRootDir string
		sdkLayerDir     string
		rootDir         string
		manager         dotnetpublish.DotnetRootManager
	)

	it.Before(func() {
		var err error
		existingRootDir, err = ioutil.TempDir("", "existing-root-dir")
		Expect(err).NotTo(HaveOccurred())

		sdkLayerDir, err = ioutil.TempDir("", "existing-root-dir")
		Expect(err).NotTo(HaveOccurred())

		rootDir, err = ioutil.TempDir("", "root-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.MkdirAll(filepath.Join(existingRootDir, "host"), os.ModePerm)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(existingRootDir, "shared", "some-dir"), os.ModePerm)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(existingRootDir, "shared", "some-file"), nil, 0600)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(existingRootDir, "dotnet"), nil, 0700)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(sdkLayerDir, "sdk"), os.ModePerm)).To(Succeed())

		manager = dotnetpublish.NewDotnetRootManager()
	})

	it.After(func() {
		Expect(os.RemoveAll(existingRootDir)).To(Succeed())
		Expect(os.RemoveAll(sdkLayerDir)).To(Succeed())
	})

	context("Setup", func() {
		it("sets up the DOTNET_ROOT directory", func() {
			err := manager.Setup(rootDir, existingRootDir, sdkLayerDir)
			Expect(err).NotTo(HaveOccurred())

			var files []string
			err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				files = append(files, path)

				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(files).To(ConsistOf([]string{
				filepath.Join(rootDir),
				filepath.Join(rootDir, "host"),
				filepath.Join(rootDir, "dotnet"),
				filepath.Join(rootDir, "shared"),
				filepath.Join(rootDir, "shared", "some-dir"),
				filepath.Join(rootDir, "shared", "some-file"),
				filepath.Join(rootDir, "sdk"),
			}))
		})
	})

	context("failure cases", func() {
		context("when the paths can't be found in the existing root", func() {
			var badPatternRootDir string
			it.Before(func() {
				badPatternRootDir = "\\"
			})

			it("errrors", func() {
				err := manager.Setup(rootDir, badPatternRootDir, sdkLayerDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not glob"))
			})
		})

		context("when root dir can not be written to", func() {
			it.Before(func() {
				Expect(os.Chmod(rootDir, 0000)).To(Succeed())
			})

			it("errors", func() {
				err := manager.Setup(rootDir, existingRootDir, sdkLayerDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not write to "))
			})
		})

		context("when a symlink cannot be created in the shared directory", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(rootDir, "shared"), 0000)).To(Succeed())
			})

			it("errors", func() {
				err := manager.Setup(rootDir, existingRootDir, sdkLayerDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not create symlink:"))
			})
		})

		context("when a symlink can not be created in the host directory", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(rootDir, "host"), 0000)).To(Succeed())
			})

			it("errors", func() {
				err := manager.Setup(rootDir, existingRootDir, sdkLayerDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not create symlink:"))
			})
		})

		context("when the dotnet cli can not be copied", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(rootDir, "dotnet"), os.ModePerm)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(rootDir, "dotnet", "some-file"), nil, 0000)).To(Succeed())
			})

			it("errors", func() {
				err := manager.Setup(rootDir, existingRootDir, sdkLayerDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not copy the dotnet cli:"))
			})
		})

		context("when a symlink can not be created in the sdk directory", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(rootDir, "sdk"), 0000)).To(Succeed())
			})

			it("errors", func() {
				err := manager.Setup(rootDir, existingRootDir, sdkLayerDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not create symlink:"))
			})
		})
	})
}
