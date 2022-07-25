package simpleextract

import (
	"context"
	//"errors"
	"fmt"
	"io"
	"os"
	"path"

	// "strconv"
	//"strings"

	"github.com/gen2brain/go-unarr"
	"github.com/mholt/archiver/v4"
)

//const FIXTURES_PATH_ string = "./fixtures"
//const OUT_PATH_ string = "./fixtures/out"

var BUFFERSIZE = 1024
//var ARCHIVE_GETTERS = []func(string) (archive, error){newArchiverArchive, newUnarrArchive}
var extractors = []func(string, string) ([]string, error){unArrExtract, archiverExtract}

func mkdir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	return nil
}



func unArrExtract(sourcePath string, targetPath string) ([]string, error) {
	arc, err := unarr.NewArchive(sourcePath)
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


func archiverExtract(sourcePath string, targetPath string) ([]string, error) {
	f, err := os.Open(sourcePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	format, input, err := archiver.Identify(sourcePath, f)
	if err != nil {
		return nil, err
	}
	err = mkdir(targetPath)
	if err != nil {
		return nil, err
	}
	err = archiverExtractX(format, input, targetPath)
	fmt.Println(err)

	return nil, nil
}




func archiverExtractX(format archiver.Format, input io.Reader, targetPath string) error {

	handler := getHandler(targetPath)

	// want to extract something?
	if ex, ok := format.(archiver.Extractor); ok {
		err := ex.Extract(context.TODO(), input, nil, handler)
		if err != nil {
			return err
		}
	}
	if decom, ok := format.(archiver.Decompressor); ok {
		rc, err := decom.OpenReader(input)
		if err != nil {
			return err
		}
		defer rc.Close()
		// fmt.Println("Decompressor instantiated")
		// read from rc to get decompressed data
	}
	return nil
}

func getHandler(targetPath string) func(ctx context.Context, f archiver.File) error {
	handler := func(ctx context.Context, f archiver.File) error {
		fmt.Println(f.FileInfo.Name())

		x, err := f.Open()
		if err != nil {
			return err
		}

		if f.IsDir() {
			mkdir(path.Join(targetPath, f.NameInArchive))
		} else {
			t, err := os.Create(path.Join(targetPath, f.NameInArchive))
			if err != nil {
				return err
			}
			defer t.Close()
			buf := make([]byte, BUFFERSIZE)
			for {
				n, err := x.Read(buf)
				if err != nil && err != io.EOF {
					return err
				}
				if n == 0 {
					break
				}

				if _, err := t.Write(buf[:n]); err != nil {
					return err
				}
			}
			t.Sync()
		}
		return nil
	}
	return handler
}


func ExtractArchive(sourcePath string, targetPath string) ([]string, error){
	//arc, err := GetArchive(path.Join(archivePath), ARCHIVE_GETTERS)
	for i := 0; i < len(extractors); i++ {
		stuff, err := extractors[i](sourcePath, targetPath)
		if err == nil {
			return stuff, nil
		}

		fmt.Println(err)
	}

	return nil, nil

}

//func (a unarrArchive) basename() string {
//	return a.fileBase.basename()
//}



//type archiverArchive struct {
//	fileBase
//}


// func panicOnErr(err error) {
// 	if err != nil {
// 		panic(err)
// 	}
// }


//func (a archiverArchive) basename() string {
//	return a.fileBase.basename()
//}

//func isDirectory(path string) (bool, error) {
//	fileInfo, err := os.Stat(path)
//	if err != nil {
//		return false, err
//	}
//	return fileInfo.IsDir(), err
//}

// Common interface, base struct and base struct method(s)

//type archive interface {
//	extractAllTo(string) ([]string, error)
//	basename() string
//}

//type fileBase struct {
//	path string
//}

//func (f fileBase) basename() string {
//	basename := path.Base(f.path)
//	return strings.TrimSuffix(basename, path.Ext(basename))
//}

// unarr struct and methods, implements common interface

//type unarrArchive struct {
//	fileBase
//}

// get new archivier, unarr structs

//func newUnarrArchive(archivePath string) (archive, error) {
//	arc, err := unarr.NewArchive(archivePath)
//	if err != nil {
//		return nil, err
//	}
//	defer arc.Close()
//	return unarrArchive{fileBase: fileBase{path: archivePath}}, nil
//}
//
//func newArchiverArchive(archivePath string) (archive, error) {
//	f, err := os.Open(archivePath)
//	if err != nil {
//		return nil, err
//	}
//	defer f.Close()
//	_, _, err = archiver.Identify(archivePath, f)
//	if err != nil {
//		return nil, err
//	}
//	return archiverArchive{fileBase: fileBase{path: archivePath}}, nil
//}

// try to get either an archiver or unarr struct from the input file

//func GetArchive(archivePath string, archiveGetters []func(string) (archive, error)) (archive, error) {
//	for i := 0; i < len(archiveGetters); i++ {
//		archive, err := archiveGetters[i](archivePath)
//		if err == nil {
//			return archive, nil
//		}
//
//		// fmt.Println(err)
//	}
//	return nil, errors.New("no compatible unarchiver found")
//}

// func getTargetSubDir(targetPath string, arc archive) (string, error) {

// 	for i := 0; i < 100; i++ {
// 		subDir := path.Join(targetPath, arc.basename()+"_"+strconv.Itoa(i))
// 		if _, err := os.Stat(subDir); os.IsNotExist(err) {
// 			err = mkdir(subDir)
// 			if err != nil {
// 				return "", err
// 			}
// 			return subDir, nil
// 		}
// 	}
// 	return "", errors.New("no subdir name")

// }



// func findByExt(searchPath string, fileExt string) ([]string, error) {
// 	tars := []string{}
// 	err := filepath.Walk(searchPath, func(path_ string, info os.FileInfo, err error) error {
// 		if err == nil && filepath.Ext(info.Name()) == fileExt {
// 			tars = append(tars, path.Join(path_, info.Name()))
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		return []string{}, err
// 	}
// 	return tars, nil
// }
