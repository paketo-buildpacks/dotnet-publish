package dotnetpublish_test

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
	dotnetpublish "github.com/paketo-buildpacks/dotnet-publish"
	"github.com/sclevine/spec"
)

func testCommandConfigurationParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		parser dotnetpublish.CommandConfigurationParser
	)

	it.Before(func() {
		parser = dotnetpublish.NewCommandConfigurationParser()
	})

	context("when passed an environment variable whose value is a set of flags", func() {
		it.Before(func() {
			os.Setenv("SOME_FLAGS_ENV_VAR", "--no-restore --packages /some/path -g --self-contained=true -l $OTHER_ENV_VAR")
			os.Setenv("OTHER_ENV_VAR", "other-value")
		})

		it.After(func() {
			os.Unsetenv("SOME_FLAGS_ENV_VAR")
			os.Unsetenv("OTHER_ENV_VAR")
		})

		it("parses the flags from the environment variable", func() {
			flags, err := parser.ParseFlagsFromEnvVar("SOME_FLAGS_ENV_VAR")
			Expect(err).NotTo(HaveOccurred())

			Expect(flags).To(Equal([]string{"--no-restore",
				"--packages",
				"/some/path",
				"-g",
				"--self-contained=true",
				"-l",
				"other-value",
			}))
		})
	})

	context("failure cases", func() {
		context("when env var isn't a well-formed set of flags", func() {
			it.Before(func() {
				os.Setenv("SOME_FLAGS_ENV_VAR", "\"")
			})

			it.After(func() {
				os.Unsetenv("SOME_FLAGS_ENV_VAR")
			})
			it("returns an error", func() {
				_, err := parser.ParseFlagsFromEnvVar("SOME_FLAGS_ENV_VAR")
				Expect(err).To(MatchError(ContainSubstring("invalid command line string")))
			})
		})
	})
}
