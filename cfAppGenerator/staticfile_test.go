package cfAppGenerator_test

import (
	"fmt"
	"path"

	. "github.com/cloudfoundry/uptimer/cfAppGenerator"

	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Staticfile", func() {
	Describe("AppPath", func() {
		var (
			calledWithDir    string
			calledWithPrefix string

			calledWithFilename []string
			calledWithData     [][]byte
			calledWithPerm     []os.FileMode

			tempDirFunc   func(dir, prefix string) (string, error)
			writeFileFunc func(filename string, data []byte, perm os.FileMode) error

			cfa CfAppGenerator
		)

		BeforeEach(func() {
			tempDirFunc = func(dir, prefix string) (string, error) {
				calledWithDir = dir
				calledWithPrefix = prefix
				return "mytempdir", nil
			}

			writeFileFunc = func(filename string, data []byte, perm os.FileMode) error {
				calledWithFilename = append(calledWithFilename, filename)
				calledWithData = append(calledWithData, data)
				calledWithPerm = append(calledWithPerm, perm)
				return nil
			}

			cfa = NewStaticApp(tempDirFunc, writeFileFunc)
		})

		It("generates a path to a folder with a bare-bones index.html", func() {
			appPath, _ := cfa.Path()

			Expect(calledWithDir).To(BeEmpty())
			Expect(calledWithPrefix).To(Equal("uptimer"))

			Expect(calledWithFilename[0]).To(Equal("mytempdir/index.html"))
			Expect(string(calledWithData[0])).To(Equal("<b>hello</b>"))
			Expect(calledWithPerm[0]).To(Equal(os.ModePerm))

			Expect(calledWithFilename[1]).To(Equal("mytempdir/Staticfile"))
			Expect(string(calledWithData[1])).To(Equal(""))
			Expect(calledWithPerm[1]).To(Equal(os.ModePerm))

			Expect(appPath).To(Equal("mytempdir"))
		})

		It("returns an error if tempDirFunc return an error", func() {
			tempDirFunc = func(dir, prefix string) (string, error) {
				return "", fmt.Errorf("that's just like, your error, man")
			}

			cfa := NewStaticApp(tempDirFunc, writeFileFunc)
			_, err := cfa.Path()

			Expect(err).To(MatchError("that's just like, your error, man"))
		})

		It("returns an error if writeFileFunc return an error writing the index.html", func() {
			writeFileFunc = func(filename string, data []byte, perm os.FileMode) error {
				if path.Base(filename) == "index.html" {
					return fmt.Errorf("nothing went wrong...really")
				}
				return nil
			}

			cfa := NewStaticApp(tempDirFunc, writeFileFunc)
			_, err := cfa.Path()

			Expect(err).To(MatchError("nothing went wrong...really"))
		})

		It("returns an error if writeFileFunc return an error writing the Staticfile", func() {
			writeFileFunc = func(filename string, data []byte, perm os.FileMode) error {
				if path.Base(filename) == "Staticfile" {
					return fmt.Errorf("nothing went wrong...really")
				}
				return nil
			}

			cfa := NewStaticApp(tempDirFunc, writeFileFunc)
			_, err := cfa.Path()

			Expect(err).To(MatchError("nothing went wrong...really"))
		})
	})
})
