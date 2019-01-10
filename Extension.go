package arbor

import (
	"github.com/perlin-network/life/exec"
)

// Extension is an Extension that runs in go (or via cgo). You can use this to define native behavior
type Extension interface {
	Run(*exec.VirtualMachine) int64
}

// ExtensionFunc are simple functions that implements a module behaviour
type ExtensionFunc func(vm *exec.VirtualMachine) int64

// Run implements the resolver
func (r ExtensionFunc) Run(vm *exec.VirtualMachine) int64 {
	return r(vm)
}

// Module defines an API to add native go modules (or in c through cgo) to Arbor
// As long as the module is in the path, this should load fine
type Module interface {
	// Resolve attempts to find a Resolver to call that corresponds to the name
	Resolve(string) Extension
	// Name gets the name of the module
	Name() string
}

// Resolver is an implementation of the Module interface
type Resolver struct {
	ModuleName string
	Execers    map[string]Extension
}

// Register registers an extension
func (r *Resolver) Register(name string, e Extension) bool {
	r.Execers[name] = e
	return true
}

// Resolve finds the extenstion that was registeredcd
func (r *Resolver) Resolve(name string) Extension {
	e, ok := r.Execers[name]
	if !ok {
		return nil
	}
	return e
}

// Name get the module name
func (r *Resolver) Name() string {
	return r.ModuleName
}
