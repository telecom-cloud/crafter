package model

var (
	BaseTypes      = []*Type{TypeBool, TypeByte, TypeInt8, TypeInt16, TypeInt32, TypeInt64, TypeUint8, TypeUint16, TypeUint32, TypeUint64, TypeFloat64, TypeString, TypeBinary}
	ContainerTypes = []*Type{TypeBaseList, TypeBaseMap, TypeBaseSet}
	BaseModel      = Model{}
)

var (
	TypeBool = &Type{
		Name:  "bool",
		Scope: &BaseModel,
		Kind:  KindBool,
	}
	TypeByte = &Type{
		Name:  "int8",
		Scope: &BaseModel,
		Kind:  KindInt8,
	}
	TypePbByte = &Type{
		Name:  "byte",
		Scope: &BaseModel,
		Kind:  KindInt8,
	}
	TypeUint8 = &Type{
		Name:  "uint8",
		Scope: &BaseModel,
		Kind:  KindInt8,
	}
	TypeUint16 = &Type{
		Name:  "uint16",
		Scope: &BaseModel,
		Kind:  KindInt16,
	}
	TypeUint32 = &Type{
		Name:  "uint32",
		Scope: &BaseModel,
		Kind:  KindInt32,
	}
	TypeUint64 = &Type{
		Name:  "uint64",
		Scope: &BaseModel,
		Kind:  KindInt64,
	}
	TypeUint = &Type{
		Name:  "uint",
		Scope: &BaseModel,
		Kind:  KindInt,
	}
	TypeInt8 = &Type{
		Name:  "int8",
		Scope: &BaseModel,
		Kind:  KindInt8,
	}
	TypeInt16 = &Type{
		Name:  "int16",
		Scope: &BaseModel,
		Kind:  KindInt16,
	}
	TypeInt32 = &Type{
		Name:  "int32",
		Scope: &BaseModel,
		Kind:  KindInt32,
	}
	TypeInt64 = &Type{
		Name:  "int64",
		Scope: &BaseModel,
		Kind:  KindInt64,
	}
	TypeInt = &Type{
		Name:  "int",
		Scope: &BaseModel,
		Kind:  KindInt,
	}
	TypeFloat32 = &Type{
		Name:  "float32",
		Scope: &BaseModel,
		Kind:  KindFloat64,
	}
	TypeFloat64 = &Type{
		Name:  "float64",
		Scope: &BaseModel,
		Kind:  KindFloat64,
	}
	TypeString = &Type{
		Name:  "string",
		Scope: &BaseModel,
		Kind:  KindString,
	}
	TypeBinary = &Type{
		Name:     "binary",
		Scope:    &BaseModel,
		Kind:     KindSlice,
		Category: CategoryBinary,
		Extra:    []*Type{TypePbByte},
	}

	TypeBaseMap = &Type{
		Name:     "map",
		Scope:    &BaseModel,
		Kind:     KindMap,
		Category: CategoryMap,
	}
	TypeBaseSet = &Type{
		Name:     "set",
		Scope:    &BaseModel,
		Kind:     KindSlice,
		Category: CategorySet,
	}
	TypeBaseList = &Type{
		Name:     "list",
		Scope:    &BaseModel,
		Kind:     KindSlice,
		Category: CategoryList,
	}
)

func NewCategoryType(typ *Type, cg Category) *Type {
	cyp := *typ
	cyp.Category = cg
	return &cyp
}

func NewStructType(name string, cg Category) *Type {
	return &Type{
		Name:     name,
		Scope:    nil,
		Kind:     KindStruct,
		Category: cg,
		Indirect: false,
		Extra:    nil,
		HasNew:   true,
	}
}

func NewFuncType(name string, cg Category) *Type {
	return &Type{
		Name:     name,
		Scope:    nil,
		Kind:     KindFunc,
		Category: cg,
		Indirect: false,
		Extra:    nil,
		HasNew:   false,
	}
}

func IsBaseType(typ *Type) bool {
	for _, t := range BaseTypes {
		if typ == t {
			return true
		}
	}
	return false
}

func NewEnumType(name string, cg Category) *Type {
	return &Type{
		Name:     name,
		Scope:    &BaseModel,
		Kind:     KindInt,
		Category: cg,
		Indirect: false,
		Extra:    nil,
		HasNew:   true,
	}
}

func NewOneofType(name string) *Type {
	return &Type{
		Name:     name,
		Scope:    &BaseModel,
		Kind:     KindInterface,
		Indirect: false,
		Extra:    nil,
		HasNew:   true,
	}
}
