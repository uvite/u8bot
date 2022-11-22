package main

import (
	"fmt"
	"github.com/c9s/bbgo/u8/vm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/uvite/u8/loader"
	"os"
)

func main() {
	logger := log.New()

	//data := []byte(`test contents`)
	fs := afero.NewOsFs()
	pwd, err := os.Getwd()
	fmt.Println(pwd)
	jsFile := "bbgo.js"
	sourceData, err := loader.ReadSource(logger, jsFile, pwd, map[string]afero.Fs{"file": fs}, nil)
	fmt.Println(sourceData.Data, err)
	a := vm.NewJsVm(sourceData.Data)
	a.GetInit()
}
