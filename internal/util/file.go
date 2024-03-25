package util

import (
	"log"
	"os"
	"strconv"
	"strings"
)

func ReadIntFromFile(path string) (int64, error) {
	fileBuffer, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	value := string(fileBuffer)
	value = strings.TrimSpace(value)
	return strconv.ParseInt(value, 0, 64)
}

func WriteIntToFile(value int, path string) error {
	touch(path)
	fileStat, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}

	return os.WriteFile(path, []byte(strconv.Itoa(value)), fileStat.Mode())
}

func touch(path string) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()
	} else if err != nil {
		panic(err)
	}
}