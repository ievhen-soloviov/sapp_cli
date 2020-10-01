package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/thatisuday/commando"
)

type apiEnvironment struct {
	URL string `json:"external_url"`
}

type apiResponse []*apiEnvironment

var projectID = "21001347"
var token = "cZkc7ixb68xifSyDVV_d"
var file []byte

func init() {

}

func main() {
	commando.SetExecutableName("sapp").
		SetVersion("1.0.0").
		SetDescription("This tool lists the available Sapp API's.")

	commando.
		Register("api").
		SetAction(getApis)

	commando.Parse(nil)
}

func getApis(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
	resp, err := http.Get("https://gitlab.com/api/v4/projects/" + projectID + "/environments?private_token=" + token)
	if err != nil {
		fmt.Println("[ERROR] Can't connect.")
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("[ERROR] Can't read response.")
		return
	}

	resp.Body.Close()

	var decoded apiResponse
	urlList := []string{}

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

	file, err = ioutil.ReadFile("./.env")
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
	err = ioutil.WriteFile("./.env", []byte(output), 0644)
	if err != nil {
		fmt.Println("[ERROR] Cannot write file.")
		return
	}

	fmt.Printf("[SUCCESS] API URL set to: %s", result)
}
