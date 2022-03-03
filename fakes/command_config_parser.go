package fakes

import "sync"

type CommandConfigParser struct {
	ParseFlagsFromEnvVarCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			EnvVar string
		}
		Returns struct {
			StringSlice []string
			Error       error
		}
		Stub func(string) ([]string, error)
	}
}

func (f *CommandConfigParser) ParseFlagsFromEnvVar(param1 string) ([]string, error) {
	f.ParseFlagsFromEnvVarCall.mutex.Lock()
	defer f.ParseFlagsFromEnvVarCall.mutex.Unlock()
	f.ParseFlagsFromEnvVarCall.CallCount++
	f.ParseFlagsFromEnvVarCall.Receives.EnvVar = param1
	if f.ParseFlagsFromEnvVarCall.Stub != nil {
		return f.ParseFlagsFromEnvVarCall.Stub(param1)
	}
	return f.ParseFlagsFromEnvVarCall.Returns.StringSlice, f.ParseFlagsFromEnvVarCall.Returns.Error
}
