package config

import (
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/flxy0/wakeru/helpers"
)

//go:embed default.config.toml
var defaultConfig embed.FS

type configVals struct {
	Name     string
	Image    string
	Generate bool
}

var (
	ConfigInstanceName      string
	ConfigInstanceImagePath string
	ConfigAllowGeneration   bool
)

func ParseConfig() {
	var configPath string

	dirContents, dirErr := os.ReadDir(".")
	helpers.LogErr(dirErr)

	for _, f := range dirContents {
		if f.Name() == "config.toml" {
			configPath = "./config.toml"
		}
	}

	var confStruct configVals

	if configPath != "" {
		_, err := toml.DecodeFile(configPath, &confStruct)
		helpers.LogErr(err)
	} else {
		fmt.Println("no config.toml found in current directory")
		fmt.Println("using standard config!")
		fmt.Println("if you'd like to customise your instance, use `wakeru init`")

		_, err := toml.DecodeFS(fs.FS(defaultConfig), "default.config.toml", &confStruct)
		helpers.LogErr(err)
	}

	ConfigInstanceName = confStruct.Name
	ConfigInstanceImagePath = confStruct.Image
	ConfigAllowGeneration = confStruct.Generate
}

func GenerateDefaultConfigInCurrentDir() {
	configFile, fileErr := os.Create("config.toml")
	helpers.LogErr(fileErr)

	defaultConfigContents, configErr := defaultConfig.ReadFile("default.config.toml")
	helpers.LogErr(configErr)

	configFile.Write(defaultConfigContents)
	configFile.Close()
}
