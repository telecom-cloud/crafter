package util

import (
	"bytes"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/telecom-cloud/crafter/pkg/meta"
)

func GetGOPATH() (gopath string, err error) {
	ps := filepath.SplitList(os.Getenv("GOPATH"))
	if len(ps) > 0 {
		gopath = ps[0]
	}
	if gopath == "" {
		cmd := exec.Command("go", "env", "GOPATH")
		var out bytes.Buffer
		cmd.Stderr = &out
		cmd.Stdout = &out
		if err := cmd.Run(); err == nil {
			gopath = strings.Trim(out.String(), " \t\n\r")
		}
	}
	if gopath == "" {
		ps := GetBuildGoPaths()
		if len(ps) > 0 {
			gopath = ps[0]
		}
	}
	isExist, err := PathExist(gopath)
	if !isExist {
		return "", err
	}
	return strings.Replace(gopath, "/", string(os.PathSeparator), -1), nil
}

// GetBuildGoPaths returns the list of Go path directories.
func GetBuildGoPaths() []string {
	var all []string
	for _, p := range filepath.SplitList(build.Default.GOPATH) {
		if p == "" || p == build.Default.GOROOT {
			continue
		}
		if strings.HasPrefix(p, "~") {
			continue
		}
		all = append(all, p)
	}
	for k, v := range all {
		if strings.HasSuffix(v, "/") || strings.HasSuffix(v, string(os.PathSeparator)) {
			v = v[:len(v)-1]
		}
		all[k] = v
	}
	return all
}

var goModReg = regexp.MustCompile(`^\s*module\s+(\S+)\s*`)

// SearchGoMod searches go.mod from the given directory (which must be an absolute path) to
// the root directory. When the go.mod is found, its module name and path will be returned.
func SearchGoMod(cwd string, recurse bool) (moduleName, path string, found bool) {
	for {
		path = filepath.Join(cwd, "go.mod")
		data, err := ioutil.ReadFile(path)
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				m := goModReg.FindStringSubmatch(line)
				if m != nil {
					return m[1], cwd, true
				}
			}
			return fmt.Sprintf("<module name not found in '%s'>", path), path, true
		}

		if !os.IsNotExist(err) {
			return
		}
		if !recurse {
			break
		}
		cwd = filepath.Dir(cwd)
		// the root directory will return itself by using "filepath.Dir()"; to prevent dead loops, so jump out
		if cwd == filepath.Dir(cwd) {
			break
		}
	}
	return
}

func InitGoMod(module string) error {
	isExist, err := PathExist("go.mod")
	if err != nil {
		return err
	}
	if isExist {
		return nil
	}
	gg, err := exec.LookPath("go")
	if err != nil {
		return err
	}
	cmd := &exec.Cmd{
		Path:   gg,
		Args:   []string{"go", "mod", "init", module},
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	return cmd.Run()
}

func IsWindows() bool {
	return meta.SysType == meta.WindowsOS
}
