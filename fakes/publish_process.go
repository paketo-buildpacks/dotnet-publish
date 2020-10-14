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
		}
		Returns struct {
			Error error
		}
		Stub func(string, string, string) error
	}
}

func (f *PublishProcess) Execute(param1 string, param2 string, param3 string) error {
	f.ExecuteCall.Lock()
	defer f.ExecuteCall.Unlock()
	f.ExecuteCall.CallCount++
	f.ExecuteCall.Receives.WorkingDir = param1
	f.ExecuteCall.Receives.RootDir = param2
	f.ExecuteCall.Receives.ProjectPath = param3
	if f.ExecuteCall.Stub != nil {
		return f.ExecuteCall.Stub(param1, param2, param3)
	}
	return f.ExecuteCall.Returns.Error
}
