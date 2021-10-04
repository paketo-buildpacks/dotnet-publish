package fakes

import "sync"

type Dotnet struct {
	PublishCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			WorkingDir  string
			RootDir     string
			ProjectPath string
			OutputPath  string
		}
		Returns struct {
			Error error
		}
		Stub func(string, string, string, string) error
	}
	RestoreCall struct {
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

func (f *Dotnet) Publish(param1 string, param2 string, param3 string, param4 string) error {
	f.PublishCall.Lock()
	defer f.PublishCall.Unlock()
	f.PublishCall.CallCount++
	f.PublishCall.Receives.WorkingDir = param1
	f.PublishCall.Receives.RootDir = param2
	f.PublishCall.Receives.ProjectPath = param3
	f.PublishCall.Receives.OutputPath = param4
	if f.PublishCall.Stub != nil {
		return f.PublishCall.Stub(param1, param2, param3, param4)
	}
	return f.PublishCall.Returns.Error
}
func (f *Dotnet) Restore(param1 string, param2 string, param3 string) error {
	f.RestoreCall.Lock()
	defer f.RestoreCall.Unlock()
	f.RestoreCall.CallCount++
	f.RestoreCall.Receives.WorkingDir = param1
	f.RestoreCall.Receives.RootDir = param2
	f.RestoreCall.Receives.ProjectPath = param3
	if f.RestoreCall.Stub != nil {
		return f.RestoreCall.Stub(param1, param2, param3)
	}
	return f.RestoreCall.Returns.Error
}
