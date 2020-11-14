package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/thatisuday/commando"
)

type apiEnvironment struct {
	URL string `json:"external_url"`
}

type apiResponse []*apiEnvironment

type configFileType struct {
	GitlabToken string `json:"token"`
	ProjectID   string `json:"projectID"`

	// list of env vars to search for
	Vars []string `json:"vars"`
}

var configFile configFileType

var defaultConfig = configFileType{
	ProjectID: "21001347",
	Vars:      []string{"APP_API_URL", "SAPP_URL"},
}

var fileMode os.FileMode = 0644

func main() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("[ERROR] Could not find executable")
		return
	}

	data, err := ioutil.ReadFile(path.Join(filepath.Dir(exePath), "config.json"))
	if err != nil {
		// no config file found, create one with default values
		newConfig := configFileType{
			ProjectID: defaultConfig.ProjectID,
			Vars:      defaultConfig.Vars,
		}
		err = writeConfigFile(newConfig)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("[ERROR] New config file created. Please run 'sapp config token'")
		return
	}

	err = json.Unmarshal(data, &configFile)
	if err != nil {
		fmt.Println("[ERROR] Could not decode data")
		return
	}

	commando.SetExecutableName("sapp").
		SetVersion("1.0.0").
		SetDescription("This tool lists the available Sapp API's.")

	commando.
		Register("api").
		AddArgument(
			"action",
			"You can either 'get' or 'set' environments.",
			"get").
		SetAction(api)

	commando.
		Register("config").
		AddArgument(
			"setting",
			"This parameter specifies, which setting is to be configured",
			"").
		SetAction(config)

	commando.Parse(nil)
}

func api(args map[string]commando.ArgValue, _ map[string]commando.FlagValue) {
	list, err := getURLs(configFile.ProjectID, configFile.GitlabToken)
	if err != nil {
		fmt.Println(err)
	}

	action := args["action"].Value
	switch action {
	case "get":
		for _, url := range list {
			fmt.Println(url)
		}
	case "set":
		set(list)
	}
}

func set(list []string) {
	prompt := promptui.Select{
		Label: "Select an API",
		Items: list,
		Size:  len(list),
	}

	_, result, err := prompt.Run()
	if err != nil {
		fmt.Println("Cancelled")
		return
	}

	file, err := ioutil.ReadFile("./.env")
	if err != nil {
		fmt.Println("[ERROR] Cannot read .env file")
		return
	}

	lines := strings.Split(string(file), "\n")

	expression := "(" + strings.Join(configFile.Vars, ")|(") + ")"

	for i, line := range lines {
		matches, err := regexp.MatchString(expression, line)
		if err != nil {
			fmt.Println("[ERROR] Could not find variable in .env file")
		}
		if matches {
			chunks := strings.Split(line, "=")
			chunks[1] = result
			lines[i] = strings.Join(chunks, "=")
		}
	}

	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile("./.env", []byte(output), fileMode)
	if err != nil {
		fmt.Println("[ERROR] Cannot write file")
		return
	}

	fmt.Printf("[SUCCESS] API URL set to: %s", result)
}

func config(args map[string]commando.ArgValue, _ map[string]commando.FlagValue) {
	setting := args["setting"].Value
	switch setting {
	case "token":
		prompt := promptui.Prompt{
			Label: "Please enter your Gitlab token",
		}
		result, err := prompt.Run()
		if err != nil {
			fmt.Println("[ERROR] Could not read response")
			return
		}
		configFile.GitlabToken = result
		err = writeConfigFile(configFile)
		if err != nil {
			fmt.Println(err)
		}

	case "project":
		prompt := promptui.Prompt{
			Label: "Please enter your Gitlab project ID",
		}
		result, err := prompt.Run()
		if err != nil {
			fmt.Println("[ERROR] Could not read response")
			return
		}
		configFile.ProjectID = result
		err = writeConfigFile(configFile)
		if err != nil {
			fmt.Println(err)
		}

	case "vars":
		prompt := promptui.Prompt{
			Label: "Please enter the list of possible variable names, separated by spaces",
		}
		result, err := prompt.Run()
		if err != nil {
			fmt.Println("[ERROR] Could not read response")
			return
		}
		configFile.Vars = strings.Split(result, " ")
		err = writeConfigFile(configFile)
		if err != nil {
			fmt.Println(err)
		}

	case "reset":
		err := writeConfigFile(defaultConfig)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("[SUCCESS] Config has been reset to default values")
	}

}

// create/edit a config file in the executable directory
func writeConfigFile(data configFileType) (err error) {
	encoded, err := json.Marshal(data)

	exePath, err := os.Executable()
	if err != nil {
		err = errors.New("[ERROR] Could not find executable")
		return
	}
	err = ioutil.WriteFile(filepath.Join(filepath.Dir(exePath), "config.json"), encoded, fileMode)
	if err != nil {
		err = errors.New("[ERROR] Could not write to config file")
		return
	}
	return
}

func getURLs(projectID, token string) (urlList []string, err error) {
	if configFile.GitlabToken == "" {
		fmt.Println("[ERROR] GitLab token not found. Please run 'sapp config token'")
		return
	}

	resp, err := http.Get("https://gitlab.com/api/v4/projects/" + projectID + "/environments?private_token=" + token + "&states=available")
	if err != nil {
		err = errors.New("[ERROR] Can't connect")
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.New("[ERROR] Can't read response")
		return
	}

	err = resp.Body.Close()
	if err != nil {
		err = errors.New("[ERROR] Could not close response")
		return
	}

	var decoded apiResponse

	err = json.Unmarshal(data, &decoded)
	for _, a := range decoded {
		if a.URL != "" {
			urlList = append(urlList, a.URL)
		}
	}
	if len(urlList) < 1 {
		err = errors.New("[ERROR] No environments found in this project")
	}

	return
}
