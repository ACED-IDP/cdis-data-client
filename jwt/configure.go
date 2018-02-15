package jwt

//go:generate mockgen -destination=mocks/mock_configure.go -package=mocks github.com/uc-cdis/cdis-data-client/jwt ConfigureInterface

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/user"
	"path"
	"strings"
)

type Credential struct {
	KeyId       string
	APIKey      string
	AccessKey   string
	APIEndpoint string
}

type Configure struct{}

type ConfigureInterface interface {
	ReadFile(string, string) string
	ReadCredentials(string) Credential
	ParseUrl() string
	ReadLines(Credential, []byte, string, string) ([]string, bool)
	UpdateConfigFile(Credential, []byte, string, string, string)
	TestMock() bool
}

func (conf *Configure) ReadFile(file_path string, file_type string) string {
	//Look in config file
	var full_file_path string
	if file_path[0] == '~' {
		usr, _ := user.Current()
		homeDir := usr.HomeDir
		full_file_path = homeDir + file_path[1:]
	} else {
		full_file_path = file_path
	}
	if _, err := os.Stat(full_file_path); err != nil {
		fmt.Println("File specified at " + full_file_path + " not found")
		return ""
	}

	content, err := ioutil.ReadFile(full_file_path)
	if err != nil {
		panic(err)
	}

	content_str := string(content[:])

	if file_type == "json" {
		content_str = strings.Replace(content_str, "\n", "", -1)
	}
	return content_str
}

func (conf *Configure) ReadCredentials(filePath string) Credential {
	var configuration Credential
	jsonContent := conf.ReadFile(filePath, "json")
	jsonContent = strings.Replace(jsonContent, "key_id", "KeyId", -1)
	jsonContent = strings.Replace(jsonContent, "api_key", "APIKey", -1)
	err := json.Unmarshal([]byte(jsonContent), &configuration)
	if err != nil {
		fmt.Println("Cannot read json file: " + err.Error())
		os.Exit(1)
	}
	return configuration
}

func (conf *Configure) ParseUrl() string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("API endpoint: ")
	scanner.Scan()
	apiEndpoint := scanner.Text()
	parsed_url, err := url.Parse(apiEndpoint)
	if err != nil {
		panic(err)
	}
	if parsed_url.Host == "" {
		fmt.Print("Invalid endpoint. A valid endpoint looks like: https://www.tests.com\n")
		os.Exit(1)
	}
	return apiEndpoint
}

func (conf *Configure) TryReadConfigFile() (string, []byte, error) {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	homeDir := usr.HomeDir
	configPath := path.Join(homeDir + "/.cdis/config")
	u := Utils{}
	content, err := u.TryReadFile(configPath)
	return configPath, content, err
}

func (conf *Configure) ReadLines(cred Credential, configContent []byte, apiEndpoint string, profile string) ([]string, bool) {
	lines := strings.Split(string(configContent), "\n")

	found := false
	for i := 0; i < len(lines); i += 6 {
		if lines[i] == "["+profile+"]" {
			if cred.KeyId != "" {
				lines[i+1] = "key_id=" + cred.KeyId
			}
			if cred.APIKey != "" {
				lines[i+2] = "api_key=" + cred.APIKey
			}
			lines[i+3] = "access_key=" + cred.AccessKey
			if apiEndpoint != "" {
				lines[i+4] = "api_endpoint=" + apiEndpoint
			}
			found = true
			break
		}
	}
	return lines, found
}

func (conf *Configure) UpdateConfigFile(cred Credential, configContent []byte, apiEndpoint string, configPath string, profile string) {
	lines, found := conf.ReadLines(cred, configContent, apiEndpoint, profile)
	if found {
		f, err := os.OpenFile(configPath, os.O_WRONLY|os.O_TRUNC, 0777)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				panic(err)
			}
		}()
		for i := 0; i < len(lines)-1; i++ {
			f.WriteString(lines[i] + "\n")
		}
	} else {
		f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				panic(err)
			}
		}()

		_, err = f.WriteString("[" + profile + "]\n" +
			"key_id=" + cred.KeyId + "\n" +
			"api_key=" + cred.APIKey + "\n" +
			"access_key=" + cred.AccessKey + "\n" +
			"api_endpoint=" + apiEndpoint + "\n\n")

		if err != nil {
			panic(err)
		}
	}
}

func (conf *Configure) TestMock() bool {
	return true
}
