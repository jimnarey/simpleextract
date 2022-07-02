package simpleextract

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gen2brain/go-unarr"
	"github.com/mholt/archiver/v4"
)

const FIXTURES_PATH_ string = "./fixtures"
const OUT_PATH_ string = "./fixtures/out"

var ARCHIVE_GETTERS = []func(string) (archive, error){newArchiverArchive, newUnarrArchive}

// func panicOnErr(err error) {
// 	if err != nil {
// 		panic(err)
// 	}
// }

func mkdir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	return nil
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), err
}

type archive interface {
	extractAllTo(string) ([]string, error)
	basename() string
}

type fileBase struct {
	path string
}

func (f fileBase) basename() string {
	basename := path.Base(f.path)
	return strings.TrimSuffix(basename, path.Ext(basename))
}

type unarrArchive struct {
	fileBase
}

func (a unarrArchive) extractAllTo(targetPath string) ([]string, error) {
	arc, err := unarr.NewArchive(a.fileBase.path)
	if err != nil {
		return nil, err
	}
	defer arc.Close()
	err = mkdir(targetPath)
	if err != nil {
		return nil, err
	}
	contents, err := arc.Extract(targetPath)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

func (a unarrArchive) basename() string {
	return a.fileBase.basename()
}

type archiverArchive struct {
	fileBase
}

func (a archiverArchive) extractAllTo(targetPath string) ([]string, error) {
	f, err := os.Open(a.fileBase.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	format, input, err := archiver.Identify(a.fileBase.path, f)
	if err != nil {
		return nil, err
	}
	err = mkdir(targetPath)
	if err != nil {
		return nil, err
	}
	handler := func(ctx context.Context, f archiver.File) error {
		// fmt.Println("Extractor instantiated")
		// fmt.Println(f.FileInfo.Name())
		// fmt.Println(f.NameInArchive)
		// x, err := f.Open()
		if err != nil {
			// fmt.Println(err)
			return err
		}
		x, err := f.Open()
		if err != nil {
			// fmt.Println(err)
			return err
		}
		// x.Read()
		var y []byte
		x.Read(y)
		fmt.Println(x)
		err = ioutil.WriteFile(path.Join(targetPath, f.NameInArchive), y, 0755)
		if err != nil {
			// fmt.Println(err)
			return err
		}
		return nil
	}
	// want to extract something?
	if ex, ok := format.(archiver.Extractor); ok {
		err := ex.Extract(context.TODO(), input, nil, handler)
		if err != nil {
			return nil, err
		}
	}

	// or maybe it's compressed and you want to decompress it?
	if decom, ok := format.(archiver.Decompressor); ok {
		rc, err := decom.OpenReader(input)
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		// fmt.Println("Decompressor instantiated")
		// read from rc to get decompressed data
	}
	return nil, nil
}

func (a archiverArchive) basename() string {
	return a.fileBase.basename()
}

func newUnarrArchive(archivePath string) (archive, error) {
	arc, err := unarr.NewArchive(archivePath)
	if err != nil {
		return nil, err
	}
	defer arc.Close()
	return unarrArchive{fileBase: fileBase{path: archivePath}}, nil
}

func newArchiverArchive(archivePath string) (archive, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	_, _, err = archiver.Identify(archivePath, f)
	if err != nil {
		return nil, err
	}
	return archiverArchive{fileBase: fileBase{path: archivePath}}, nil
}

func GetArchive(archivePath string, archiveGetters []func(string) (archive, error)) (archive, error) {
	for i := 0; i < len(archiveGetters); i++ {
		archive, err := archiveGetters[i](archivePath)
		if err == nil {
			return archive, nil
		}

		// fmt.Println(err)
	}
	return nil, errors.New("no compatible unarchiver found")
}

func getTargetSubDir(targetPath string, arc archive) (string, error) {

	for i := 0; i < 100; i++ {
		subDir := path.Join(targetPath, arc.basename()+"_"+strconv.Itoa(i))
		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			err = mkdir(subDir)
			if err != nil {
				return "", err
			}
			return subDir, nil
		}
	}
	return "", errors.New("no subdir name")

}

func ExtractArchive(archivePath string, targetPath string) error {
	arc, err := GetArchive(path.Join(archivePath), ARCHIVE_GETTERS)
	if err != nil {
		return err
	}
	arc.extractAllTo(targetPath)
	return nil
}

func findByExt(searchPath string, fileExt string) ([]string, error) {
	tars := []string{}
	err := filepath.Walk(searchPath, func(path_ string, info os.FileInfo, err error) error {
		if err == nil && filepath.Ext(info.Name()) == fileExt {
			tars = append(tars, path.Join(path_, info.Name()))
		}
		return nil
	})
	if err != nil {
		return []string{}, err
	}
	return tars, nil
}
