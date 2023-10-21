package shared

import (
	"fmt"
	"runtime"

	"github.com/juicity/juicity/config"
)

func PrintVersion(cgoEnabled int) {
	fmt.Printf("juicity-client version %v\n", config.Version)
	fmt.Printf("go version %v %v/%v\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Printf("CGO_ENABLED: %v\n", cgoEnabled)
}

func GetVersion(cgoEnabled int) string {
	return fmt.Sprintf("juicity-client version %v\ngo version %v %v/%v\nCGO_ENABLED: %v", config.Version, runtime.Version(), runtime.GOOS, runtime.GOARCH, cgoEnabled)
}
