package util

import (
	"fmt"
	"go/format"
	"path/filepath"
	"strings"
)

type File struct {
	Path        string
	Content     string
	NoRepeat    bool
	FileTplName string
}

// Lint is used to statically analyze and format go code
func (file *File) Lint() error {
	name := filepath.Base(file.Path)
	if strings.HasSuffix(name, ".go") {
		out, err := format.Source(Str2Bytes(file.Content))
		if err != nil {
			return fmt.Errorf("lint file '%s' failed, err: %v", name, err.Error())
		}
		file.Content = Bytes2Str(out)
	}
	return nil
}
