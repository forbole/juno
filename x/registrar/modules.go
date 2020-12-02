package registrar

import (
	"github.com/desmos-labs/juno/x"
	"github.com/rs/zerolog/log"
)

var (
	modules x.Modules
)

// RegisterModules registers the given modules so that they can be used later by the GetModules method.
func RegisterModules(m ...x.Module) {
	modules = m
}

// GetModules returns the list of module implementations based on the given module names.
// For each module name that is specified but not found, a warning log is printed.
func GetModules(names []string) []x.Module {
	var modulesImpls []x.Module
	for _, name := range names {
		module, found := modules.FindByName(name)
		if found {
			modulesImpls = append(modulesImpls, module)
		} else {
			log.Warn().Msgf("%s module is required but not registered. Be sure to register it using registrar.RegisterModule", name)
		}
	}
	return modulesImpls
}
