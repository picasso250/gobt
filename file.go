package gobt

import (
	"io/ioutil"
	"os"
	"strings"
)

func ensureFile(info *MetainfoInfo) error {
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
