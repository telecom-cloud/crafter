package options

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/telecom-cloud/crafter/pkg/meta"
	"github.com/telecom-cloud/crafter/pkg/util"
	"github.com/telecom-cloud/crafter/pkg/util/logs"
	"github.com/urfave/cli/v2"
)

type Option struct {
	// Mode              meta.Mode // operating mode（0-compiler, 1-plugin)
	ServiceGroup   string
	CmdType        string // command type
	Verbose        bool   // print verbose log
	Cwd            string // execution path
	OutDir         string // output path
	ModelDir       string // model path
	ClientDir      string // client path
	BaseDomain     string // request domain
	ForceClientDir string // client dir (not use namespace as a subpath)
	ApiImportDir   string

	IdlType       string   // idl type
	IdlPaths      []string // master idl path
	RawOptPkg     []string // user-specified package import path
	OptPkgMap     map[string]string
	Includes      []string
	PkgPrefix     string
	TrimGoPackage string // trim go_package for protobuf, avoid to generate multiple directory

	Gopath      string // $GOPATH
	Gosrc       string // $GOPATH/src
	Gomod       string
	Gopkg       string // $GOPATH/src/{{gopkg}}
	ServiceName string // service name
	Use         string
	NeedGoMod   bool

	JSONEnumStr          bool
	QueryEnumAsInt       bool
	UnsetOmitempty       bool
	ProtobufCamelJSONTag bool
	ProtocOptions        []string // options to pass through to protoc
	ProtobufPlugins      []string
	SnakeName            bool
	RmTags               []string
	Excludes             []string
	NoRecurse            bool
	HandlerByMethod      bool
	ForceNew             bool
	ForceUpdateClient    bool
	SnakeStyleMiddleware bool
	EnableExtends        bool
	SortRouter           bool

	CustomizeLayout     string
	CustomizeLayoutData string
	CustomizePackage    string
	ModelBackend        string
}

func NewOption() *Option {
	return &Option{
		OptPkgMap:     make(map[string]string),
		Includes:      make([]string, 0, 4),
		Excludes:      make([]string, 0, 4),
		ProtocOptions: make([]string, 0, 4),
	}
}

// Parse initializes a new Option based on its own information
func (opt *Option) Parse(c *cli.Context, cmd string) (*Option, error) {
	// v2 cli cannot put the StringSlice flag to struct, so we need to parse it here
	opt.parseStringSlice(c)
	option := opt.Fork()
	option.CmdType = cmd

	err := option.checkPath()
	if err != nil {
		return nil, err
	}

	err = option.checkIDL()
	if err != nil {
		return nil, err
	}

	err = option.checkPackage()
	if err != nil {
		return nil, err
	}

	return option, nil
}

func (opt *Option) parseStringSlice(c *cli.Context) {
	opt.IdlPaths = c.StringSlice("idl")
	opt.Includes = c.StringSlice("proto_path")
	opt.Excludes = c.StringSlice("exclude_file")
	opt.RawOptPkg = c.StringSlice("option_package")
	opt.ProtocOptions = c.StringSlice("protoc")
	opt.ProtobufPlugins = c.StringSlice("protoc-plugins")
	opt.RmTags = c.StringSlice("rm_tag")
}

func (opt *Option) UpdateByManifest(m *meta.Manifest) {
	if opt.ModelDir == "" && m.ModelDir != "" {
		logs.Infof("use \"model_dir\" in \".sc\" as the model generated dir\n")
		opt.ModelDir = m.ModelDir
	}
}

// checkPath sets the project path and verifies that the model、handler、router and client path is compliant
func (opt *Option) checkPath() error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current path failed: %s", err)
	}
	opt.Cwd = dir
	if opt.OutDir == "" {
		opt.OutDir = dir
	}
	if !filepath.IsAbs(opt.OutDir) {
		ap := filepath.Join(opt.Cwd, opt.OutDir)
		opt.OutDir = ap
	}
	if opt.ModelDir != "" && filepath.IsAbs(opt.ModelDir) {
		return fmt.Errorf("model path %s must be relative to out_dir", opt.ModelDir)
	}
	if opt.ClientDir != "" && filepath.IsAbs(opt.ClientDir) {
		return fmt.Errorf("client path %s must be relative to out_dir", opt.ClientDir)
	}
	return nil
}

// checkIDL check if the idl path exists, set and check the idl type
func (opt *Option) checkIDL() error {
	for i, path := range opt.IdlPaths {
		abPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("idl path %s is not absolute", path)
		}
		ext := filepath.Ext(abPath)
		if ext == "" || ext[0] != '.' {
			return fmt.Errorf("idl path %s is not a valid file", path)
		}
		ext = ext[1:]
		switch ext {
		case meta.IdlProto:
			opt.IdlType = meta.IdlProto
		default:
			return fmt.Errorf("IDL type %s is not supported", ext)
		}
		opt.IdlPaths[i] = abPath
	}
	return nil
}

func (opt *Option) IsUpdate() bool {
	return opt.CmdType == meta.CmdUpdate
}

func (opt *Option) IsNew() bool {
	return opt.CmdType == meta.CmdNew
}

// checkPackage check and set the gopath、 module and package name
func (opt *Option) checkPackage() error {
	gopath, err := util.GetGOPATH()
	if err != nil {
		return fmt.Errorf("get gopath failed: %s", err)
	}
	if gopath == "" {
		return fmt.Errorf("GOPATH is not set")
	}

	opt.Gopath = gopath
	opt.Gosrc = filepath.Join(gopath, "src")

	// Generate the project under gopath, use the relative path as the package name
	if strings.HasPrefix(opt.Cwd, opt.Gosrc) {
		if gopkg, err := filepath.Rel(opt.Gosrc, opt.Cwd); err != nil {
			return fmt.Errorf("get relative path to GOPATH/src failed: %s", err)
		} else {
			opt.Gopkg = gopkg
		}
	}
	if len(opt.Gomod) == 0 { // not specified "go module"
		// search go.mod recursively
		module, path, ok := util.SearchGoMod(opt.Cwd, true)
		if ok { // find go.mod in upper level, use it as project module, don't generate go.mod
			rel, err := filepath.Rel(path, opt.Cwd)
			if err != nil {
				return fmt.Errorf("can not get relative path, err :%v", err)
			}
			opt.Gomod = filepath.Join(module, rel)
			logs.Debugf("find module '%s' from '%s/go.mod', so use it as module name", module, path)
		}
		if len(opt.Gomod) == 0 { // don't find go.mod in upper level, use relative path as module name, generate go.mod
			logs.Debugf("use gopath's relative path '%s' as the module name", opt.Gopkg)
			// gopkg will be "" under non-gopath
			opt.Gomod = opt.Gopkg
			opt.NeedGoMod = true
		}
		opt.Gomod = util.PathToImport(opt.Gomod, "")
	} else { // specified "go module"
		// search go.mod in current path
		module, path, ok := util.SearchGoMod(opt.Cwd, false)
		if ok { // go.mod exists in current path, check module name, don't generate go.mod
			if module != opt.Gomod {
				return fmt.Errorf("module name given by the '-module/mod' option ('%s') is not consist with the name defined in go.mod ('%s' from %s), try to remove '-module/mod' option in your command\n", opt.Gomod, module, path)
			}
		} else { // go.mod don't exist in current path, generate go.mod
			opt.NeedGoMod = true
		}
	}

	if len(opt.Gomod) == 0 {
		return fmt.Errorf("can not get go module, please specify a module name with the '-module/mod' flag")
	}

	if len(opt.RawOptPkg) > 0 {
		opt.OptPkgMap = make(map[string]string, len(opt.RawOptPkg))
		for _, op := range opt.RawOptPkg {
			ps := strings.SplitN(op, "=", 2)
			if len(ps) != 2 {
				return fmt.Errorf("invalid option package: %s", op)
			}
			opt.OptPkgMap[ps[0]] = ps[1]
		}
		opt.RawOptPkg = nil
	}
	return nil
}

func (opt *Option) Pack() ([]string, error) {
	data, err := util.PackArgs(opt)
	if err != nil {
		return nil, fmt.Errorf("pack Option failed: %s", err)
	}
	return data, nil
}

func (opt *Option) Unpack(data []string) error {
	err := util.UnpackArgs(data, opt)
	if err != nil {
		return fmt.Errorf("unpack Option failed: %s", err)
	}
	return nil
}

// Fork can copy its own parameters to a new Option
func (opt *Option) Fork() *Option {
	option := NewOption()
	*option = *opt
	util.CopyString2StringMap(opt.OptPkgMap, option.OptPkgMap)
	util.CopyStringSlice(&opt.Includes, &option.Includes)
	util.CopyStringSlice(&opt.Excludes, &option.Excludes)
	util.CopyStringSlice(&opt.ProtocOptions, &option.ProtocOptions)
	return option
}

func (opt *Option) GetGoPackage() (string, error) {
	if opt.Gomod != "" {
		return opt.Gomod, nil
	} else if opt.Gopkg != "" {
		return opt.Gopkg, nil
	}
	return "", fmt.Errorf("project package name is not set")
}

func IdlTypeToCompiler(idlType string) (string, error) {
	switch idlType {
	case meta.IdlProto:
		return meta.TpCompilerProto, nil
	default:
		return "", fmt.Errorf("IDL type %s is not supported", idlType)
	}
}

func (opt *Option) ModelPackagePrefix() (string, error) {
	ret := opt.Gomod
	if opt.ModelDir == "" {
		path, err := util.RelativePath(meta.ModelDir)
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		if err != nil {
			return "", err
		}
		ret += path
	} else {
		path, err := util.RelativePath(opt.ModelDir)
		if err != nil {
			return "", err
		}
		ret += "/" + path
	}
	return strings.ReplaceAll(ret, string(filepath.Separator), "/"), nil
}

func (opt *Option) ModelOutDir() string {
	ret := opt.OutDir
	if opt.ModelDir == "" {
		ret = filepath.Join(ret, meta.ModelDir)
	} else {
		ret = filepath.Join(ret, opt.ModelDir)
	}
	return ret
}

func (opt *Option) GetModelDir() (string, error) {
	if opt.ModelDir == "" {
		return util.RelativePath(meta.ModelDir)
	}
	return util.RelativePath(opt.ModelDir)
}

func (opt *Option) GetClientDir() (string, error) {
	if opt.ClientDir == "" {
		return "", nil
	}
	return util.RelativePath(opt.ClientDir)
}

func (opt *Option) SetManifest(m *meta.Manifest) {
	m.Version = meta.Version
	m.ModelDir = opt.ModelDir
}
