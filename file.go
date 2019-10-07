package gobt

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// File file
// todo should be int64
type File struct {
	Length int
	Path   []string
}

func (f *File) longPath() string {
	return pathBuild(f.Path...)
}

// NewFileFromMap builds a File
func NewFileFromMap(m map[string]interface{}) File {
	return File{
		Length: m["length"].(int),
		Path:   stringSlice(m["path"].([]interface{})),
	}
}

func ensureFile(info *MetainfoInfo) error {
	if err := os.Chdir(downloadRoot); err != nil {
		return err
	}

	if len(info.Files) != 0 {
		return ensureFiles(info)
	}
	return ensureOneFile(info)
}

func pathBuild(path ...string) string {
	return strings.Join(path, string([]rune([]rune{os.PathSeparator})))
}
func ensureFiles(info *MetainfoInfo) error {
	filename := info.filename()
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err := os.Mkdir(filename, 0664)
		if err != nil {
			return err
		}
	}

	if err := os.Chdir(filename); err != nil {
		return err
	}

	for _, file := range info.Files {
		path := file.Path
		prefix := string(append([]rune(filename), os.PathSeparator))
		err := ensureFileOneByPathList(prefix, path)
		if err != nil {
			return err
		}
	}

	infoFilename := info.infofilename()
	if _, err := os.Stat(infoFilename); os.IsNotExist(err) {
		piecesCount := len([]byte(info.Pieces)) / infoHashSize
		err := allZeroBitField(piecesCount).ToFile(infoFilename)
		if err != nil {
			return err
		}
	}
	return nil
}
func ensureFileOneByPathList(rootDir string, pathList []string) error {
	r := []rune(rootDir)
	for i, path := range pathList {
		if i == len(pathList)-1 {
			r = append(r, []rune(path)...)
			filename := string(r)
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				b := make([]byte, 0)
				err := ioutil.WriteFile(filename, b, 0664)
				if err != nil {
					return err
				}
			}
		} else {
			r = append(r, []rune(path)...)
			dir := string(r)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				err := os.Mkdir(dir, 0664)
				if err != nil {
					return err
				}
			}
			r = append(r, os.PathSeparator)
		}
	}
	return nil
}
func ensureOneFile(info *MetainfoInfo) error {
	filename := info.filename()
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		b := make([]byte, 0)
		err := ioutil.WriteFile(filename, b, 0664)
		if err != nil {
			return err
		}
	}

	infoFilename := info.infofilename()
	if _, err := os.Stat(infoFilename); os.IsNotExist(err) {
		piecesCount := len([]byte(info.Pieces)) / infoHashSize
		err := allZeroBitField(piecesCount).ToFile(infoFilename)
		if err != nil {
			return err
		}
	}
	return nil
}
func checkHash(info *MetainfoInfo, index int) bool {
	if len(info.Files) == 0 {
		// single file mode
		b := make([]byte, 0, info.PieceLength)
		file, err := os.Open(info.filename()) // For read access.
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		_, err = io.ReadFull(file, b)
		if err != nil {
			if err == io.ErrUnexpectedEOF {
				// it's ok
			} else {
				log.Fatal(err)
			}
		}
	} else {
		// multi file mode
		// seek to start
		i := index
		for _, f := range info.Files {
			if index > 0 {
				if index >= f.Length {
					continue
				}
				os.Fopen(f.longPath())
			}
		}
	}
	data := []byte("This page intentionally left blank.")
	fmt.Printf("% x", sha1.Sum(data))
}
