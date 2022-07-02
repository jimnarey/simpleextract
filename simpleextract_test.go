package main_test

import (
	"os"
	"path"
	"simpleextract"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

const FIXTURES_PATH string = "./fixtures"
const OUT_PATH string = "./fixtures/out"

func TestSimpleExtract(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Simple Extract Suite")
}

var _ = Describe("Simple Extract", func() {

	DescribeTable("Extracting compatible archives",
		func(filePath string, targetPath string, outFile string) {

			targetDir := path.Join(OUT_PATH, path.Base(filePath))
			simpleextract.ExtractArchive(filePath, targetDir)
			_, err := os.Stat(path.Join(targetDir, outFile))

			Expect(err).To(BeNil())
		},

		Entry("When archive is 7z", path.Join(FIXTURES_PATH, "file.7z"), OUT_PATH, "file.txt"),
		Entry("When archive is rar", path.Join(FIXTURES_PATH, "file.rar"), OUT_PATH, "file.txt"),
		Entry("When archive is tar", path.Join(FIXTURES_PATH, "file.tar"), OUT_PATH, "file.txt"),
		Entry("When archive is tar.7z", path.Join(FIXTURES_PATH, "file.tar.7z"), OUT_PATH, "file.txt"),
		Entry("When archive is tar.bz2", path.Join(FIXTURES_PATH, "file.tar.bz2"), OUT_PATH, "file.txt"),
		Entry("When archive is tar.gz", path.Join(FIXTURES_PATH, "file.tar.gz"), OUT_PATH, "file.txt"),
		Entry("When archive is tar.xz", path.Join(FIXTURES_PATH, "file.tar.xz"), OUT_PATH, "file.txt"),
		Entry("When archive is zip", path.Join(FIXTURES_PATH, "file.zip"), OUT_PATH, "file.txt"),
	)

})
