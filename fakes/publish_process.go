package fakes

import "sync"

type PublishProcess struct {
	ExecuteCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			WorkingDir     string
			RootDir        string
			NugetCachePath string
			ProjectPath    string
			OutputPath     string
			Flags          []string
		}
		Returns struct {
			Error error
		}
		Stub func(string, string, string, string, string, []string) error
	}
}

func (f *PublishProcess) Execute(param1 string, param2 string, param3 string, param4 string, param5 string, param6 []string) error {
	f.ExecuteCall.mutex.Lock()
	defer f.ExecuteCall.mutex.Unlock()
	f.ExecuteCall.CallCount++
	f.ExecuteCall.Receives.WorkingDir = param1
	f.ExecuteCall.Receives.RootDir = param2
	f.ExecuteCall.Receives.NugetCachePath = param3
	f.ExecuteCall.Receives.ProjectPath = param4
	f.ExecuteCall.Receives.OutputPath = param5
	f.ExecuteCall.Receives.Flags = param6
	if f.ExecuteCall.Stub != nil {
		return f.ExecuteCall.Stub(param1, param2, param3, param4, param5, param6)
	}
	return f.ExecuteCall.Returns.Error
}
