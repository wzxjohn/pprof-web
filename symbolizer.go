package main

import (
	"github.com/google/pprof/driver"
	"github.com/google/pprof/profile"
)

type webSym struct {
	Address string
}

func (w *webSym) Symbolize(_ string, _ driver.MappingSources, prof *profile.Profile) error {
	if len(prof.Mapping) > 0 {
		if prof.Mapping[0].File == "" {
			prof.Mapping[0].File = "unknown"
		}
		prof.Mapping[0].File += "@" + w.Address
	}
	return nil
}
