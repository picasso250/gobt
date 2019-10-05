package gobt

import (
	"fmt"
	"io/ioutil"
	"log"
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

func TestRealFile(t *testing.T) {
	dat, err := ioutil.ReadFile("a.torrent")
	if err != nil {
		log.Fatal(err)
	}
	vv, err := Parse(dat)
	if err != nil {
		t.Errorf("file parse error")
	}
	v := vv.(map[string]interface{})
	if v["announce"] == nil {
		t.Errorf("no announce key")
	}
	if v["info"] == nil {
		t.Errorf("no info key")
	}
	info := v["info"].(map[string]interface{})
	if info["piece length"] == nil {
		t.Errorf("info no piece length key")
	}
	fmt.Println(info["piece length"])
	if info["pieces"] == nil {
		t.Errorf("info no pieces key")
	}
	fmt.Println(len(info["pieces"].(string)))
	if info["length"] == nil && info["files"] == nil {
		t.Errorf("info no length and files")
	}
	fmt.Println(info["length"])
	fmt.Println(info["files"])
}
