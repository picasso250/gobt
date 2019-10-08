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
type File struct {
	Length int64
	Path   []string
}

func (f *File) longPath() string {
	return pathBuild(f.Path...)
}

// NewFileFromMap builds a File
func NewFileFromMap(m map[string]interface{}) File {
	return File{
		Length: m["length"].(int64),
		Path:   stringSlice(m["path"].([]interface{})),
	}
}

func writeToFile(info *MetainfoInfo, index int, offset int64, piece []byte) error {
	offset = int64(index)*int64(info.PieceLength) + offset
	if len(info.Files) != 0 {
		return writeToFiles(info, offset, piece)
	}
	return writeToOneFile(info, offset, piece)
}

func writeToOneFile(info *MetainfoInfo, offset int64, piece []byte) error {
	filename := info.filename()
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return err
	}
	_, err = f.Seek(offset, 0)
	if err != nil {
		return err
	}

	return writeAll(f, piece)
}
func writeAll(w io.Writer, b []byte) error {
	for {
		n, err := w.Write(b)
		if err != nil {
			return err
		}
		if n == len(b) {
			break
		}
		b = b[n:]
	}
	return nil
}
func writeToFiles(info *MetainfoInfo, offset int64, piece []byte) error {
	i, offset, err := seekStart(info.Files, offset)
	if err != nil {
		return err
	}
	return writeToFilesDo(info.Files[i:], offset, piece)
}
func writeToFilesDo(files []File, offset int64, piece []byte) error {
	if len(piece) == 0 {
		return nil
	}
	for _, file := range files {
		f, err := os.OpenFile(file.longPath(), os.O_WRONLY|os.O_CREATE, 0664)
		if err != nil {
			return err
		}
		defer f.Close()
		if offset != 0 {
			_, err = f.Seek(offset, 0)
			if err != nil {
				return err
			}
		}

		b := piece
		if offset+int64(len(piece)) > file.Length {
			// 0----offset----file.Length----len(piece)
			b = piece[:file.Length-offset]
		}
		err = writeAll(f, b)
		if err != nil {
			return err
		}
		piece = piece[len(b):]
		if len(piece) == 0 {
			break
		}
	}
	return nil
}

func ensureFile(info *MetainfoInfo) (bitfield, error) {
	if err := os.Chdir(downloadRoot); err != nil {
		return nil, err
	}

	if len(info.Files) != 0 {
		return ensureFiles(info)
	}
	return ensureOneFile(info)
}

func pathBuild(path ...string) string {
	return strings.Join(path, string([]rune([]rune{os.PathSeparator})))
}
func ensureFiles(info *MetainfoInfo) (bf bitfield, err error) {
	filename := info.filename()
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err := os.Mkdir(filename, 0664)
		if err != nil {
			return nil, err
		}
	}

	if err := os.Chdir(filename); err != nil {
		return nil, err
	}

	for _, file := range info.Files {
		path := file.Path
		prefix := string(append([]rune(filename), os.PathSeparator))
		err := ensureFileOneByPathList(prefix, path)
		if err != nil {
			return nil, err
		}
	}

	return ensureInfoFile(info.infofilename(), info.piecesCount())
}
func ensureInfoFile(infoFilename string, piecesCount int) (bf bitfield, err error) {
	if _, err := os.Stat(infoFilename); os.IsNotExist(err) {
		bf = allZeroBitField(piecesCount)
		err := bf.ToFile(infoFilename)
		if err != nil {
			return nil, err
		}
	} else {
		return ioutil.ReadFile(infoFilename)
	}
	return
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
func ensureOneFile(info *MetainfoInfo) (bf bitfield, err error) {
	filename := info.filename()
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		b := make([]byte, 0)
		err := ioutil.WriteFile(filename, b, 0664)
		if err != nil {
			return nil, err
		}
	}

	return ensureInfoFile(info.infofilename(), info.piecesCount())
}
func checkHash(info *MetainfoInfo, index int, ih hash) (flag bool, err error) {
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
		b, err = readSomeFileContent(info, index, 0, int64(info.PieceLength))
		if err != nil {
			return false, err
		}
	}
	s := sha1.Sum(b)
	return bytes.Compare(s[:], ih[:]) == 0, nil
}
func readSomeFileContent(info *MetainfoInfo, index int, offset int64, length int64) ([]byte, error) {
	fileIndex, offset, err := seekStart(info.Files, int64(index)*int64(info.PieceLength)+offset)
	if err != nil {
		return nil, err
	}
	return readMuliFileBlock(info.Files[fileIndex:], offset, length)
}
func readPieceLength(info *MetainfoInfo, index int, offset int64) int64 {
	length := int64(info.PieceLength)
	sum := int64(0)
	for _, file := range info.Files[index:] {
		sum += file.Length
	}
	sum -= offset
	if sum < length {
		length = sum
	}
	return length
}
func readMuliFileBlock(fileList []File, offset int64, length int64) ([]byte, error) {
	if length <= 0 {
		log.Fatalf("invalid length %d", length)
	}
	b := make([]byte, length)
	bufStart := int64(0)
	for _, file := range fileList {
		f, err := os.Open(file.longPath())
		if err != nil {
			return nil, err
		}
		defer f.Close()
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
		bufStart += int64(n)
		if bufStart == length-1 {
			break
		}
	}
	if bufStart != length-1 {
		return nil, errors.New("not enough data read")
	}
	return b, nil
}
func seekStart(files []File, index int64) (i int, pos int64, err error) {
	for i, f := range files {
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
