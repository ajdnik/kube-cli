package executable

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/ajdnik/kube-cli/version"
)

// Info represents contextual information about
// the executable.
type Info struct {
	Path    string
	Name    string
	Arch    string
	OS      string
	Version string
}

// GetInfo returns executable info.
func GetInfo() (Info, error) {
	var info Info
	path, err := os.Executable()
	if err != nil {
		return info, err
	}
	info = Info{
		Path:    path,
		Name:    filepath.Base(path),
		Arch:    runtime.GOARCH,
		OS:      runtime.GOOS,
		Version: version.GetVersion(),
	}
	return info, nil
}
