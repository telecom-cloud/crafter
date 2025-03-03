package util

import (
	"reflect"
	"strings"
	"unicode/utf8"
	"unsafe"
)

func Str2Bytes(in string) (out []byte) {
	op := (*reflect.SliceHeader)(unsafe.Pointer(&out))
	ip := (*reflect.StringHeader)(unsafe.Pointer(&in))
	op.Data = ip.Data
	op.Cap = ip.Len
	op.Len = ip.Len
	return
}

func Bytes2Str(in []byte) (out string) {
	op := (*reflect.SliceHeader)(unsafe.Pointer(&out))
	ip := (*reflect.SliceHeader)(unsafe.Pointer(&in))
	op.Data = ip.Data
	op.Len = ip.Len
	return
}

// TrimLastChar can remove the last char for s
func TrimLastChar(s string) string {
	r, size := utf8.DecodeLastRuneInString(s)
	if r == utf8.RuneError && (size == 0 || size == 1) {
		size = 0
	}
	return s[:len(s)-size]
}

// TrimSuffix can remove the service suffix for s
func TrimSuffix(s string) string {
	return strings.TrimSuffix(s, "Service")
}

// AddSlashForComments can adjust the format of multi-line comments
func AddSlashForComments(s string) string {
	s = strings.Replace(s, "\n", "\n//", -1)
	return s
}

// CamelString converts the string 's' to a camel string
func CamelString(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || k == false) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return Bytes2Str(data[:])
}

// SnakeString converts the string 's' to a snake string
func SnakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	for _, d := range Str2Bytes(s) {
		if d >= 'A' && d <= 'Z' {
			if j {
				data = append(data, '_')
				j = false
			}
		} else if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(Bytes2Str(data))
}
