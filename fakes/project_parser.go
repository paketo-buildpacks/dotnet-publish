package fakes

import "sync"

type ProjectParser struct {
	FindProjectFileCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Root string
		}
		Returns struct {
			String string
			Error  error
		}
		Stub func(string) (string, error)
	}
	NPMIsRequiredCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Path string
		}
		Returns struct {
			Bool  bool
			Error error
		}
		Stub func(string) (bool, error)
	}
	NodeIsRequiredCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Path string
		}
		Returns struct {
			Bool  bool
			Error error
		}
		Stub func(string) (bool, error)
	}
	ParseVersionCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Path string
		}
		Returns struct {
			String string
			Error  error
		}
		Stub func(string) (string, error)
	}
}

func (f *ProjectParser) FindProjectFile(param1 string) (string, error) {
	f.FindProjectFileCall.mutex.Lock()
	defer f.FindProjectFileCall.mutex.Unlock()
	f.FindProjectFileCall.CallCount++
	f.FindProjectFileCall.Receives.Root = param1
	if f.FindProjectFileCall.Stub != nil {
		return f.FindProjectFileCall.Stub(param1)
	}
	return f.FindProjectFileCall.Returns.String, f.FindProjectFileCall.Returns.Error
}
func (f *ProjectParser) NPMIsRequired(param1 string) (bool, error) {
	f.NPMIsRequiredCall.mutex.Lock()
	defer f.NPMIsRequiredCall.mutex.Unlock()
	f.NPMIsRequiredCall.CallCount++
	f.NPMIsRequiredCall.Receives.Path = param1
	if f.NPMIsRequiredCall.Stub != nil {
		return f.NPMIsRequiredCall.Stub(param1)
	}
	return f.NPMIsRequiredCall.Returns.Bool, f.NPMIsRequiredCall.Returns.Error
}
func (f *ProjectParser) NodeIsRequired(param1 string) (bool, error) {
	f.NodeIsRequiredCall.mutex.Lock()
	defer f.NodeIsRequiredCall.mutex.Unlock()
	f.NodeIsRequiredCall.CallCount++
	f.NodeIsRequiredCall.Receives.Path = param1
	if f.NodeIsRequiredCall.Stub != nil {
		return f.NodeIsRequiredCall.Stub(param1)
	}
	return f.NodeIsRequiredCall.Returns.Bool, f.NodeIsRequiredCall.Returns.Error
}
func (f *ProjectParser) ParseVersion(param1 string) (string, error) {
	f.ParseVersionCall.mutex.Lock()
	defer f.ParseVersionCall.mutex.Unlock()
	f.ParseVersionCall.CallCount++
	f.ParseVersionCall.Receives.Path = param1
	if f.ParseVersionCall.Stub != nil {
		return f.ParseVersionCall.Stub(param1)
	}
	return f.ParseVersionCall.Returns.String, f.ParseVersionCall.Returns.Error
}
