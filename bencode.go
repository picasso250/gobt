package gobt

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"unicode"
)

// PrintMetainfo print the metainfo
func PrintMetainfo(m map[string]interface{}) {
	printValue(m, "pieces")
}
func printValue(v interface{}, bytesKey ...string) {
	v = prepareValueIter(v, bytesKey...)
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	e.Encode(v)
}
func prepareValueIter(v interface{}, bytesKey ...string) interface{} {
	m := make(map[string]bool, len(bytesKey))
	for _, k := range bytesKey {
		m[k] = true
	}
	switch v := v.(type) {
	default:
		return v
	case []byte:
		return string(v)
	case int64, int:
		return v
	case []interface{}:
		ret := make([]interface{}, len(v))
		for i, v := range v {
			ret[i] = prepareValueIter(v, bytesKey...)
		}
		return ret
	case map[string]interface{}:
		ret := make(map[string]interface{}, len(v))
		for k, v := range v {
			if m[k] {
				ret[k] = toHex(v.([]byte))
			} else {
				ret[k] = prepareValueIter(v, bytesKey...)
			}
		}
		return ret
	}
}
func toHex(src []byte) string {
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)
	return string(dst)
}
func printStringln(b []byte) {
	if isPrint(b) {
		fmt.Println(string(b))
	} else {
		fmt.Println(b)
	}
}
func printString(b []byte) {
	if isPrint(b) {
		fmt.Print(string(b))
	} else {
		fmt.Print(b)
	}
}
func toStringSlice(v []interface{}) []string {
	ret := make([]string, len(v))
	for i, v := range v {
		ret[i] = string(v.([]byte))
	}
	return ret
}
func isPrint(b []byte) bool {
	for _, c := range []rune(string(b)) {
		if !unicode.IsPrint(c) {
			return false
		}
	}
	return true
}

// Encode bencoding
func Encode(v interface{}) ([]byte, error) {
	switch v := v.(type) {
	default:
		val := reflect.ValueOf(v)
		switch val.Kind() {
		default:
			return nil, errors.New("unknown type")
		case reflect.Slice:
			return encodeSlice(val)
		case reflect.Map:
			return encodeMap(val)
		}
	case string:
		length := len(v)
		lenStr := strconv.Itoa(length)
		return []byte(lenStr + ":" + v), nil
	case []byte:
		length := len(v)
		lenStr := strconv.Itoa(length)
		b := bytes.Join([][]byte{[]byte(lenStr), v}, []byte{':'})
		return b, nil
	case int:
		str := strconv.Itoa(v)
		return []byte("i" + str + "e"), nil
	case int64:
		str := strconv.FormatInt(v, 10)
		return []byte("i" + str + "e"), nil
	}
}

func encodeSlice(v reflect.Value) ([]byte, error) {
	b := make([]byte, 0, v.Len())
	b = append(b, 'l')
	for i := 0; i < v.Len(); i++ {
		str, err := Encode(v.Index(i).Interface())
		if err != nil {
			return nil, err
		}
		b = append(b, []byte(str)...)
	}
	b = append(b, 'e')
	return b, nil
}
func encodeMap(v reflect.Value) ([]byte, error) {
	b := make([]byte, 0, v.Len())
	b = append(b, 'd')
	m := sortMapValueByKey(v)
	for k, v := range m {
		kstr, err := Encode(k)
		if err != nil {
			return nil, err
		}
		b = append(b, []byte(kstr)...)
		str, err := Encode(v)
		if err != nil {
			return nil, err
		}
		b = append(b, []byte(str)...)
	}
	b = append(b, 'e')
	return b, nil
}
func sortMapValueByKey(m reflect.Value) (ret map[string]interface{}) {

	ks := m.MapKeys()
	keys := make([]string, 0, m.Len())
	for _, k := range ks {
		keys = append(keys, k.String())
	}
	sort.Strings(keys)

	ret = map[string]interface{}{}
	for _, k := range keys {
		ret[k] = m.MapIndex(reflect.ValueOf(k)).Interface()
	}
	return
}

// Parse 解码 bencode
func Parse(b []byte) (interface{}, error) {
	if len(b) == 0 {
		return nil, errors.New("empty bencode string")
	}
	v, left, err := parseValue(b)
	if err != nil {
		return v, err
	}
	if len(left) == 0 {
		return v, nil
	}
	return v, errors.New("still left")
}

func parseValue(b []byte) (interface{}, []byte, error) {
	if len(b) == 0 {
		return nil, nil, errors.New("empty value")
	}
	switch {
	case '0' <= b[0] && b[0] <= '9':
		s, left, err := parseString(b)
		if err != nil {
			return nil, left, err
		}
		return s, left, nil
	case b[0] == 'i':
		i, left, err := parseInt(b)
		if err != nil {
			return nil, left, err
		}
		return i, left, nil
	case b[0] == 'l':
		lst := make([]interface{}, 0)
		if len(b) == 1 {
			return lst, nil, errors.New("list after l no character")
		}
		if b[1] == 'e' && len(b) >= 2 {
			return lst, b[2:], nil
		}
		bt := b[1:]
		for {
			v, left, err := parseValue(bt)
			if err != nil {
				return lst, nil, err
			}
			bt = left
			lst = append(lst, v)
			if len(left) >= 1 && left[0] == 'e' {
				return lst, left[1:], nil
			}
		}
	case b[0] == 'd':
		m := make(map[string]interface{})
		if len(b) == 1 {
			return m, nil, errors.New("dictionary after d no character")
		}
		if b[1] == 'e' && len(b) >= 2 {
			return m, b[2:], nil
		}
		bt := b[1:]
		for {
			// key
			if len(bt) == 0 {
				return m, nil, errors.New("dictionary has no key")
			}
			if !('0' <= bt[0] && bt[0] <= '9') {
				return m, nil, errors.New("dictionary key is not string")
			}
			k, left, err := parseValue(bt)
			if err != nil {
				return m, nil, err
			}
			bt = left
			// value
			if len(left) >= 1 && left[0] == 'e' {
				return m, left[1:], errors.New("key has no value")
			}
			v, left, err := parseValue(bt)
			m[string(k.([]byte))] = v
			bt = left
			if len(left) >= 1 && left[0] == 'e' {
				return m, left[1:], nil
			}
		}
	}
	return nil, nil, errors.New("not any type")
}

func parseString(b []byte) ([]byte, []byte, error) {
	i := bytes.IndexByte(b, ':')
	if i == -1 {
		return nil, nil, errors.New("no : in string")
	}
	if i == 0 {
		return nil, nil, errors.New("length cannot be found in string")
	}
	len, err := strconv.Atoi(string(b[:i]))
	if err != nil {
		return nil, nil, err
	}
	b = b[i+1:]
	return (b[:len]), b[len:], nil
}
func parseInt(b []byte) (int64, []byte, error) {
	if b[0] != 'i' {
		return 0, nil, errors.New("integer not start with i")
	}
	idx := bytes.IndexByte(b, 'e')
	if idx == -1 {
		return 0, nil, errors.New("integer not end with e")
	}
	// todo Integers have no size limitation
	i, err := strconv.ParseInt(string(b[1:idx]), 10, 64)
	if err != nil {
		return 0, nil, err
	}
	return i, b[idx+1:], nil
}
