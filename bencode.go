package gobt

import (
	"bytes"
	"errors"
	"strconv"
)

// Parse 解码 bencode
func Parse(b []byte) (interface{}, error) {
	if len(b) == 0 {
		return nil, errors.New("empty bencode string")
	}
	v, left, err := parseValue(b[1:])
	if err != nil {
		return v, err
	}
	if len(left) == 0 {
		return v, nil
	}
	return v, errors.New("still left")
}

func parseValue(b []byte) (interface{}, []byte, error) {
	switch {
	case '0' <= b[0] && b[0] <= '9':
		s, left, err := parseString(b)
		if err == nil {
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
		if b[1] == 'e' && len(b) == 2 {
			return lst, nil, nil
		}
		bt := b[1:]
		for {
			v, left, err := parseValue(bt)
			if err != nil {
				return lst, nil, err
			}
			bt = left
			lst = append(lst, v)
			if len(left) == 1 && left[1] == 'e' {
				return lst, left[1:], nil
			}
		}
	case b[0] == 'd':
		m := make(map[string]interface{})
		if len(b) == 1 {
			return m, nil, errors.New("dictionary after d no character")
		}
		if b[1] == 'e' && len(b) == 2 {
			return m, nil, nil
		}
		bt := b[1:]
		for {
			// key
			if len(bt) == 0 {
				return m, nil, errors.New("dictionary has no key")
			}
			if !('0' <= b[0] && b[0] <= '9') {
				return m, nil, errors.New("dictionary key is not string")
			}
			k, left, err := parseValue(bt)
			if err != nil {
				return m, nil, err
			}
			bt = left
			// value
			if len(left) == 1 && left[1] == 'e' {
				return m, left[1:], errors.New("key has no value")
			}
			v, left, err := parseValue(bt)
			m[k.(string)] = v
			bt = left
			if len(left) == 1 && left[1] == 'e' {
				return m, left[1:], nil
			}
		}
	}
	return nil, nil, errors.New("not any type")
}

func parseString(b []byte) (string, []byte, error) {
	i := bytes.IndexByte(b, ':')
	if i == -1 {
		return "", nil, errors.New("no : in string")
	}
	if i == 0 {
		return "", nil, errors.New("length cannot be found in string")
	}
	len, err := strconv.Atoi(string(b[:i-1]))
	if err != nil {
		return "", nil, err
	}
	b = b[i+1:]
	return string(b[:len]), b[len:], nil
}
func parseInt(b []byte) (int, []byte, error) {
	if b[0] != 'i' {
		return 0, nil, errors.New("integer not start with i")
	}
	idx := bytes.IndexByte(b, 'e')
	if idx == -1 {
		return 0, nil, errors.New("integer not end with e")
	}
	i, err := strconv.Atoi(string(b[1:idx]))
	if err != nil {
		return 0, nil, err
	}
	return i, b[idx+1:], nil
}
