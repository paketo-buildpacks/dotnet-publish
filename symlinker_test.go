package dotnetpublish_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	dotnetpublish "github.com/paketo-buildpacks/dotnet-publish"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSymlinker(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
		symlinker  dotnetpublish.Symlinker
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		symlinker = dotnetpublish.NewSymlinker()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("Link", func() {
		it.Before(func() {
			Expect(os.WriteFile(filepath.Join(workingDir, "old"), []byte("old file contents"), os.ModePerm)).To(Succeed())
		})
		it("symlinks the given old location to the new location", func() {
			Expect(symlinker.Link(filepath.Join(workingDir, "old"), filepath.Join(workingDir, "new"))).To(Succeed())

			link, err := os.Readlink(filepath.Join(workingDir, "new"))
			Expect(err).NotTo(HaveOccurred())
			Expect(link).To(Equal(filepath.Join(workingDir, "old")))
		})

		context("when newname includes directories that don't yet exist", func() {
			it("creates the directories and then makes the symlink", func() {
				Expect(symlinker.Link(filepath.Join(workingDir, "old"), filepath.Join(workingDir, "subdir", "new"))).To(Succeed())

				link, err := os.Readlink(filepath.Join(workingDir, "subdir", "new"))
				Expect(err).NotTo(HaveOccurred())
				Expect(link).To(Equal(filepath.Join(workingDir, "old")))
			})
		})

		context("failure cases", func() {
			context("directory cannot be created for symlink", func() {
				it.Before(func() {
					Expect(os.Mkdir(filepath.Join(workingDir, "subdir"), 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.RemoveAll(filepath.Join(workingDir, "subdir"))).To(Succeed())
				})

				it("returns an error", func() {
					err := symlinker.Link(filepath.Join(workingDir, "old"), filepath.Join(workingDir, "subdir", "anotherdir", "new"))
					Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("failed to make directory for symlink: mkdir %s", filepath.Join(workingDir, "subdir", "anotherdir"))))
				})
			})

			context("symlink cannot be created", func() {
				it.Before(func() {
					Expect(os.WriteFile(filepath.Join(workingDir, "preexisting"), []byte("preexisting file contents"), os.ModePerm)).To(Succeed())
				})
				it("returns an error", func() {
					err := symlinker.Link(filepath.Join(workingDir, "old"), filepath.Join(workingDir, "preexisting"))
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	context("Unlink", func() {
		it.Before(func() {
			Expect(os.WriteFile(filepath.Join(workingDir, "old"), []byte("old file contents"), os.ModePerm)).To(Succeed())
			Expect(os.Symlink(filepath.Join(workingDir, "old"), filepath.Join(workingDir, "new"))).To(Succeed())
		})
		it("unlinks the given old location from the new location", func() {
			Expect(symlinker.Unlink(filepath.Join(workingDir, "new"))).To(Succeed())

			_, err := os.Readlink(filepath.Join(workingDir, "new"))
			Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue())
		})

		context("the given file doesn't exist", func() {
			it("does nothing", func() {
				err := symlinker.Unlink(filepath.Join(workingDir, "other"))
				Expect(err).NotTo(HaveOccurred())
				_, err = os.Readlink(filepath.Join(workingDir, "other"))
				Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue())
			})
		})

		context("failure cases", func() {
			context("the given file doesn't have a symlink", func() {
				it("returns an error", func() {
					err := symlinker.Unlink(filepath.Join(workingDir, "old"))
					Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("cannot unlink %s because it is not a symlink", filepath.Join(workingDir, "old")))))
				})
			})
		})
	})

}
