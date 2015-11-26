package main
/*
 *  Add to .gitignore
 *     *.gen.go
 */

import (
	"os"
	"strings"
	"path/filepath"
	"log"
)

func main() {
	if len(os.Args) < 2 {
    gopath_envr := os.Getenv("GOPATH")
    if len(gopath_envr) > 0 {
      log.Println("GOPATH: ", gopath_envr)
      gopaths:= strings.Split(gopath_envr, ";")
      for _, gopath := range gopaths {
        src_path, _ := filepath.Abs(gopath + "/src/")
        GenErrorInSrc(src_path)
      }
    }
	} else {
    GenErrorInSrc(os.Args[1])
	}
}
