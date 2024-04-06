package persistence

import (
	"encoding/json"
	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/util"
	"log"
	"os"
	"path"
)

var (
	BaseDir = path.Join(configuration.BaseDir, "persistence")
)

func init() {
	err := os.MkdirAll(BaseDir, 0755)
	if err != nil {
		log.Fatal(err)
	}
}

func SaveInt(key string, value int) error {
	file := path.Join(BaseDir, key+".sav")
	return util.WriteIntToFile(value, file)
}

func ReadInt(key string) (int64, error) {
	file := path.Join(BaseDir, key+".sav")
	return util.ReadIntFromFile(file)
}

func SaveStruct(key string, value interface{}) error {
	file := path.Join(BaseDir, key+".sav")
	jsonString, _ := json.MarshalIndent(value, "", "  ")
	return os.WriteFile(file, jsonString, os.ModePerm)
}

func ReadStruct(key string, target interface{}) error {
	file := path.Join(BaseDir, key+".sav")
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, target)
	if err != nil {
		return err
	}

	return nil
}
