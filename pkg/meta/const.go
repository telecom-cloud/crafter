package meta

import "runtime"

// Version cft version
const Version = "v0.0.2"

const DefaultServiceName = "crafter_service"

// Mode cft run modes
type Mode int

// SysType is the running program's operating system type
const SysType = runtime.GOOS

const WindowsOS = "windows"

const EnvPluginMode = "CRAFTER_PLUGIN_MODE"

// cft Commands
const (
	CmdUpdate = "update"
	CmdNew    = "new"
	CmdModel  = "model"
	CmdClient = "client"
	CmdError  = "error"
	CmdDoc    = "doc"
)

const (
	IdlProto = "proto"
)

const (
	TpCompilerProto = "protoc"
)

// cft Plugins
const (
	ProtocPluginName = "protoc-gen-crafter"
	ThriftPluginName = "thrift-gen-crafter"
)

// cft Errors
const (
	LoadError           = 1
	GenerateLayoutError = 2
	PersistError        = 3
	PluginError         = 4
)

const (
	ModelDir = "biz/model"
)

// Backend Model Backends
type Backend string

const (
	BackendGolang Backend = "golang"
	BackendJava   Backend = "java"
	BackendPython Backend = "python"
	BackendRuby   Backend = "ruby"
	BackendRust   Backend = "rust"
)

const (
	SetBodyParam      = "SetBodyParam(req).\n"
	ContentTypeFormat = "\"Content-Type\": \"%s\","
)
