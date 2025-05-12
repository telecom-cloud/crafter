# 命令行指南

本文档提供了Crafter中常用命令行工具的使用说明。

## 基本命令

### 命令行参数描述

```shell
NAME:
cft - A idl parser and code generator for microservice projects

USAGE:
cft [global options] command [command options]

VERSION:
v0.0.1

COMMANDS:
new      Generate a new Crafter project
update   Update an existing Crafter project
model    Generate model code only
client   Generate crafter client based on IDL
error    Generate error code only
doc      Generate doc only
help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
--verbose      turn on verbose mode (default: false)
--help, -h     show help
--version, -v  print the version
```

### 查看版本信息

```shell
bash cft --version
```

## 高级用法

### 创建项目

执行以下命令来创建一个新的Crafter项目：

```shell
bash cft new my_project
```

### 更新项目

```shell
bash cft update my_project
```

### 生成客户端

```shell
NAME:
   cft client - Generate crafter client based on IDL

USAGE:
   cft client [command options]

OPTIONS:
   --service_group value                                              specify the service group
   --idl value [ --idl value ]                                        Specify the IDL file path. (.proto)
   --module value, --mod value                                        Specify the Go module name.
   --base_domain value                                                Specify the request domain.
   --model_dir value                                                  Specify the model relative path (based on "out_dir").
   --client_dir value                                                 Specify the client path. If not specified, IDL generated path is used for 'client' command; no client code is generated for 'new' command
   --force_client_dir value                                           Specify the client path, and won't use namespaces as subpaths
   --force_client                                                     Force update 'crafter_client.go' (default: false)
   --proto_path value, -I value [ --proto_path value, -I value ]      Add an IDL search path for includes. (Valid only if idl is protobuf)
   --protoc value, -p value [ --protoc value, -p value ]              Specify arguments for the protoc. ({flag}={value})
   --no_recurse                                                       Generate master model only. (default: false)
   --trim_gopackage value, --trim_pkg value                           Trim the prefix of go_package for protobuf.
   --query_enumint                                                    Use num instead of string for query enum parameter. (default: false)
   --unset_omitempty                                                  Remove 'omitempty' tag for generated struct. (default: false)
   --pb_camel_json_tag                                                Convert Name style for json tag to camel(Only works protobuf). (default: false)
   --snake_tag                                                        Use snake_case style naming for tags. (Only works for 'form', 'query', 'json') (default: false)
   --rm_tag value [ --rm_tag value ]                                  Remove the default tag(json/query/form). If the annotation tag is set explicitly, it will not be removed.
   --exclude_file value, -E value [ --exclude_file value, -E value ]  Specify the files that do not need to be updated.
   --customize_package value                                          Specify the path for package template.
   --protoc-plugins value [ --protoc-plugins value ]                  Specify plugins for the protoc. ({plugin_name}:{options}:{out_dir})
   --help, -h                                                         show help
```

### 生成基础模型

```shell
NAME:
   cft model - Generate model code only

USAGE:
   cft model [command options]

OPTIONS:
   --idl value [ --idl value ]                                        Specify the IDL file path. (.proto)
   --module value, --mod value                                        Specify the Go module name.
   --out_dir value                                                    Specify the project path.
   --model_dir value                                                  Specify the model relative path (based on "out_dir").
   --proto_path value, -I value [ --proto_path value, -I value ]      Add an IDL search path for includes. (Valid only if idl is protobuf)
   --protoc value, -p value [ --protoc value, -p value ]              Specify arguments for the protoc. ({flag}={value})
   --no_recurse                                                       Generate master model only. (default: false)
   --trim_gopackage value, --trim_pkg value                           Trim the prefix of go_package for protobuf.
   --unset_omitempty                                                  Remove 'omitempty' tag for generated struct. (default: false)
   --pb_camel_json_tag                                                Convert Name style for json tag to camel(Only works protobuf). (default: false)
   --snake_tag                                                        Use snake_case style naming for tags. (Only works for 'form', 'query', 'json') (default: false)
   --rm_tag value [ --rm_tag value ]                                  Remove the default tag(json/query/form). If the annotation tag is set explicitly, it will not be removed.
   --exclude_file value, -E value [ --exclude_file value, -E value ]  Specify the files that do not need to be updated.
   --help, -h                                                         show help
```

### 生成错误码

```shell
NAME:
   cft error - Generate error code only

USAGE:
   cft error [command options]

OPTIONS:
   --idl value [ --idl value ]                                        Specify the IDL file path. (.proto)
   --module value, --mod value                                        Specify the Go module name.
   --out_dir value                                                    Specify the project path.
   --model_dir value                                                  Specify the model relative path (based on "out_dir").
   --proto_path value, -I value [ --proto_path value, -I value ]      Add an IDL search path for includes. (Valid only if idl is protobuf)
   --protoc value, -p value [ --protoc value, -p value ]              Specify arguments for the protoc. ({flag}={value})
   --no_recurse                                                       Generate master model only. (default: false)
   --trim_gopackage value, --trim_pkg value                           Trim the prefix of go_package for protobuf.
   --unset_omitempty                                                  Remove 'omitempty' tag for generated struct. (default: false)
   --pb_camel_json_tag                                                Convert Name style for json tag to camel(Only works protobuf). (default: false)
   --snake_tag                                                        Use snake_case style naming for tags. (Only works for 'form', 'query', 'json') (default: false)
   --rm_tag value [ --rm_tag value ]                                  Remove the default tag(json/query/form). If the annotation tag is set explicitly, it will not be removed.
   --exclude_file value, -E value [ --exclude_file value, -E value ]  Specify the files that do not need to be updated.
   --help, -h                                                         show help
```

### 生成帮助文档

```shell
NAME:
   cft doc - Generate doc only

USAGE:
   cft doc [command options]

OPTIONS:
   --idl value [ --idl value ]                                        Specify the IDL file path. (.proto)
   --module value, --mod value                                        Specify the Go module name.
   --out_dir value                                                    Specify the project path.
   --model_dir value                                                  Specify the model relative path (based on "out_dir").
   --proto_path value, -I value [ --proto_path value, -I value ]      Add an IDL search path for includes. (Valid only if idl is protobuf)
   --protoc value, -p value [ --protoc value, -p value ]              Specify arguments for the protoc. ({flag}={value})
   --no_recurse                                                       Generate master model only. (default: false)
   --trim_gopackage value, --trim_pkg value                           Trim the prefix of go_package for protobuf.
   --unset_omitempty                                                  Remove 'omitempty' tag for generated struct. (default: false)
   --pb_camel_json_tag                                                Convert Name style for json tag to camel(Only works protobuf). (default: false)
   --snake_tag                                                        Use snake_case style naming for tags. (Only works for 'form', 'query', 'json') (default: false)
   --rm_tag value [ --rm_tag value ]                                  Remove the default tag(json/query/form). If the annotation tag is set explicitly, it will not be removed.
   --exclude_file value, -E value [ --exclude_file value, -E value ]  Specify the files that do not need to be updated.
   --help, -h                                                         show help
```