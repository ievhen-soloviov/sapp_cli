package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/thatisuday/commando"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type apiEnvironment struct {
	URL   string `json:"external_url"`
}

type apiResponse []*apiEnvironment

type configFile struct {
	GitlabToken string `json:"token"`
}

var projectID = "21001347"
//var token string
var fileMode os.FileMode = 0644

func getToken() (localToken string, err error) {
	localToken, exists := os.LookupEnv("GITLAB_TOKEN")
	if exists == true {
		return
	}

	exePath, err := os.Executable()
	if err != nil {
		err = errors.New("[ERROR] Could not find executable")
		return
	}

	data, err := ioutil.ReadFile(path.Join(filepath.Dir(exePath), "config.json"))
	if err != nil {
		err = errors.New("[ERROR] Could not read config file. Please run 'sapp config'")
		return
	}

	decoded := configFile{}
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		err = errors.New("[ERROR] Could not decode data")
		return
	}
	localToken = decoded.GitlabToken

	return
}

func main() {
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
		SetAction(config)

	commando.Parse(nil)
}

func api(args map[string]commando.ArgValue, _ map[string]commando.FlagValue) {
	token, err := getToken()
	if err != nil {
		fmt.Println(err)
		return
	}

	list, err := getURLs(projectID, token)
	if err != nil {
		fmt.Println(err)
	}

	action := args["action"].Value
	switch action {
	case "":
		fmt.Println("Please enter an argument: 'get' of 'set'")
		return
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

	for i, line := range lines {
		if strings.HasPrefix(line, "APP_API_URL") || strings.HasPrefix(line, "SAPP_URL") {
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

func config(_ map[string]commando.ArgValue, _ map[string]commando.FlagValue) {
	prompt := promptui.Prompt{
		Label: "Please enter your Gitlab token",
	}
	result, err := prompt.Run()
	if err != nil {
		fmt.Println("[ERROR] Could not read response")
		return
	}

	config := configFile{
		GitlabToken: result,
	}
	encoded, err := json.Marshal(config)

	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("[ERROR] Could not find executable")
		return
	}
	err = ioutil.WriteFile(filepath.Join(filepath.Dir(exePath), "config.json"), encoded, fileMode)
	if err != nil {
		fmt.Println("[ERROR] Could not write to config file")
		return
	}
}

func getURLs(projectID, token string) (urlList []string, err error) {
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
