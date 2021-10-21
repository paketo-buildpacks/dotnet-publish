package fakes

import "sync"

type PublishProcess struct {
	ExecuteCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			WorkingDir  string
			RootDir     string
			ProjectPath string
			OutputPath  string
			Flags       []string
		}
		Returns struct {
			Error error
		}
		Stub func(string, string, string, string, []string) error
	}
}

func (f *PublishProcess) Execute(param1 string, param2 string, param3 string, param4 string, param5 []string) error {
	f.ExecuteCall.Lock()
	defer f.ExecuteCall.Unlock()
	f.ExecuteCall.CallCount++
	f.ExecuteCall.Receives.WorkingDir = param1
	f.ExecuteCall.Receives.RootDir = param2
	f.ExecuteCall.Receives.ProjectPath = param3
	f.ExecuteCall.Receives.OutputPath = param4
	f.ExecuteCall.Receives.Flags = param5
	if f.ExecuteCall.Stub != nil {
		return f.ExecuteCall.Stub(param1, param2, param3, param4, param5)
	}
	return f.ExecuteCall.Returns.Error
}
