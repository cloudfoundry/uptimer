package cfAppGenerator

import "os"

type CfAppGenerator interface {
	Path() (string, error)
}

type tempDirFunc func(dir, prefix string) (string, error)
type writeFileFunc func(filename string, data []byte, perm os.FileMode) error

func NewStaticApp(t tempDirFunc, w writeFileFunc) CfAppGenerator {
	return &staticfileAppGenerator{
		tempDirFunc:   t,
		writeFileFunc: w,
	}
}
