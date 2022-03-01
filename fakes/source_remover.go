package fakes

import "sync"

type SourceRemover struct {
	RemoveCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			WorkingDir       string
			PublishOutputDir string
			ExcludedFiles    []string
		}
		Returns struct {
			Error error
		}
		Stub func(string, string, ...string) error
	}
}

func (f *SourceRemover) Remove(param1 string, param2 string, param3 ...string) error {
	f.RemoveCall.Lock()
	defer f.RemoveCall.Unlock()
	f.RemoveCall.CallCount++
	f.RemoveCall.Receives.WorkingDir = param1
	f.RemoveCall.Receives.PublishOutputDir = param2
	f.RemoveCall.Receives.ExcludedFiles = param3
	if f.RemoveCall.Stub != nil {
		return f.RemoveCall.Stub(param1, param2, param3...)
	}
	return f.RemoveCall.Returns.Error
}
