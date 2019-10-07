package gobt

import (
	"bytes"
	"crypto/sha1"
	"errors"
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
		piecesCount := len([]byte(info.Pieces)) / hashSize
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
		piecesCount := len([]byte(info.Pieces)) / hashSize
		err := allZeroBitField(piecesCount).ToFile(infoFilename)
		if err != nil {
			return err
		}
	}
	return nil
}
func checkHash(info *MetainfoInfo, index int, ih hash) (bool, error) {
	b := make([]byte, 0, info.PieceLength)
	if len(info.Files) == 0 {
		// single file mode
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
		fileIndex, offset, err := seekStart(info, index)
		if err != nil {
			return false, err
		}
		length := readPieceLength(info, fileIndex, offset)
		b, err = readMuliFileBlock(info.Files[fileIndex:], int64(offset), length)
		if err != nil {
			return false, err
		}
	}
	s := sha1.Sum(b)
	return bytes.Compare(s[:], ih[:]) == 0, nil
}
func readPieceLength(info *MetainfoInfo, index int, offset int) int {
	length := info.PieceLength
	sum := 0
	for _, file := range info.Files[index:] {
		sum += file.Length
	}
	sum -= offset
	if sum < length {
		length = sum
	}
	return length
}
func readMuliFileBlock(fileList []File, offset int64, length int) ([]byte, error) {
	if length <= 0 {
		log.Fatalf("invalid length %d", length)
	}
	b := make([]byte, length)
	bufStart := 0
	for _, file := range fileList {
		f, err := os.Open(file.longPath())
		if err != nil {
			return nil, err
		}
		if offset != 0 {
			_, err = f.Seek(offset, 0)
			if err != nil {
				return nil, err
			}
		}
		n, err := io.ReadFull(f, b[bufStart:])
		if err != nil {
			if err != io.ErrUnexpectedEOF {
				return nil, err
			}
		}
		bufStart += n
		if bufStart == length-1 {
			break
		}
	}
	if bufStart != length-1 {
		return nil, errors.New("not enough data read")
	}
	return b, nil
}
func seekStart(info *MetainfoInfo, index int) (i int, pos int, err error) {
	for i, f := range info.Files {
		if index == 0 {
			return i, 0, nil
		}
		if index < f.Length {
			return i, index, nil
		}
		index -= f.Length
	}
	return 0, 0, errors.New("index out range")
}
