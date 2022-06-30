package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gen2brain/go-unarr"
	"github.com/mholt/archiver"
)

const FIXTURES_PATH string = "./fixtures"

var ARCHIVE_GETTERS = []func(string) (archive, error){newArchiverArchive, newUnarrArchive}

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
	extractAllTo(string) error
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

func (a unarrArchive) extractAllTo(targetPath string) error {
	arc, err := unarr.NewArchive(a.fileBase.path)
	if err != nil {
		return err
	}
	defer arc.Close()
	_, err = arc.Extract(targetPath)
	if err != nil {
		return err
	}
	return nil
}

func (a unarrArchive) basename() string {
	return a.fileBase.basename()
}

type archiverArchive struct {
	fileBase
}

func (a archiverArchive) extractAllTo(targetPath string) error {
	err := archiver.Unarchive(a.fileBase.path, targetPath)
	if err == nil {
		return nil
	}
	fmt.Println(err)
	err = mkdir(targetPath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	tarPath := path.Join(targetPath, a.fileBase.basename()+".tar")
	err = archiver.DecompressFile(a.fileBase.path, tarPath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (a archiverArchive) basename() string {
	return a.fileBase.basename()
}

func newUnarrArchive(path string) (archive, error) {
	arc, err := unarr.NewArchive(path)
	if err != nil {
		return nil, err
	}
	defer arc.Close()
	return unarrArchive{fileBase: fileBase{path: path}}, nil
}

func newArchiverArchive(path string) (archive, error) {
	_, err := archiver.ByExtension(path)
	if err != nil {
		return nil, err
	}
	return archiverArchive{fileBase: fileBase{path: path}}, nil
}

func getArchive(path string, archiveGetters []func(string) (archive, error)) (archive, error) {
	for i := 0; i < len(archiveGetters); i++ {
		archive, err := archiveGetters[i](path)
		if err == nil {
			return archive, nil
		}

		fmt.Println(err)
	}
	return nil, errors.New("No compatible unarchiver found")
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
	return "", errors.New("No subdir name")

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

	rootPath := FIXTURES_PATH

	archives, err := ioutil.ReadDir(rootPath)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(archives); i++ {
		fullPath := path.Join(rootPath, archives[i].Name())
		isDir, _ := isDirectory(fullPath)
		if !isDir {
			targetDir := path.Join(rootPath, "out", path.Base(archives[i].Name())+"-"+strconv.Itoa(i))
			fmt.Println(targetDir)
			err := ExtractArchive(fullPath, targetDir)
			if err != nil {
				fmt.Println("**Extract Fail**")
				fmt.Println(err)
			}
		}

	}

}
