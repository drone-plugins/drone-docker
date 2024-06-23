package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/errors"
)

const (
	v2HubRegistryURL string = "https://registry.hub.docker.com/v2/"
	v1RegistryURL    string = "https://index.docker.io/v1/" // Default registry
	v2RegistryURL    string = "https://index.docker.io/v2/" // v2 registry is not supported
)

type (
	Auth struct {
		Auth string `json:"auth"`
	}

	Config struct {
		Auths       map[string]Auth   `json:"auths"`
		CredHelpers map[string]string `json:"credHelpers,omitempty"`
	}
)

type RegistryCredentials struct {
	Registry string
	Username string
	Password string
}

func NewConfig() *Config {
	return &Config{
		Auths:       make(map[string]Auth),
		CredHelpers: make(map[string]string),
	}
}

func (c *Config) SetAuth(registry, username, password string) {
	authBytes := []byte(username + ":" + password)
	encodedString := base64.StdEncoding.EncodeToString(authBytes)
	log.Printf("auth : %s", encodedString)
	c.Auths[registry] = Auth{Auth: encodedString}
}

func (c *Config) SetCredHelper(registry, helper string) {
	c.CredHelpers[registry] = helper
}

func (c *Config) CreateDockerConfigJson(credentials []RegistryCredentials) ([]byte, error) {
	for _, cred := range credentials {
		if cred.Registry != "" {

			if cred.Username == "" {
				return nil, fmt.Errorf("Username must be specified for registry: %s", cred.Registry)
			}
			if cred.Password == "" {
				return nil, fmt.Errorf("Password must be specified for registry: %s", cred.Registry)
			}
			c.SetAuth(cred.Registry, cred.Username, cred.Password)
		}
	}

	jsonBytes, err := json.Marshal(c)
	log.Printf("jsonBytes config : %s", jsonBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize docker config json")
	}

	return jsonBytes, nil
}

func WriteDockerConfig(data []byte, path string) (string error) {
	err := os.MkdirAll(path, 0600)
	if err != nil {
		if !os.IsExist(err) {
			return errors.Wrap(err, fmt.Sprintf("failed to create %s directory", path))
		}
	}

	filePath := path + "/config.json"
	log.Printf("Config data is %s", data)
	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create docker config file at %s", path))
	}
	return nil
}
