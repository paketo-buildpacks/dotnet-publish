package fakes

import "sync"

type RootManager struct {
	SetupCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Root         string
			ExistingRoot string
			SdkLocation  string
		}
		Returns struct {
			Error error
		}
		Stub func(string, string, string) error
	}
}

func (f *RootManager) Setup(param1 string, param2 string, param3 string) error {
	f.SetupCall.Lock()
	defer f.SetupCall.Unlock()
	f.SetupCall.CallCount++
	f.SetupCall.Receives.Root = param1
	f.SetupCall.Receives.ExistingRoot = param2
	f.SetupCall.Receives.SdkLocation = param3
	if f.SetupCall.Stub != nil {
		return f.SetupCall.Stub(param1, param2, param3)
	}
	return f.SetupCall.Returns.Error
}
