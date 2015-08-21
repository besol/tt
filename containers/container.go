package containers

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/flexiant/tt/utils"
	"github.com/hoisie/mustache"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"strings"
)

type Container struct {
	Version     string
	SubTemplate SubTemplate
	Config      map[string]string
}

type SubTemplate struct {
	Before string
	After  string
}

func parseTemplate(filename string, data *Container) (string, error) {
	fileLocation := fmt.Sprintf("%s/%s.mustache", utils.GetOriginFolder(), filename)
	if utils.Exists(fileLocation) {
		log.Debugf("  + Parsing %s template", fileLocation)
		before := fmt.Sprintf("%s/%s/%s/%s.before.mustache", utils.GetOriginFolder(), data.Config["client"], data.Config["enviroment"], filename)
		after := fmt.Sprintf("%s/%s/%s/%s.after.mustache", utils.GetOriginFolder(), data.Config["client"], data.Config["enviroment"], filename)

		if utils.Exists(before) {
			log.Debugf("    + Parsing %s:before template", before)
			data.SubTemplate.Before = mustache.RenderFile(before, data)
		} else {
			log.Debugf("    + Not Parsing %s:before template", before)
		}

		if utils.Exists(after) {
			log.Debugf("      + Parsing %s:after template", after)
			data.SubTemplate.After = mustache.RenderFile(after, data)
		} else {
			log.Debugf("    + Not Parsing %s:after template", after)
		}

		parsedData := mustache.RenderFile(fileLocation, data)
		return parsedData, nil
	}
	return "", errors.New(fmt.Sprintf("Can not find %s", fileLocation))
}

func writeTemplate(homeDir string, filename string, data *Container) error {
	fileLocation := filepath.Join(homeDir, filename)
	log.Debugf("+ Creating file template  %s", fileLocation)
	templateContent, err := parseTemplate(filename, data)
	if err == nil {
		file, err := os.Create(fileLocation)
		if err != nil {
			return err
		}
		_, err = file.Write([]byte(templateContent))
		if err != nil {
			return err
		}
		defer file.Close()
	} else {
		return err
	}
	return nil
}

func PrepareContainer(container *Container) (homeFolder string, error error) {

	configContent, err := utils.ReadFileIntoByte(os.Getenv("TT_CONFIG"))
	if err != nil {
		return "", errors.New(fmt.Sprintf("We could not find config file at %s", os.Getenv("TT_CONFIG")))
	}
	var options map[string]string
	err = yaml.Unmarshal(configContent, &options)
	if err != nil {
		return "", err
	}

	container.Config = options
	fmt.Printf("\n#%v\n", container.Config)

	walkDirectories := []string{fmt.Sprintf("%s/common", utils.GetOriginFolder()), fmt.Sprintf("%s/%s/%s", utils.GetOriginFolder(), container.Config["client"], container.Config["enviroment"])}

	directories, files, _, err := utils.WalkDirectory(walkDirectories)

	tree := strings.Split(utils.GetOriginFolder(), "/")

	homeFolder = filepath.Clean(fmt.Sprintf("%s/%s/%s/%s", os.TempDir(), tree[len(tree)-1], container.Config["client"], container.Config["enviroment"]))

	if utils.DirectoryExists(homeFolder) {
		log.Debugf("  + Removing folder %s", homeFolder)
		err = os.RemoveAll(homeFolder)
		if err != nil {
			return "", err
		}
	}

	log.Debugf("+ Creating folder %s", homeFolder)
	err = os.MkdirAll(homeFolder, 0777)
	if err != nil {
		return "", err
	}

	err = writeTemplate(homeFolder, "Dockerfile", container)
	if err != nil {
		return "", err
	}

	err = writeTemplate(homeFolder, "docker-entrypoint.sh", container)
	if err != nil {
		return "", err
	}

	writeTemplate(homeFolder, "docker-entrypoint.new.sh", container)

	ReplaceFolderStrings := []string{utils.GetOriginFolder(), "common", fmt.Sprintf("%s/", container.Config["enviroment"]), fmt.Sprintf("%s/", container.Config["client"])}

	for _, directory := range directories {
		for _, replaceString := range ReplaceFolderStrings {
			directory = strings.Replace(directory, replaceString, "", 1)
		}

		folder := fmt.Sprintf("%s", filepath.Join(homeFolder, directory))
		err := os.MkdirAll(folder, 0777)
		if err != nil {
			return "", err
		}
		log.Debugf("+ Creating folder %s", string(folder))
	}
	for _, file := range files {
		originFile := file
		for _, replaceString := range ReplaceFolderStrings {
			file = strings.Replace(file, replaceString, "", 1)
		}

		destinationFile := fmt.Sprintf("%s", filepath.Join(homeFolder, file))
		utils.CopyFile(originFile, destinationFile)
		log.Debugf("+ Creating file %s", string(destinationFile))
	}

	return filepath.Clean(homeFolder), nil
}
