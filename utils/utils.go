package utils

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func GetOriginFolder() string {
	return filepath.Clean(os.Getenv("TT_ORIGIN"))
}

func FlagsRequired(c *cli.Context, flags []string) {
	parameters := false
	for _, flag := range flags {
		if !c.IsSet(flag) {
			log.Warn(fmt.Sprintf("Please use parameter --%s", flag))
			parameters = true
		}
	}

	if parameters {
		fmt.Printf("\n")
		cli.ShowCommandHelp(c, c.Command.Name)
		os.Exit(2)
	}
}

func Exists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func DirectoryExists(name string) bool {
	stats, err := os.Stat(name)
	if err != nil {
		return false
	} else if stats.IsDir() {
		return true
	}
	return false
}

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func ReadFileIntoByte(fileLocation string) ([]byte, error) {
	if Exists(fileLocation) {
		file, err := os.Open(fileLocation)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}

		return bytes, nil
	}
	return nil, errors.New(fmt.Sprintf("We could not find file %s", fileLocation))
}

func WalkDirectory(walkDirectories []string) (directoryList []string, fileList []string, templateList []string, err error) {
	for _, value := range walkDirectories {
		log.Debugf("+ Going to Walk %s", value)
		if DirectoryExists(value) {
			err = filepath.Walk(value, func(path string, f os.FileInfo, err error) error {
				log.Debugf("  + %s", f.Name())
				if f.IsDir() {
					directoryList = append(directoryList, filepath.Clean(path))
				} else {
					if strings.HasPrefix(f.Name(), ".") {
						return nil
					}
					if !strings.Contains(f.Name(), ".mustache") {
						fileList = append(fileList, filepath.Clean(path))
					} else {
						if !strings.Contains(f.Name(), ".after") || !strings.Contains(f.Name(), ".before") {
							templateList = append(templateList, filepath.Clean(path))
						}
					}
				}
				return nil
			})
		}
		if err != nil {
			return []string{}, []string{}, []string{}, err
		}
	}
	return directoryList, fileList, templateList, nil
}

func CopyFile(src string, dst string) error {
	if Exists(src) {
		// Read all content of src to data
		data, err := ioutil.ReadFile(src)
		if err != nil {
			return err
		}
		info, err := os.Stat(src)
		if err != nil {
			return err
		}
		mode := info.Mode()
		// Write data to dst
		err = ioutil.WriteFile(dst, data, mode)
		if err != nil {
			return err
		}
		return nil
	} else {
		return errors.New(fmt.Sprintf("We can not find %s file to copy to %s", src, dst))
	}
}
