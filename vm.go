package arbor

import (
	"fmt"
	"plugin"

	"github.com/perlin-network/life/exec"
)

//VM is the arbor virtual machine
type VM struct {
	Life       *exec.VirtualMachine
	passedVM   *exec.VirtualMachine
	entrypoint string
	entryID    int
	StackTop   int64
	CallStack  []int64
	resolvers  map[string]Module
}

// NewVirtualMachine returns a new arbor VirtualMachine
func NewVirtualMachine(wasmCode []byte, entrypoint string, paths ...string) (*VM, error) {
	realVM := new(VM)
	realVM.resolvers = make(map[string]Module)
	if err := realVM.LoadModules(paths...); err != nil {
		return nil, err
	}
	vm, err := exec.NewVirtualMachine(wasmCode, exec.VMConfig{}, realVM, nil)
	if err != nil {
		return nil, err
	}
	realVM.Life = vm
	realVM.entrypoint = entrypoint
	return realVM, nil

}

// Run runs the virtual machine
func (v *VM) Run() (int64, error) {
	entryID, ok := v.Life.GetFunctionExport(v.entrypoint) // can be changed to your own exported function
	if !ok {
		return int64(-1), fmt.Errorf("entry function not found")
	}
	v.entryID = entryID
	ret, err := v.Life.Run(v.entryID)
	if err != nil {
		return int64(-1), err
	}
	return ret, nil
}

// LoadModules loads a list of modules
func (v *VM) LoadModules(paths ...string) error {
	for _, path := range paths {
		if err := v.Load(path); err != nil {
			return err
		}
	}
	return nil
}

// Load loads a module from a path
func (v *VM) Load(path string) error {
	fmt.Println("Loading a module", path)
	plug, err := plugin.Open(path)
	if err != nil {
		return err
	}
	resolver, err := plug.Lookup("Env")
	if err != nil {
		return err
	}
	if module, ok := resolver.(Module); ok {
		fmt.Println("Module has been loaded!")
		v.resolvers[module.Name()] = module
		return nil
	}
	return fmt.Errorf("Could not open extentions")
}

// PrintStackTrace prints the stack Trace
func (v *VM) PrintStackTrace() {
	fmt.Println("Stack Top:", v.StackTop, "MemoryMax:", v.Life.Config.MaxMemoryPages, "Memory Size:", len(v.Life.Memory))
	fmt.Println("\tMax Call Stack:", v.Life.Memory[v.StackTop])
	v.Life.PrintStackTrace()
	fmt.Println(v.Life.StackTrace)
}

// StackPush pushes a stack pointer down
func (v *VM) StackPush(_ *exec.VirtualMachine) int64 {
	v.CallStack = append(v.CallStack, v.StackTop)
	return v.StackTop
}

// IncrementStack adds values to the stack
func (v *VM) IncrementStack(_ *exec.VirtualMachine) int64 {
	fmt.Println("incrementing")
	increment := v.Life.GetCurrentFrame().Locals[0]
	fmt.Println("Executing access")
	val := v.Life.Memory[v.StackTop : v.StackTop+4]
	fmt.Println("done Executing access")
	fmt.Println(val)
	fmt.Println("done")
	v.StackTop += increment
	return v.StackTop
}

// GetStackTop just returns the top of the stack
func (v *VM) GetStackTop(_ *exec.VirtualMachine) int64 { return v.StackTop }

// StackPop pops a call stack off
func (v *VM) StackPop(_ *exec.VirtualMachine) int64 {
	fmt.Println("Popping stack")
	lastEntry := len(v.CallStack) - 1
	newTop, newStack := v.CallStack[lastEntry], v.CallStack[:lastEntry]
	fmt.Println("poping stack", v.StackTop, "->", newTop)
	v.StackTop = newTop
	v.CallStack = newStack
	return v.StackTop
}

//ResolveFunc finds the function you are looking for
func (v *VM) ResolveFunc(module, field string) exec.FunctionImport {
	if module == "env" {
		if field == "__popstack__" {
			return v.StackPop
		}
		if field == "__pushstack__" {
			return v.StackPush
		}
		if field == "__incrementstack__" {
			return v.IncrementStack
		}
		if field == "__stacktop__" {
			return v.GetStackTop
		}
	}
	mod, ok := v.resolvers[module]
	fmt.Println(v.resolvers)
	if !ok {
		panic(fmt.Errorf("unknown import resolved: %s", module))
	}
	ext := mod.Resolve(field)
	if ext == nil {
		panic(fmt.Errorf("%s has no function %s", module, field))
	}
	return v.prepFunctionForExecution(ext)
}

// RegisterModule registers a module for resolving
func (v *VM) RegisterModule(mod Module) {
	v.resolvers[mod.Name()] = mod
}

//ResolveGlobal just dies
func (v *VM) ResolveGlobal(module, field string) int64 {
	switch module {
	case "env":
		if field == "STACKTOP_ASM" {
			return v.StackTop
		}
		panic(fmt.Errorf("%s global not found", field))
	}
	panic("we're not resolving global variables for now")
}

func (v *VM) prepFunctionForExecution(ext Extension) exec.FunctionImport {
	return func(vm *exec.VirtualMachine) int64 {
		fmt.Println("Fucking peice of shit")
		v.passedVM = vm
		retVal := ext.Run(v)
		v.passedVM = nil
		return retVal
	}
}
