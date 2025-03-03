package model

import (
	"fmt"
	"strconv"
)

type BoolExpression struct {
	Src bool
}

func (boolExpr BoolExpression) Expression() string {
	if boolExpr.Src {
		return "true"
	} else {
		return "false"
	}
}

type StringExpression struct {
	Src string
}

func (stringExpr StringExpression) Expression() string {
	return fmt.Sprintf("%q", stringExpr.Src)
}

type NumberExpression struct {
	Src string
}

func (numExpr NumberExpression) Expression() string {
	return numExpr.Src
}

type ListExpression struct {
	ElementType *Type
	Elements    []Literal
}

type IntExpression struct {
	Src int
}

func (intExpr IntExpression) Expression() string {
	return strconv.Itoa(intExpr.Src)
}

type DoubleExpression struct {
	Src float64
}

func (doubleExpr DoubleExpression) Expression() string {
	return strconv.FormatFloat(doubleExpr.Src, 'f', -1, 64)
}

func (listExpr ListExpression) Expression() string {
	ret := "[]" + listExpr.ElementType.Name + "{\n"
	for _, e := range listExpr.Elements {
		ret += e.Expression() + ",\n"
	}
	ret += "\n}"
	return ret
}

type MapExpression struct {
	KeyType   *Type
	ValueType *Type
	Elements  map[string]Literal
}

func (mapExpr MapExpression) Expression() string {
	ret := "map[" + mapExpr.KeyType.Name + "]" + mapExpr.ValueType.Name + "{\n"
	for k, e := range mapExpr.Elements {
		ret += fmt.Sprintf("%q: %s,\n", k, e.Expression())
	}
	ret += "\n}"
	return ret
}
