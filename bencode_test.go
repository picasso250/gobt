package gobt

import (
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {
	str := "spam"
	b, err := Encode(str)
	if err != nil {
		t.Errorf("spam encode error")
	}
	if string(b) != "4:spam" {
		t.Errorf("spam encode fail")
	}

	b, err = Encode(3)
	if err != nil {
		t.Errorf("3 encode error")
	}
	if string(b) != "i3e" {
		t.Errorf("3 encode fail")
	}

	b, err = Encode(int64(-3))
	if err != nil {
		t.Errorf("-3 encode error")
	}
	if string(b) != "i-3e" {
		t.Errorf("-3 encode fail")
	}

	lst := []string{"spam", "eggs"}
	b, err = Encode(lst)
	if err != nil {
		fmt.Println(err)
		t.Errorf("list encode error")
	}
	if string(b) != "l4:spam4:eggse" {
		t.Errorf("list encode fail")
	}

	dictionary := map[string]interface{}{"cow": "moo", "spam": "eggs"}
	b, err = Encode(dictionary)
	if err != nil {
		t.Errorf("dictionary encode error")
	}
	if string(b) != "d3:cow3:moo4:spam4:eggse" {
		t.Errorf("dictionary encode fail")
	}
	dictionary = map[string]interface{}{"spam": []string{"a", "b"}}
	b, err = Encode(dictionary)
	if err != nil {
		t.Errorf("dictionary encode error")
	}
	if string(b) != "d4:spaml1:a1:bee" {
		t.Errorf("dictionary encode fail")
	}
}
func TestEncodeBytes(t *testing.T) {
	str := []byte("spam")
	b, err := Encode(str)
	if err != nil {
		t.Errorf("spam encode error")
	}
	if string(b) != "4:spam" {
		t.Errorf("spam encode fail")
	}
}

func TestParse(t *testing.T) {
	var v interface{}
	var err error

	v, err = Parse([]byte("4:test"))
	if err != nil {
		t.Errorf("string test error")
	}
	if string(v.([]byte)) != "test" {
		t.Errorf("string test fail")
	}

	v, err = Parse([]byte("i1234e"))
	if err != nil {
		t.Errorf("integer 1234 error")
	}
	if v.(int64) != 1234 {
		t.Errorf("integer 1234 fail")
	}

	v, err = Parse([]byte("i-1234e"))
	if err != nil {
		t.Errorf("integer -1234 error")
	}
	if v.(int64) != -1234 {
		t.Errorf("integer -1234 fail")
	}

	v, err = Parse([]byte("i0e"))
	if err != nil {
		t.Errorf("integer 0 error")
	}
	if v.(int64) != 0 {
		t.Errorf("integer 0 fail")
	}

	var rl []interface{}
	rl = []interface{}{[]byte("test"), []byte("abcde")}
	v, err = Parse([]byte("l4:test5:abcdee"))
	if err != nil {
		t.Errorf("list error")
	}
	if !reflect.DeepEqual(v.([]interface{}), rl) {
		t.Errorf("list fail")
	}

	var rd map[string]interface{}
	rd = map[string]interface{}{"age": int64(20)}
	v, err = Parse([]byte("d3:agei20ee"))
	if err != nil {
		t.Errorf("dictionary age error")
	}
	if !reflect.DeepEqual(v, rd) {
		t.Errorf("dictionary age fail")
	}

	rd = map[string]interface{}{
		"path":     []byte("C:\\"),
		"filename": []byte("test.txt"),
	}
	v, err = Parse([]byte("d4:path3:C:\\8:filename8:test.txte"))
	if err != nil {
		t.Errorf("dictionary path error")
	}
	if !reflect.DeepEqual(v, rd) {
		t.Errorf("dictionary path fail")
	}

}

func TestRealFile(t *testing.T) {
	dat, err := ioutil.ReadFile("b.torrent")
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
	if info["pieces"] == nil {
		t.Errorf("info no pieces key")
	}
	if info["length"] == nil && info["files"] == nil {
		t.Errorf("info no length and files")
	}
	PrintMetainfo(v)
}
