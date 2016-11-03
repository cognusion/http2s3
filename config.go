package main

import (
	"encoding/json"
	"fmt"
	"github.com/imdario/mergo"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Config is a simple string map with helper functions
type Config map[string]*string

// Returns a pointer to a true string map of the config
func (c Config) Map() *map[string]string {
	configMap := make(map[string]string)
	for k, v := range c {
		configMap[k] = *v
	}
	return &configMap
}

// Returns true if 'key' exists in the map (value irrelevant)
func (c Config) Exists(key string) (exists bool) {
	exists = false
	if _, ok := c[key]; ok {
		exists = true
	}
	return
}

// Returns true if 'key' exists in the map and the value is non-empty
func (c Config) IsNotNull(key string) (isnotnull bool) {
	isnotnull = false
	if v, ok := c[key]; ok {
		if *v != "" {
			isnotnull = true
		}
	}
	return
}

// Returns true if 'key' doesn't exist in the map, or the value is empty
func (c Config) IsNull(key string) (isnull bool) {
	isnull = true
	if v, ok := c[key]; ok {
		if *v != "" {
			isnull = false
		}
	}
	return
}

// Sets the pointer value of 'key'
func (c Config) PSet(key string, value *string) {
	c[key] = value
}

// Sets the value of 'key'
func (c Config) Set(key, value string) {
	c[key] = &value
}

// Returns the string value of 'key', or empty string
func (c Config) Get(key string) (value string) {
	if v, ok := c[key]; ok {
		value = *v
	}
	return
}

// Returns a string array, split on commas from the string value of 'key', or an empty array
func (c Config) GetArray(key string) (value []string) {
	if v, ok := c[key]; ok {
		value = strings.Split(*v, ",")
	}
	return
}

// Loads all of the .json files in the specified directory
func loadConfigs(srcDir string) {
	debugOut.Printf("Looking for configs in '%s'\n", srcDir)
	for _, f := range readDirectoryJsons(srcDir) {
		debugOut.Printf("\tReading config '%s'\n", f)
		err := LoadJSONFile(f, GlobalConfig)
		if err != nil {
			fmt.Printf("Error loading config file '%s': %v\n", f, err)
			os.Exit(1)
		}
	}
}

// Return a string array of all .json files in the specified directory
func readDirectoryJsons(srcDir string) []string {
	// We can skip this error, since our pattern is fixed and known-good.
	files, _ := filepath.Glob(srcDir + "*.json")
	return files
}

// Load a single JSON file, and merge it into the specified Config
func LoadJSONFile(filePath string, conf Config) error {
	defer Track("LoadJSONFile", Now(), debugOut)

	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		debugOut.Printf("Error reading config file '%s': %s\n", filePath, err)
		return err
	}

	newConf, err := JsonToConfig(buf)
	if err != nil {
		debugOut.Printf("Error parsing JSON in config file '%s': %s\n", filePath, err)
		return err
	}

	err = mergo.MergeWithOverwrite(&conf, newConf)
	return nil
}

// Given a JSON byte array, return a Config object or an error
func JsonToConfig(j []byte) (Config, error) {
	newMap := make(Config)
	err := json.Unmarshal(j, &newMap)
	if err != nil {
		return nil, err
	}
	return newMap, nil
}
