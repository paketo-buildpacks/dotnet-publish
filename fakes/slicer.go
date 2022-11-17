package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit/v2"
)

type Slicer struct {
	SliceCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			AssetsFile string
		}
		Returns struct {
			Pkgs      packit.Slice
			EarlyPkgs packit.Slice
			Projects  packit.Slice
			Err       error
		}
		Stub func(string) (packit.Slice, packit.Slice, packit.Slice, error)
	}
}

func (f *Slicer) Slice(param1 string) (packit.Slice, packit.Slice, packit.Slice, error) {
	f.SliceCall.mutex.Lock()
	defer f.SliceCall.mutex.Unlock()
	f.SliceCall.CallCount++
	f.SliceCall.Receives.AssetsFile = param1
	if f.SliceCall.Stub != nil {
		return f.SliceCall.Stub(param1)
	}
	return f.SliceCall.Returns.Pkgs, f.SliceCall.Returns.EarlyPkgs, f.SliceCall.Returns.Projects, f.SliceCall.Returns.Err
}
