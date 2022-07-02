package main

import (
	"errors"
	"fmt"
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

var testArchives_ = []string{"file.7z", "file.rar", "file.tar", "file.tar.7z", "file.tar.bz2", "file.tar.gz", "file.tar.xz", "file.zip"}

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
		fmt.Println("*archiver not compatible")
		return nil, err
	}
	return archiverArchive{fileBase: fileBase{path: archivePath}}, nil
}

func getArchive(archivePath string, archiveGetters []func(string) (archive, error)) (archive, error) {
	for i := 0; i < len(archiveGetters); i++ {
		archive, err := archiveGetters[i](archivePath)
		if err == nil {
			return archive, nil
		}

		fmt.Println(err)
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
	arc, err := getArchive(path.Join(archivePath), ARCHIVE_GETTERS)
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

func main() {

	for i := 0; i < len(testArchives_); i++ {
		arc, err := getArchive(path.Join(FIXTURES_PATH_, testArchives_[i]), ARCHIVE_GETTERS)
		if err != nil {

		}
		fmt.Println(arc)

	}

}
