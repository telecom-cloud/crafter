package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/telecom-cloud/crafter/cmd/cft/app/options"
	"github.com/telecom-cloud/crafter/pkg/generator"
	"github.com/telecom-cloud/crafter/pkg/meta"
	"github.com/telecom-cloud/crafter/pkg/plugin/protobuf"
	"github.com/telecom-cloud/crafter/pkg/template"
	"github.com/telecom-cloud/crafter/pkg/util"
	"github.com/telecom-cloud/crafter/pkg/util/logs"
)

// global opts. MUST fork it when use
var globalOpts = options.NewOption()

func PluginMode() {
	mode := os.Getenv(meta.EnvPluginMode)
	if len(os.Args) <= 1 && mode != "" {
		switch mode {
		case meta.ProtocPluginName:
			plugin := new(protobuf.Plugin)
			os.Exit(plugin.Run())
		case meta.ThriftPluginName:
		}
	}
}

func New(c *cli.Context) error {
	opts, err := globalOpts.Parse(c, meta.CmdNew)
	if err != nil {
		return cli.Exit(err, meta.LoadError)
	}
	setLogVerbose(opts.Verbose)
	logs.Debugf("opts: %#v\n", opts)

	exist, err := util.PathExist(filepath.Join(opts.OutDir, meta.ManifestFile))
	if err != nil {
		return cli.Exit(err, meta.LoadError)
	}

	if exist && !opts.ForceNew {
		return cli.Exit(fmt.Errorf("the current is already a crafter project, if you want to regenerate it you can specify \"-force\""), meta.LoadError)
	}

	err = generateLayout(opts)
	if err != nil {
		return cli.Exit(err, meta.GenerateLayoutError)
	}

	err = TriggerPlugin(opts)
	if err != nil {
		return cli.Exit(err, meta.PluginError)
	}
	// ".cft" file converges to the cft tool
	manifest := new(meta.Manifest)
	opts.SetManifest(manifest)
	err = manifest.Persist(opts.OutDir)
	if err != nil {
		return cli.Exit(fmt.Errorf("persist manifest failed: %v", err), meta.PersistError)
	}

	return nil
}

func Update(c *cli.Context) error {
	// begin to update
	opts, err := globalOpts.Parse(c, meta.CmdUpdate)
	if err != nil {
		return cli.Exit(err, meta.LoadError)
	}
	setLogVerbose(opts.Verbose)
	logs.Debugf("opts: %#v\n", opts)

	manifest := new(meta.Manifest)
	err = manifest.InitAndValidate(opts.OutDir)
	if err != nil {
		return cli.Exit(err, meta.LoadError)
	}
	opts.UpdateByManifest(manifest)

	err = TriggerPlugin(opts)
	if err != nil {
		return cli.Exit(err, meta.PluginError)
	}
	opts.SetManifest(manifest)
	err = manifest.Persist(opts.OutDir)
	if err != nil {
		return cli.Exit(fmt.Errorf("persist manifest failed: %v", err), meta.PersistError)
	}

	return nil
}

func Model(c *cli.Context) error {
	opts, err := globalOpts.Parse(c, meta.CmdModel)
	if err != nil {
		return cli.Exit(err, meta.LoadError)
	}
	setLogVerbose(opts.Verbose)
	logs.Debugf("opts: %#v\n", opts)

	err = TriggerPlugin(opts)
	if err != nil {
		return cli.Exit(err, meta.PluginError)
	}

	return nil
}

func Client(c *cli.Context) error {
	opts, err := globalOpts.Parse(c, meta.CmdClient)
	if err != nil {
		return cli.Exit(err, meta.LoadError)
	}
	setLogVerbose(opts.Verbose)
	logs.Debugf("opts: %#v\n", opts)

	err = TriggerPlugin(opts)
	if err != nil {
		return cli.Exit(err, meta.PluginError)
	}

	return nil
}

func Error(c *cli.Context) error {
	opts, err := globalOpts.Parse(c, meta.CmdError)
	if err != nil {
		return cli.Exit(err, meta.LoadError)
	}
	setLogVerbose(opts.Verbose)
	logs.Debugf("opts: %#v\n", opts)

	err = TriggerPlugin(opts)
	if err != nil {
		return cli.Exit(err, meta.PluginError)
	}

	return nil
}

func Doc(c *cli.Context) error {
	opts, err := globalOpts.Parse(c, meta.CmdDoc)
	if err != nil {
		return cli.Exit(err, meta.LoadError)
	}
	setLogVerbose(opts.Verbose)
	logs.Debugf("opts: %#v\n", opts)

	err = TriggerPlugin(opts)
	if err != nil {
		return cli.Exit(err, meta.PluginError)
	}

	return nil
}

func NewCommand() *cli.App {
	// flags
	verboseFlag := cli.BoolFlag{Name: "verbose,vv", Usage: "turn on verbose mode", Destination: &globalOpts.Verbose}
	serviceGroupFlag := cli.StringFlag{Name: "service_group,sg", Usage: "specify the service group", Destination: &globalOpts.ServiceGroup}

	idlFlag := cli.StringSliceFlag{Name: "idl", Usage: "Specify the IDL file path. (.proto)"}
	moduleFlag := cli.StringFlag{Name: "module", Aliases: []string{"mod"}, Usage: "Specify the Go module name.", Destination: &globalOpts.Gomod}
	serviceNameFlag := cli.StringFlag{Name: "service", Usage: "Specify the service name.", Destination: &globalOpts.ServiceName}
	outDirFlag := cli.StringFlag{Name: "out_dir", Usage: "Specify the project path.", Destination: &globalOpts.OutDir}
	modelDirFlag := cli.StringFlag{Name: "model_dir", Usage: "Specify the model relative path (based on \"out_dir\").", Destination: &globalOpts.ModelDir}
	baseDomainFlag := cli.StringFlag{Name: "base_domain", Usage: "Specify the request domain.", Destination: &globalOpts.BaseDomain}
	clientDirFlag := cli.StringFlag{Name: "client_dir", Usage: "Specify the client path. If not specified, IDL generated path is used for 'client' command; no client code is generated for 'new' command", Destination: &globalOpts.ClientDir}
	forceClientDirFlag := cli.StringFlag{Name: "force_client_dir", Usage: "Specify the client path, and won't use namespaces as subpaths", Destination: &globalOpts.ForceClientDir}

	optPkgFlag := cli.StringSliceFlag{Name: "option_package", Aliases: []string{"P"}, Usage: "Specify the package path. ({include_path}={import_path})"}
	includesFlag := cli.StringSliceFlag{Name: "proto_path", Aliases: []string{"I"}, Usage: "Add an IDL search path for includes. (Valid only if idl is protobuf)"}
	excludeFilesFlag := cli.StringSliceFlag{Name: "exclude_file", Aliases: []string{"E"}, Usage: "Specify the files that do not need to be updated."}
	protoOptionsFlag := cli.StringSliceFlag{Name: "protoc", Aliases: []string{"p"}, Usage: "Specify arguments for the protoc. ({flag}={value})"}
	protoPluginsFlag := cli.StringSliceFlag{Name: "protoc-plugins", Usage: "Specify plugins for the protoc. ({plugin_name}:{options}:{out_dir})"}
	noRecurseFlag := cli.BoolFlag{Name: "no_recurse", Usage: "Generate master model only.", Destination: &globalOpts.NoRecurse}
	forceNewFlag := cli.BoolFlag{Name: "force", Aliases: []string{"f"}, Usage: "Force new a project, which will overwrite the generated files", Destination: &globalOpts.ForceNew}
	forceUpdateClientFlag := cli.BoolFlag{Name: "force_client", Usage: "Force update 'crafter_client.go'", Destination: &globalOpts.ForceUpdateClient}

	queryEnumIntFlag := cli.BoolFlag{Name: "query_enumint", Usage: "Use num instead of string for query enum parameter.", Destination: &globalOpts.QueryEnumAsInt}
	unsetOmitemptyFlag := cli.BoolFlag{Name: "unset_omitempty", Usage: "Remove 'omitempty' tag for generated struct.", Destination: &globalOpts.UnsetOmitempty}
	protoCamelJSONTag := cli.BoolFlag{Name: "pb_camel_json_tag", Usage: "Convert Name style for json tag to camel(Only works protobuf).", Destination: &globalOpts.ProtobufCamelJSONTag}
	snakeNameFlag := cli.BoolFlag{Name: "snake_tag", Usage: "Use snake_case style naming for tags. (Only works for 'form', 'query', 'json')", Destination: &globalOpts.SnakeName}
	rmTagFlag := cli.StringSliceFlag{Name: "rm_tag", Usage: "Remove the default tag(json/query/form). If the annotation tag is set explicitly, it will not be removed."}
	customLayout := cli.StringFlag{Name: "customize_layout", Usage: "Specify the path for layout template.", Destination: &globalOpts.CustomizeLayout}
	customLayoutData := cli.StringFlag{Name: "customize_layout_data_path", Usage: "Specify the path for layout template render data.", Destination: &globalOpts.CustomizeLayoutData}
	customPackage := cli.StringFlag{Name: "customize_package", Usage: "Specify the path for package template.", Destination: &globalOpts.CustomizePackage}
	trimGoPackage := cli.StringFlag{Name: "trim_gopackage", Aliases: []string{"trim_pkg"}, Usage: "Trim the prefix of go_package for protobuf.", Destination: &globalOpts.TrimGoPackage}

	// app
	app := cli.NewApp()
	app.Name = "cft"
	app.Usage = "A idl parser and code generator for Crafter projects"
	app.Version = meta.Version
	// The default separator for multiple parameters is modified to ";"
	app.SliceFlagSeparator = ";"

	// global flags
	app.Flags = []cli.Flag{
		&verboseFlag,
	}

	// Commands
	app.Commands = []*cli.Command{
		{
			Name:  meta.CmdNew,
			Usage: "Generate a new Crafter project",
			Flags: []cli.Flag{
				&idlFlag,
				&serviceNameFlag,
				&moduleFlag,
				&outDirFlag,
				&modelDirFlag,
				&clientDirFlag,

				&includesFlag,
				&protoOptionsFlag,
				&optPkgFlag,
				&trimGoPackage,
				&noRecurseFlag,
				&forceNewFlag,

				&unsetOmitemptyFlag,
				&protoCamelJSONTag,
				&snakeNameFlag,
				&rmTagFlag,
				&excludeFilesFlag,
				&customLayout,
				&customLayoutData,
				&customPackage,
				&protoPluginsFlag,
			},
			Action: New,
		},
		{
			Name:  meta.CmdUpdate,
			Usage: "Update an existing Crafter project",
			Flags: []cli.Flag{
				&idlFlag,
				&moduleFlag,
				&outDirFlag,
				&modelDirFlag,
				&clientDirFlag,
				&includesFlag,
				&protoOptionsFlag,
				&optPkgFlag,
				&trimGoPackage,
				&noRecurseFlag,
				&unsetOmitemptyFlag,
				&protoCamelJSONTag,
				&snakeNameFlag,
				&rmTagFlag,
				&excludeFilesFlag,
				&customPackage,
				&protoPluginsFlag,
			},
			Action: Update,
		},
		{
			Name:  meta.CmdModel,
			Usage: "Generate model code only",
			Flags: []cli.Flag{
				&idlFlag,
				&moduleFlag,
				&outDirFlag,
				&modelDirFlag,

				&includesFlag,
				&protoOptionsFlag,
				&noRecurseFlag,
				&trimGoPackage,

				&unsetOmitemptyFlag,
				&protoCamelJSONTag,
				&snakeNameFlag,
				&rmTagFlag,
				&excludeFilesFlag,
			},
			Action: Model,
		},
		{
			Name:  meta.CmdClient,
			Usage: "Generate crafter client based on IDL",
			Flags: []cli.Flag{
				&serviceGroupFlag,
				&idlFlag,
				&moduleFlag,
				&baseDomainFlag,
				&modelDirFlag,
				&clientDirFlag,
				&forceClientDirFlag,
				&forceUpdateClientFlag,

				&includesFlag,
				&protoOptionsFlag,
				&noRecurseFlag,
				&trimGoPackage,

				&queryEnumIntFlag,
				&unsetOmitemptyFlag,
				&protoCamelJSONTag,
				&snakeNameFlag,
				&rmTagFlag,
				&excludeFilesFlag,
				&customPackage,
				&protoPluginsFlag,
			},
			Action: Client,
		},
		{
			Name:  meta.CmdError,
			Usage: "Generate error code only",
			Flags: []cli.Flag{
				&idlFlag,
				&moduleFlag,
				&outDirFlag,
				&modelDirFlag,

				&includesFlag,
				&protoOptionsFlag,
				&noRecurseFlag,
				&trimGoPackage,

				&unsetOmitemptyFlag,
				&protoCamelJSONTag,
				&snakeNameFlag,
				&rmTagFlag,
				&excludeFilesFlag,
			},
			Action: Error,
		},
		{
			Name:  meta.CmdDoc,
			Usage: "Generate doc only",
			Flags: []cli.Flag{
				&idlFlag,
				&moduleFlag,
				&outDirFlag,
				&modelDirFlag,

				&includesFlag,
				&protoOptionsFlag,
				&noRecurseFlag,
				&trimGoPackage,

				&unsetOmitemptyFlag,
				&protoCamelJSONTag,
				&snakeNameFlag,
				&rmTagFlag,
				&excludeFilesFlag,
			},
			Action: Doc,
		},
	}
	return app
}

func setLogVerbose(verbose bool) {
	if verbose {
		logs.SetLevel(logs.LevelDebug)
	} else {
		logs.SetLevel(logs.LevelWarn)
	}
}

func generateLayout(opts *options.Option) error {
	lg := &generator.LayoutGenerator{
		TemplateGenerator: template.TemplateGenerator{
			OutputDir: opts.OutDir,
			Excludes:  opts.Excludes,
		},
	}

	layout := generator.Layout{
		GoModule:    opts.Gomod,
		ServiceName: opts.ServiceName,
		HasIdl:      0 != len(opts.IdlPaths),
		ModelDir:    opts.ModelDir,
		NeedGoMod:   opts.NeedGoMod,
	}

	if opts.CustomizeLayout == "" {
		// generate by default
		err := lg.GenerateByService(layout)
		if err != nil {
			return fmt.Errorf("generating layout failed: %v", err)
		}
	} else {
		// generate by customized layout
		configPath, dataPath := opts.CustomizeLayout, opts.CustomizeLayoutData
		logs.Infof("get customized layout info, layout_config_path: %s, template_data_path: %s", configPath, dataPath)
		exist, err := util.PathExist(configPath)
		if err != nil {
			return fmt.Errorf("check customized layout config file exist failed: %v", err)
		}
		if !exist {
			return errors.New("layout_config_path doesn't exist")
		}
		lg.ConfigPath = configPath
		// generate by service info
		if dataPath == "" {
			err := lg.GenerateByService(layout)
			if err != nil {
				return fmt.Errorf("generating layout failed: %v", err)
			}
		} else {
			// generate by customized data
			err := lg.GenerateByConfig(dataPath)
			if err != nil {
				return fmt.Errorf("generating layout failed: %v", err)
			}
		}
	}

	err := lg.Persist()
	if err != nil {
		return fmt.Errorf("generating layout failed: %v", err)
	}
	return nil
}

func TriggerPlugin(opts *options.Option) error {
	if len(opts.IdlPaths) == 0 {
		return nil
	}
	cmd, err := options.BuildPluginCmd(opts)
	if err != nil {
		return fmt.Errorf("build plugin command failed: %v", err)
	}

	compiler, err := options.IdlTypeToCompiler(opts.IdlType)
	if err != nil {
		return fmt.Errorf("get compiler failed: %v", err)
	}

	logs.Debugf("begin to trigger plugin, compiler: %s, idl_paths: %v", compiler, opts.IdlPaths)
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("plugin %s_gen_crafter returns error: %v, cause:\n%v", compiler, err, string(buf))
	}

	// the plugin returned the log.
	if len(buf) != 0 {
		fmt.Println(string(buf))
	}
	logs.Debugf("end run plugin %s_gen_crafter", compiler)
	return nil
}
