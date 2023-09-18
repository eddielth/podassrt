package conf

import (
	"fmt"
	gotoml "github.com/pelletier/go-toml"
	"io/ioutil"
)

// Load the configuration file
func LoadFromFile(configuration interface{}) error {
	contents, err := ioutil.ReadFile("./res/configuration.toml")
	if err != nil {
		return fmt.Errorf("could not load configuration file: %v", err.Error())
	}
	// Decode the configuration from TOML
	err = gotoml.Unmarshal(contents, configuration)
	if err != nil {
		return fmt.Errorf("unable to parse configuration file: %v", err.Error())
	}
	return nil
}

type RestartPolicy struct {
	Name    string
	Targets []string
}

type Configuration struct {
	Common struct {
		KubeConfigFile string
		Namespace      string
		LabelKey       string
	}
	RestartPolicy []RestartPolicy
}
