package util

import (
	"log"
	"os"
	"strconv"
	"strings"
)

func ReadTextFromFile(path string) (value string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ReadIntFromFile(path string) (int64, error) {
	fileBuffer, err := os.ReadFile(path)
	if err != nil {
		return -1, err
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

func ReadFloatFromFile(path string) (float64, error) {
	fileBuffer, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	value := string(fileBuffer)
	value = strings.TrimSpace(value)
	return strconv.ParseFloat(value, 64)
}

func WriteFloatToFile(value float64, path string) error {
	touch(path)
	fileStat, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}

	return os.WriteFile(path, []byte(strconv.FormatFloat(value, 'f', -1, 64)), fileStat.Mode())
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
