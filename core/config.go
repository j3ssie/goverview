package core

import (
	"github.com/j3ssie/goverview/libs"
	"os"
	"path"
)

// InitConfig init config
func InitConfig(options *libs.Options) {
	if options.TmpDir == "" {
		options.TmpDir = path.Join(os.TempDir(), "goverview-log")
	}
}
