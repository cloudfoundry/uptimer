package cfAppGenerator

import (
	"os"
	"path"
)

type staticfileAppGenerator struct {
	tempDirFunc   tempDirFunc
	writeFileFunc writeFileFunc
}

func (g *staticfileAppGenerator) Path() (string, error) {
	tempDir, err := g.tempDirFunc("", "uptimer")
	if err != nil {
		return "", err
	}

	indexFilePath := path.Join(tempDir, "index.html")
	err = g.writeFileFunc(indexFilePath, []byte("<b>hello</b>"), os.ModePerm)
	if err != nil {
		return "", err
	}

	staticFilePath := path.Join(tempDir, "Staticfile")
	err = g.writeFileFunc(staticFilePath, []byte{}, os.ModePerm)
	if err != nil {
		return "", err
	}

	return tempDir, nil
}
