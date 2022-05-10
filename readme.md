# ScholarshipApp command line tool

## Install

`go install github.com/ievhen-soloviov/sapp_cli@v0.1.4`

Check that `$GOPATH/bin` is included in your PATH

## Configuration

At first launch, a json config file with default configurations is created in the executable directory. Token has no default value.

`sapp_cli config token`: enter your Gitlab token (you can get it at Gitlab account settings https://gitlab.com/profile/personal_access_tokens)

`sapp_cli config project`: enter your GitLab project ID (default: ScholarshipApp API ID)

`sapp_cli config vars`: enter a list of possible variable names to search for in .env files (separated by spaces). Default: `APP_API_URL`, `SAPP_URL`

`sapp_cli config reset`: brings back default values.

## Config File

`token`: your GitLab token

`projectID`

`vars`

`urls`: - a list of custom API URLs

## API URL actions

Use `sapp_cli api get` to display the list of all active environments.

Use `sapp_cli api set` to select an active environment URL and copy it into the `.env` file in the current directory.
