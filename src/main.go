package main
/*
 *  Add to .gitignore
 *     *.gen.go
 */

import (
	"os"
	"strings"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
    gopath_envr := os.Getenv("GOPATH")
    gopaths:= strings.Split(gopath_envr, ";")
    for _, gopath := range gopaths {
      src_path, _ := filepath.Abs(gopath + "/src/")
      GenErrorInSrc(src_path)
    }
	} else {
    GenErrorInSrc(os.Args[1])
	}
}
