package pkg

import "path/filepath"

func BaseNoExt(fileName string) string {
	fileName = filepath.Base(fileName)
	if(fileName == "/") {
		return ""
	}
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}
