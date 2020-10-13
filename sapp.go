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
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/thatisuday/commando"
)

type apiEnvironment struct {
	URL   string `json:"external_url"`
}

type apiResponse []*apiEnvironment

type configFile struct {
	GitlabToken string `json:"token"`
}

var projectID = "21001347"
var token string
var fileMode os.FileMode = 0644

func getToken() (string, error) {
	localToken, exists := os.LookupEnv("GITLAB_TOKEN")
	if exists == true {
		return localToken, nil
	}

	exePath, err := os.Executable()
	if err != nil {
		err = errors.New("[ERROR] Could not find executable")
		return "", err
	}

	data, err := ioutil.ReadFile(path.Join(filepath.Dir(exePath), "config.json"))
	if err != nil {
		err = errors.New("[ERROR] Could not read config file. Please run 'sapp config'")
		return "", err
	}

	decoded := configFile{}
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		err = errors.New("[ERROR] Could not decode data")
		return "", err
	}
	return decoded.GitlabToken, nil
}

func main() {
	commando.SetExecutableName("sapp").
		SetVersion("1.0.0").
		SetDescription("This tool lists the available Sapp API's.")

	commando.
		Register("api").
		SetAction(getApis)

	commando.
		Register("config").
		SetAction(config)

	commando.Parse(nil)
}

func getApis(_ map[string]commando.ArgValue, _ map[string]commando.FlagValue) {
	token, err := getToken()
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := http.Get("https://gitlab.com/api/v4/projects/" + projectID + "/environments?private_token=" + token + "&states=available")
	if err != nil {
		fmt.Println("[ERROR] Can't connect.")
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("[ERROR] Can't read response.")
		return
	}

	err = resp.Body.Close()
	if err != nil {
		fmt.Println("[ERROR] Could not close response.")
		return
	}

	var decoded apiResponse
	var urlList []string

	err = json.Unmarshal(data, &decoded)
	for _, a := range decoded {
		if a.URL != "" {
			urlList = append(urlList, a.URL)
		}
	}

	prompt := promptui.Select{
		Label: "Select an API",
		Items: urlList,
		Size:  len(urlList),
	}

	_, result, err := prompt.Run()
	if err != nil {
		fmt.Println("Cancelled.")
		return
	}

	file, err := ioutil.ReadFile("./.env")
	if err != nil {
		fmt.Println("[ERROR] Cannot read .env file.")
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
		fmt.Println("[ERROR] Cannot write file.")
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
		fmt.Println("[ERROR] Could not read response.")
		return
	}
	token = result

	config := configFile{
		GitlabToken: result,
	}
	encoded, err := json.Marshal(config)

	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("[ERROR] Could not find executable.")
		return
	}
	err = ioutil.WriteFile(filepath.Join(filepath.Dir(exePath), "config.json"), encoded, fileMode)
	if err != nil {
		fmt.Println("[ERROR] Could not write to config file.")
		return
	}
}
