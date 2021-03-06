package arbor

import (
	"fmt"

	"io/ioutil"
	"os"
	"path"
)

// LoadFile loads our file to run. It returns the content as a byte array, bool if it is pre-compiled, and an error if anything is broken
func LoadFile(fileName string, isWasm bool) ([]byte, bool, error) {
	if isWasm {
		content, _, err := loadFileNoCache(fileName)
		return content, true, err
	}
	return maybeLoadCacheFile(fileName)
}

// loadFileNoCache just loads a straight file with out checking the cache first
func loadFileNoCache(fileName string) ([]byte, bool, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, false, err
	}
	fileLoc := path.Join(pwd, fileName)
	data, err := ioutil.ReadFile(fileLoc)
	if err != nil {
		return nil, false, err
	}
	return data, false, nil
}

// maybeLoadCacheFile loads the cached file first
func maybeLoadCacheFile(fileName string) ([]byte, bool, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, false, err
	}
	fileLoc := path.Join(pwd, fileName)
	fileInfo, err := os.Stat(fileLoc)
	if err != nil {
		return nil, false, err
	}
	cacheName := path.Join(".ab_cache", fmt.Sprintf("%s.abc", fileName))
	cacheLoc := path.Join(pwd, cacheName)
	cacheInfo, err := os.Stat(cacheLoc)
	if !os.IsNotExist(err) && fileInfo.ModTime().After(cacheInfo.ModTime()) {
		content, _, err := loadFileNoCache(fileName)
		return content, true, err
	}
	return loadFileNoCache(fileName)
}

// RunWasm runs a Wasm file
func RunWasm(wasmCode []byte, entrypoint string, extensions ...string) (int64, error) {
	vm, err := NewVirtualMachine(wasmCode, entrypoint, extensions...)
	if err != nil {
		return int64(-1), err
	}
	ret, err := vm.Run()
	if err != nil {
		vm.PrintStackTrace()
		return int64(-1), err
	}
	return ret, nil
	// return int64(-1), fmt.Errorf("not implemented")
}

// RunWat runs a Wat file
func RunWat() (int64, error) {
	return int64(-1), fmt.Errorf("not implemented")
}

//RunArbor runs an arbor file
func RunArbor(wasmCode []byte, entrypoint string) (int64, error) {
	return int64(-1), fmt.Errorf("not implemented")
}
