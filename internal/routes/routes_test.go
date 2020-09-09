package routes

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRoutesConfigFromYaml(t *testing.T) {
	assert := assert.New(t)

	yamlFile, err := ioutil.TempFile("", "testRoutes.yml")
	assert.Nil(err)
	defer yamlFile.Close()
	defer os.Remove(yamlFile.Name())

	routesYaml := `routes:
  - incoming_request_path: '/route1'
    forwarded_request_url: 'http://service1.com'
    forwarded_request_path: '/api/ping'

  - incoming_request_path: '/route2'
    forwarded_request_url: 'http://service2.com'
    forwarded_request_path: '/api/pong'`

	_, err = yamlFile.WriteString(routesYaml)
	assert.Nil(err)

	t.Run("when the YAML file exists and contains valid YAML", func(t *testing.T) {
		routesConfig, err := NewRoutesConfigFromYaml(yamlFile.Name())
		assert.Nil(err)

		routes := routesConfig.Routes

		route1 := routes[0]
		assert.Equal("/route1", route1.IncomingRequestPath)
		assert.Equal("http://service1.com", route1.ForwardedRequestURL)
		assert.Equal("/api/ping", route1.ForwardedRequestPath)

		route2 := routes[1]
		assert.Equal("/route2", route2.IncomingRequestPath)
		assert.Equal("http://service2.com", route2.ForwardedRequestURL)
		assert.Equal("/api/pong", route2.ForwardedRequestPath)
	})

	t.Run("when the YAML file includes expandable ENV variables", func(t *testing.T) {
		url := "http://env-service.com"
		os.Setenv("FORWARDED_REQUEST_URL", url)

		envRoutesYaml := `routes:
  - incoming_request_path: '/route1'
    forwarded_request_url: ${FORWARDED_REQUEST_URL}
    forwarded_request_path: '/api/ping'`

		envYamlFile, err := ioutil.TempFile("", "testRoutesWithEnv.yml")
		assert.Nil(err)
		defer envYamlFile.Close()
		defer os.Remove(envYamlFile.Name())

		_, err = envYamlFile.WriteString(envRoutesYaml)
		assert.Nil(err)

		routesConfig, err := NewRoutesConfigFromYaml(envYamlFile.Name())
		assert.Nil(err)

		route := routesConfig.Routes[0]

		assert.Equal(url, route.ForwardedRequestURL)
	})

	t.Run("when the YAML file exists and contains invalid YAML", func(t *testing.T) {
		file, err := ioutil.TempFile("", "testInvalid.yml")
		assert.Nil(err)
		defer file.Close()
		defer os.Remove(file.Name())

		routesJson := `{"routes": []}`

		_, err = file.WriteString(routesJson)
		assert.Nil(err)

		routesConfig, err := NewRoutesConfigFromYaml(file.Name())
		assert.Nil(err)
		assert.Empty(routesConfig.Routes)
	})

	t.Run("when the given file does not exist", func(t *testing.T) {
		_, err := NewRoutesConfigFromYaml("does_not_exist.yml")
		assert.EqualError(err, "open does_not_exist.yml: no such file or directory")
	})
}
