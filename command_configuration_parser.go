package dotnetpublish

import (
	"os"
	"strings"

	"github.com/mattn/go-shellwords"
)

type CommandConfigurationParser struct {
}

func NewCommandConfigurationParser() CommandConfigurationParser {
	return CommandConfigurationParser{}
}

func (p CommandConfigurationParser) ParseFlagsFromEnvVar(envVarName string) (flags []string, err error) {
	shellwordsParser := shellwords.NewParser()
	shellwordsParser.ParseEnv = true

	if rawFlags, ok := os.LookupEnv(envVarName); ok {
		var err error
		flags, err = shellwordsParser.Parse(rawFlags)
		if err != nil {
			return nil, err
		}
	}
	return flags, nil
}

func containsFlag(flags []string, match string) bool {
	for _, flag := range flags {
		if strings.HasPrefix(flag, match) {
			return true
		}
	}
	return false
}
