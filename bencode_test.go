package gobt

import (
	// "fmt"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	var v interface{}
	var err error

	v, err = Parse([]byte("4:test"))
	if err != nil {
		t.Errorf("string test error")
	}
	if v != "test" {
		t.Errorf("string test fail")
	}

	v, err = Parse([]byte("i1234e"))
	if err != nil {
		t.Errorf("integer 1234 error")
	}
	if v != 1234 {
		t.Errorf("integer 1234 fail")
	}

	v, err = Parse([]byte("i-1234e"))
	if err != nil {
		t.Errorf("integer -1234 error")
	}
	if v != -1234 {
		t.Errorf("integer -1234 fail")
	}

	v, err = Parse([]byte("i0e"))
	if err != nil {
		t.Errorf("integer 0 error")
	}
	if v != 0 {
		t.Errorf("integer 0 fail")
	}

	var rl []interface{}
	rl = []interface{}{"test", "abcde"}
	v, err = Parse([]byte("l4:test5:abcdee"))
	if err != nil {
		t.Errorf("list error")
	}
	if !Equal(v.([]interface{}), rl) {
		t.Errorf("list fail")
	}

	var rd map[string]interface{}
	rd = map[string]interface{}{"age": 20}
	v, err = Parse([]byte("d3:agei20ee"))
	if err != nil {
		t.Errorf("dictionary age error")
	}
	if !reflect.DeepEqual(v, rd) {
		t.Errorf("dictionary age fail")
	}

	rd = map[string]interface{}{"path": "C:\\", "filename": "test.txt"}
	v, err = Parse([]byte("d4:path3:C:\\8:filename8:test.txte"))
	if err != nil {
		t.Errorf("dictionary path error")
	}
	if !reflect.DeepEqual(v, rd) {
		t.Errorf("dictionary path fail")
	}
}
