package data

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"

	"gopkg.in/yaml.v2"
)

type Server struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type User struct {
	User  string `yaml:"user"`
	Token string `yaml:"token"`
}

type Config struct {
	Server Server `yaml:"server"`
	User   User   `yaml:"user"`
}

func GetConfig() Config {
	defaults := map[string]interface{}{
		"server": map[string]interface{}{
			"host": "localhost",
			"port": "8068",
		},
		"user": map[string]interface{}{
			"user":  "",
			"token": "",
		},
	}

	if _, err := os.Stat(GetConfigFile()); err != nil {
		config, _ := toStruct(defaults)
		return config
	}

	file, err := os.Open(GetConfigFile())
	if err != nil {
		config, _ := toStruct(defaults)
		return config
	}
	defer file.Close()

	data, _ := ioutil.ReadAll(file)
	var conf map[string]interface{}

	if err = yaml.Unmarshal(data, &conf); err != nil {
		config, _ := toStruct(defaults)
		return config
	}

	config, _ := toStruct(merge(defaults, conf))
	return config
}

func WriteConfig(conf Config) error {
	data, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}

	// Truncates the file it exists already. A Config object always contains the
	// entire set of configuration options, including any customized ones.
	return ioutil.WriteFile(GetConfigFile(), data, os.ModePerm)
}

func GetConfigFile() string {
	usrDir, _ := os.UserHomeDir()
	return path.Join(usrDir, ".luxbox.yaml")
}

func toStruct(data map[string]interface{}) (Config, error) {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return Config{}, nil
	}

	var config Config
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func merge(left, right map[string]interface{}) map[string]interface{} {
	for key, leftVal := range left {
		rightVal, ok := right[key]
		if ok {
			leftMap, leftMapOk := mapify(leftVal)
			rightMap, rightMapOk := mapify(rightVal)

			if leftMapOk && rightMapOk {
				rightVal = merge(leftMap, rightMap)
			}

			left[key] = rightVal
		}
	}

	return left
}

func mapify(i interface{}) (map[string]interface{}, bool) {
	value := reflect.ValueOf(i)
	if value.Kind() == reflect.Map {
		m := map[string]interface{}{}
		for _, k := range value.MapKeys() {
			m[k.Interface().(string)] = value.MapIndex(k).Interface()
		}

		return m, true
	}

	return map[string]interface{}{}, false
}
