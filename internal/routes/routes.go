package routes

import (
	"io/ioutil"
	"os"

	"github.com/go-yaml/yaml"
)

type Route struct {
	IncomingRequestPath  string `yaml:"incoming_request_path"`
	ForwardedRequestURL  string `yaml:"forwarded_request_url"`
	ForwardedRequestPath string `yaml:"forwarded_request_path"`
}

type RoutesConfig struct {
	Routes []Route `yaml:routes`
}

func NewRoutesConfigFromYaml(yamlPath string) (*RoutesConfig, error) {
	routesConfig := RoutesConfig{}

	routesYaml, err := ioutil.ReadFile(yamlPath)
	routesYaml = []byte(os.ExpandEnv(string(routesYaml)))

	if err != nil {
		return nil, err
	}

	yaml.Unmarshal(routesYaml, &routesConfig)

	return &routesConfig, nil
}
