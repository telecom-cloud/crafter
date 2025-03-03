package protobuf

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"

	"github.com/telecom-cloud/crafter/cmd/cft/app/options"
	"github.com/telecom-cloud/crafter/pkg/generator"
	"github.com/telecom-cloud/crafter/pkg/generator/model"
	"github.com/telecom-cloud/crafter/pkg/meta"
	tpl "github.com/telecom-cloud/crafter/pkg/template"
	"github.com/telecom-cloud/crafter/pkg/util"
	"github.com/telecom-cloud/crafter/pkg/util/logs"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

type Plugin struct {
	*protogen.Plugin
	Package      string
	Recursive    bool
	OutDir       string
	ModelDir     string
	UseDir       string
	IdlClientDir string
	RmTags       RemoveTags
	PkgMap       map[string]string
	logger       *logs.StdLogger
}

type RemoveTags []string

func (rm *RemoveTags) Exist(tag string) bool {
	for _, rmTag := range *rm {
		if rmTag == tag {
			return true
		}
	}
	return false
}

func (plugin *Plugin) Run() int {
	plugin.setLogger()
	args := &options.Option{}
	defer func() {
		if args == nil {
			return
		}
		if args.Verbose {
			verboseLog := plugin.verboseLogger()
			if len(verboseLog) != 0 {
				fmt.Fprintf(os.Stderr, verboseLog)
			}
		} else {
			warning := plugin.warningLogger()
			if len(warning) != 0 {
				fmt.Fprintf(os.Stderr, warning)
			}
		}
	}()
	// read protoc request
	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		logs.Errorf("read request failed: %s\n", err.Error())
		return meta.PluginError
	}

	req := &pluginpb.CodeGeneratorRequest{}
	err = proto.Unmarshal(in, req)
	if err != nil {
		logs.Errorf("unmarshal request failed: %s\n", err.Error())
		return meta.PluginError
	}

	args, err = plugin.parseArgs(*req.Parameter)
	if err != nil {
		logs.Errorf("parse args failed: %s\n", err.Error())
		return meta.PluginError
	}
	CheckTagOption(args)
	// generate
	err = plugin.Handle(req, args)
	if err != nil {
		logs.Errorf("generate failed: %s\n", err.Error())
		return meta.PluginError
	}
	return 0
}

// setLogger 设置日志记录器
func (plugin *Plugin) setLogger() {
	plugin.logger = logs.NewStdLogger(logs.LevelInfo)
	plugin.logger.Defer = true
	plugin.logger.ErrOnly = true
	logs.SetLogger(plugin.logger)
}

// warningLogger 警告日志
func (plugin *Plugin) warningLogger() string {
	warns := plugin.logger.Warn()
	plugin.logger.Flush()
	logs.SetLogger(logs.NewStdLogger(logs.LevelInfo))
	return warns
}

// verboseLogger 详细日志
func (plugin *Plugin) verboseLogger() string {
	info := plugin.logger.Out()
	warns := plugin.logger.Warn()
	verboseLog := string(info) + warns
	plugin.logger.Flush()
	logs.SetLogger(logs.NewStdLogger(logs.LevelInfo))
	return verboseLog
}

// parseArgs 解析参数
func (plugin *Plugin) parseArgs(param string) (*options.Option, error) {
	args := new(options.Option)
	params := strings.Split(param, ",")
	err := args.Unpack(params)
	if err != nil {
		return nil, err
	}
	plugin.Package, err = args.GetGoPackage()
	if err != nil {
		return nil, err
	}
	plugin.Recursive = !args.NoRecurse
	plugin.ModelDir, err = args.GetModelDir()
	if err != nil {
		return nil, err
	}
	plugin.OutDir = args.OutDir
	plugin.PkgMap = args.OptPkgMap
	plugin.UseDir = args.Use
	return args, nil
}

// OutputResponse 输出protobuf响应结果
func (plugin *Plugin) OutputResponse(resp *pluginpb.CodeGeneratorResponse) error {
	out, err := proto.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal response failed: %s", err.Error())
	}
	_, err = os.Stdout.Write(out)
	if err != nil {
		return fmt.Errorf("write response failed: %s", err.Error())
	}
	return nil
}

// Handle 处理protobuf请求
func (plugin *Plugin) Handle(req *pluginpb.CodeGeneratorRequest, args *options.Option) error {
	// 修复 Go 包路径，确保生成的 Go 文件位于正确的目录结构中
	plugin.fixGoPackage(req, plugin.PkgMap, args.TrimGoPackage)
	// 初始化 Protoc 插件实例
	opts := protogen.Options{}
	gen, err := opts.New(req)
	plugin.Plugin = gen
	plugin.RmTags = args.RmTags
	if err != nil {
		return fmt.Errorf("new protoc plugin failed: %s", err.Error())
	}
	// 开始生成文件
	err = plugin.GenerateFiles(gen)
	if err != nil {
		// 错误处理：将错误信息添加到响应中，并返回错误
		err = fmt.Errorf("generate model file failed: %s", err.Error())
		gen.Error(err)
		resp := gen.Response()
		return plugin.OutputResponse(resp)
	}

	// 处理模型生成命令，使用model子命令无需额外生成代码
	if args.CmdType == meta.CmdModel {
		resp := gen.Response()
		// plugin stop working
		err = plugin.OutputResponse(resp)
		if err != nil {
			return fmt.Errorf("write response failed: %s", err.Error())
		}

		return nil
	}

	// 构建依赖关系映射表
	files := gen.Request.ProtoFile
	maps := make(map[string]*descriptorpb.FileDescriptorProto, len(files))
	for _, file := range files {
		maps[file.GetName()] = file
	}
	// 获取主文件及其依赖项,即非import的proto文件
	main := maps[gen.Request.FileToGenerate[len(gen.Request.FileToGenerate)-1]]
	dependencies := make(map[string]*descriptorpb.FileDescriptorProto, len(main.GetDependency()))
	for _, dep := range main.GetDependency() {
		// 处理缺失依赖文件的情况
		if f, ok := maps[dep]; !ok {
			err := fmt.Errorf("dependency file not found: %s", dep)
			gen.Error(err)
			resp := gen.Response()
			return plugin.OutputResponse(resp)
		} else {
			dependencies[dep] = f
		}
	}

	pkgFiles, err := plugin.generatePackageFiles(main, dependencies, args)
	if err != nil {
		err := fmt.Errorf("generate package files failed: %s", err.Error())
		gen.Error(err)
		resp := gen.Response()
		err2 := plugin.OutputResponse(resp)
		if err2 != nil {
			return err
		}
		return nil
	}

	// 构造插件响应
	resp := gen.Response()
	// all files that need to be generated are returned to protoc
	for _, pkgFile := range pkgFiles {
		filePath := pkgFile.Path
		content := pkgFile.Content
		renderFile := &pluginpb.CodeGeneratorResponse_File{
			Name:    &filePath,
			Content: &content,
		}
		resp.File = append(resp.File, renderFile)
	}

	// plugin stop working
	err = plugin.OutputResponse(resp)
	if err != nil {
		return fmt.Errorf("write response failed: %s", err.Error())
	}

	return nil
}

// fixGoPackage will update go_package to store all the model files in ${model_dir}
func (plugin *Plugin) fixGoPackage(req *pluginpb.CodeGeneratorRequest, pkgMap map[string]string, trimGoPackage string) {
	gopkg := plugin.Package
	for _, f := range req.ProtoFile {
		if strings.HasPrefix(f.GetPackage(), "google.protobuf") {
			continue
		}
		if len(trimGoPackage) != 0 && strings.HasPrefix(f.GetOptions().GetGoPackage(), trimGoPackage) {
			*f.Options.GoPackage = strings.TrimPrefix(*f.Options.GoPackage, trimGoPackage)
		}

		opt := getGoPackage(f, pkgMap)
		if !strings.Contains(opt, gopkg) {
			if strings.HasPrefix(opt, "/") {
				opt = gopkg + opt
			} else {
				opt = gopkg + "/" + opt
			}
		}
		impt, _ := plugin.fixModelPathAndPackage(opt)
		*f.Options.GoPackage = impt
	}
}

// fixModelPathAndPackage will modify the go_package to adapt the go_package of the cft,
// for example adding the go module and model dir.
func (plugin *Plugin) fixModelPathAndPackage(pkg string) (impt, path string) {
	if strings.HasPrefix(pkg, plugin.Package) {
		impt = util.ImportToPathAndConcat(pkg[len(plugin.Package):], "")
	}
	if plugin.ModelDir != "" && plugin.ModelDir != "." {
		modelImpt := util.PathToImport(string(filepath.Separator)+plugin.ModelDir, "")
		// trim model dir for go package
		if strings.HasPrefix(impt, modelImpt) {
			impt = impt[len(modelImpt):]
		}
		impt = util.PathToImport(plugin.ModelDir, "") + impt
	}
	path = util.ImportToPath(impt, "")
	// bugfix: impt may have "/" suffix
	//impt = plugin.PackageDescription + "/" + impt
	impt = filepath.Join(plugin.Package, impt)
	if util.IsWindows() {
		impt = util.PathToImport(impt, "")
	}
	return
}

func (plugin *Plugin) GenerateFiles(pluginPb *protogen.Plugin) error {
	idl := pluginPb.Request.FileToGenerate[len(pluginPb.Request.FileToGenerate)-1]
	pluginPb.SupportedFeatures = gengo.SupportedFeatures
	for _, f := range pluginPb.Files {
		if f.Proto.GetName() == idl {
			err := plugin.GenerateFile(pluginPb, f)
			if err != nil {
				return err
			}
			impt := string(f.GoImportPath)
			if strings.HasPrefix(impt, plugin.Package) {
				impt = impt[len(plugin.Package):]
			}
			plugin.IdlClientDir = impt
		} else if plugin.Recursive {
			if strings.HasPrefix(f.Proto.GetPackage(), "google.protobuf") {
				continue
			}
			err := plugin.GenerateFile(pluginPb, f)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (plugin *Plugin) GenerateFile(gen *protogen.Plugin, f *protogen.File) error {
	// generate go file
	impt := string(f.GoImportPath)
	if strings.HasPrefix(impt, plugin.Package) {
		impt = impt[len(plugin.Package):]
	}
	f.GeneratedFilenamePrefix = filepath.Join(util.ImportToPath(impt, ""), util.BaseName(f.Proto.GetName(), ".proto"))
	f.Generate = true
	// if use third-party model, no model code is generated within the project
	if len(plugin.UseDir) != 0 {
		return nil
	}
	file, err := generateFile(gen, f, plugin.RmTags)
	if err != nil || file == nil {
		return fmt.Errorf("generate file %s failed: %s", f.Proto.GetName(), err.Error())
	}
	return nil
}

// generateFile generates the contents of a .pb.go file.
func generateFile(gen *protogen.Plugin, file *protogen.File, rmTags RemoveTags) (*protogen.GeneratedFile, error) {
	filename := file.GeneratedFilenamePrefix + ".pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	f := newFileInfo(file)

	genStandaloneComments(g, f, int32(FileDescriptorProto_Syntax_field_number))
	genGeneratedHeader(gen, g, f)
	genStandaloneComments(g, f, int32(FileDescriptorProto_Package_field_number))

	packageDoc := genPackageKnownComment(f)
	g.P(packageDoc, "package ", f.GoPackageName)
	g.P()

	// Emit a static check that enforces a minimum version of the proto package.
	if gengo.GenerateVersionMarkers {
		g.P("const (")
		g.P("// Verify that this generated code is sufficiently up-to-date.")
		g.P("_ = ", protoimplPackage.Ident("EnforceVersion"), "(", protoimpl.GenVersion, " - ", protoimplPackage.Ident("MinVersion"), ")")
		g.P("// Verify that runtime/protoimpl is sufficiently up-to-date.")
		g.P("_ = ", protoimplPackage.Ident("EnforceVersion"), "(", protoimplPackage.Ident("MaxVersion"), " - ", protoimpl.GenVersion, ")")
		g.P(")")
		g.P()
	}

	for i, imps := 0, f.Desc.Imports(); i < imps.Len(); i++ {
		genImport(gen, g, f, imps.Get(i))
	}
	for _, enum := range f.allEnums {
		genEnum(g, f, enum)
	}
	var err error
	for _, message := range f.allMessages {
		err = genMessage(g, f, message, rmTags)
		if err != nil {
			return nil, err
		}
	}
	genExtensions(g, f)

	genReflectFileDescriptor(gen, g, f)

	return g, nil
}

func genMessage(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo, rmTags RemoveTags) error {
	if m.Desc.IsMapEntry() {
		return nil
	}

	// Message type declaration.
	g.Annotate(m.GoIdent.GoName, m.Location)
	leadingComments := appendDeprecationSuffix(m.Comments.Leading,
		m.Desc.Options().(*descriptorpb.MessageOptions).GetDeprecated())
	g.P(leadingComments,
		"type ", m.GoIdent, " struct {")
	err := genMessageFields(g, f, m, rmTags)
	if err != nil {
		return err
	}
	g.P("}")
	g.P()

	genMessageKnownFunctions(g, f, m)
	genMessageDefaultDecls(g, f, m)
	genMessageMethods(g, f, m)
	genMessageOneofWrapperTypes(g, f, m)
	return nil
}

func genMessageFields(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo, rmTags RemoveTags) error {
	sf := f.allMessageFieldsByPtr[m]
	genMessageInternalFields(g, f, m, sf)
	var err error
	for _, field := range m.Fields {
		err = genMessageField(g, f, m, field, sf, rmTags)
		if err != nil {
			return err
		}
	}
	return nil
}

func genMessageField(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo, field *protogen.Field, sf *structFields, rmTags RemoveTags) error {
	if oneof := field.Oneof; oneof != nil && !oneof.Desc.IsSynthetic() {
		// It would be a bit simpler to iterate over the oneofs below,
		// but generating the field here keeps the contents of the Go
		// struct in the same order as the contents of the source
		// .proto file.
		if oneof.Fields[0] != field {
			return nil // only generate for first appearance
		}

		tags := structTags{
			{"protobuf_oneof", string(oneof.Desc.Name())},
		}
		if m.isTracked {
			tags = append(tags, gotrackTags...)
		}

		g.Annotate(m.GoIdent.GoName+"."+oneof.GoName, oneof.Location)
		leadingComments := oneof.Comments.Leading
		if leadingComments != "" {
			leadingComments += "\n"
		}
		ss := []string{fmt.Sprintf(" Types that are assignable to %s:\n", oneof.GoName)}
		for _, field := range oneof.Fields {
			ss = append(ss, "\t*"+field.GoIdent.GoName+"\n")
		}
		leadingComments += protogen.Comments(strings.Join(ss, ""))
		g.P(leadingComments,
			oneof.GoName, " ", oneofInterfaceName(oneof), tags)
		sf.append(oneof.GoName)
		return nil
	}
	goType, pointer := fieldGoType(g, f, field)
	if pointer {
		goType = "*" + goType
	}
	tags := structTags{
		{"protobuf", fieldProtobufTagValue(field)},
		//{"json", fieldJSONTagValue(field)},
	}
	if field.Desc.IsMap() {
		key := field.Message.Fields[0]
		val := field.Message.Fields[1]
		tags = append(tags, structTags{
			{"protobuf_key", fieldProtobufTagValue(key)},
			{"protobuf_val", fieldProtobufTagValue(val)},
		}...)
	}

	err := injectTagsToStructTags(field.Desc, &tags, true, rmTags)
	if err != nil {
		return err
	}

	if m.isTracked {
		tags = append(tags, gotrackTags...)
	}

	name := field.GoName
	if field.Desc.IsWeak() {
		name = WeakFieldPrefix_goname + name
	}
	g.Annotate(m.GoIdent.GoName+"."+name, field.Location)
	leadingComments := appendDeprecationSuffix(field.Comments.Leading,
		field.Desc.Options().(*descriptorpb.FieldOptions).GetDeprecated())
	g.P(leadingComments,
		name, " ", goType, tags,
		trailingComment(field.Comments.Trailing))
	sf.append(field.GoName)
	return nil
}

func (plugin *Plugin) buildPackage(ast *descriptorpb.FileDescriptorProto, deps map[string]*descriptorpb.FileDescriptorProto, args *options.Option) (*generator.PackageDescription, error) {
	if ast == nil {
		return nil, fmt.Errorf("ast is nil")
	}

	pkg := getGoPackage(ast, map[string]string{})
	main := &model.Model{
		FilePath:    ast.GetName(),
		Package:     pkg,
		PackageName: util.BaseName(pkg, ""),
	}
	info := FileInfos{
		Official:  deps,
		PbReflect: nil,
	}
	resolver, err := NewResolver(ast, info, main, map[string]string{})
	if err != nil {
		return nil, fmt.Errorf("new protobuf resolver failed, err:%v", err)
	}
	err = resolver.LoadAll(ast)
	if err != nil {
		return nil, err
	}

	services, err := astToService(ast, resolver, args.CmdType, plugin.Plugin)
	if err != nil {
		return nil, err
	}
	var models model.Models
	for _, s := range services {
		models.MergeArray(s.DependencyModels)
	}

	return &generator.PackageDescription{
		Services: services,
		IdlName:  ast.GetName(),
		Package:  util.BaseName(pkg, ""),
		Models:   models,
	}, nil
}

func (plugin *Plugin) generatePackageFiles(ast *descriptorpb.FileDescriptorProto, deps map[string]*descriptorpb.FileDescriptorProto, args *options.Option) ([]util.File, error) {
	options := CheckTagOption(args)
	idl, err := plugin.buildPackage(ast, deps, args)
	if err != nil {
		return nil, err
	}

	customPackageTemplate := args.CustomizePackage
	pkg, err := args.GetGoPackage()
	if err != nil {
		return nil, err
	}
	modelDir, err := args.GetModelDir()
	if err != nil {
		return nil, err
	}
	clientDir, err := args.GetClientDir()
	if err != nil {
		return nil, err
	}

	sg := generator.HttpPackageGenerator{
		ServiceGroup: args.ServiceGroup,
		ConfigPath:   customPackageTemplate,
		ModelDir:     modelDir,
		UseDir:       args.Use,
		ClientDir:    clientDir,
		TemplateGenerator: tpl.TemplateGenerator{
			OutputDir: args.OutDir,
			Excludes:  args.Excludes,
		},
		ProjPackage:          pkg,
		Options:              options,
		CmdType:              args.CmdType,
		IdlClientDir:         plugin.IdlClientDir,
		ForceClientDir:       args.ForceClientDir,
		BaseDomain:           args.BaseDomain,
		QueryEnumAsInt:       args.QueryEnumAsInt,
		SnakeStyleMiddleware: args.SnakeStyleMiddleware,
		ForceUpdateClient:    args.ForceUpdateClient,
	}

	if args.ModelBackend != "" {
		sg.Backend = meta.Backend(args.ModelBackend)
	}
	generator.SetDefaultTemplateConfig()

	err = sg.GeneratePackage(idl)
	if err != nil {
		return nil, fmt.Errorf("generate http package error: %v", err)
	}

	files, err := sg.GetFormatAndExcludedFiles()
	if err != nil {
		return nil, fmt.Errorf("persist http package error: %v", err)
	}

	return files, nil
}
