package options

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/telecom-cloud/crafter/pkg/meta"
	"github.com/telecom-cloud/crafter/pkg/util"
	"github.com/telecom-cloud/crafter/pkg/util/logs"
)

func findToolPath() (string, error) {
	tool := meta.TpCompilerProto

	path, err := exec.LookPath(tool)
	logs.Debugf("[DEBUG]path:%v", path)
	if err != nil {
		goPath, err := util.GetGOPATH()
		if err != nil {
			return "", fmt.Errorf("get 'GOPATH' failed for find %s : %v", tool, path)
		}
		path = filepath.Join(goPath, "bin", tool)
	}

	isExist, err := util.PathExist(path)
	if err != nil {
		return "", fmt.Errorf("check '%s' path error: %v", path, err)
	}

	if !isExist {
		return "", fmt.Errorf("%s is not installed, please install it first", tool)
	}

	return path, nil
}

func BuildPluginCmd(opt *Option) (*exec.Cmd, error) {
	binary, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to detect current executable, err: %v", err)
	}

	optPacks, err := opt.Pack()
	if err != nil {
		return nil, err
	}
	option := strings.Join(optPacks, ",")

	path, err := findToolPath()
	if err != nil {
		return nil, err
	}
	cmd := &exec.Cmd{
		Path: path,
	}

	if opt.IdlType == meta.IdlProto {
		// protoc
		os.Setenv(meta.EnvPluginMode, meta.ProtocPluginName)
		// set crafter plugin
		cmd.Args = append(cmd.Args, meta.TpCompilerProto)
		for _, inc := range opt.Includes {
			cmd.Args = append(cmd.Args, "-I", inc)
		}
		for _, inc := range opt.IdlPaths {
			cmd.Args = append(cmd.Args, "-I", filepath.Dir(inc))
		}
		cmd.Args = append(cmd.Args,
			"--plugin=protoc-gen-crafter="+binary,
			"--crafter_out="+opt.OutDir,
			"--crafter_opt="+option,
		)
		// set protoc other plugin
		for _, p := range opt.ProtobufPlugins {
			pluginParams := strings.Split(p, ":")
			if len(pluginParams) != 3 {
				logs.Warnf("Failed to get the correct protoc plugin parameters for %s. "+
					"Please specify the protoc plugin in the form of \"plugin_name:options:out_dir\"", p)
				os.Exit(1)
			}
			// pluginParams[0] -> plugin name, pluginParams[1] -> plugin options, pluginParams[2] -> out_dir
			cmd.Args = append(cmd.Args,
				fmt.Sprintf("--%s_out=%s", pluginParams[0], pluginParams[2]),
				fmt.Sprintf("--%s_opt=%s", pluginParams[0], pluginParams[1]),
			)
		}
		for _, kv := range opt.ProtocOptions {
			cmd.Args = append(cmd.Args, "--"+kv)
		}
	}

	// set proto path
	cmd.Args = append(cmd.Args, opt.IdlPaths...)
	// print cmd
	logs.Infof(strings.Join(cmd.Args, " "))
	logs.Flush()
	return cmd, nil
}
