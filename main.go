package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func usage() {
	fmt.Println("Usage: aws-mux [profile] && source ~/.aws/env")
	os.Exit(0)
}

func getProfileName() string {
	profileName := "dev"
	if len(os.Args) > 1 {
		for _, i := range os.Args {
			if i == "-h" || i == "--help" {
				usage()
			}
		}
		profileName = os.Args[1]
	}
	return profileName
}

func getConfig() Config {
	// Get credentials file
	filename := fmt.Sprintf("%s/.aws/credentials", os.Getenv("HOME"))
	credentialsData, err := ioutil.ReadFile(filename)
	handleError(err)

	// Get config file
	filename = fmt.Sprintf("%s/.aws/config", os.Getenv("HOME"))
	configData, err := ioutil.ReadFile(filename)
	handleError(err)

	// Concatenate config files and parse the data
	data := fmt.Sprintf("%s\n%s", credentialsData, configData)
	parser := NewParser()
	return parser.Parse([]byte(data))
}

func main() {
	// Make sure we don't use the wrong aws creds
	for _, env := range os.Environ() {
		if strings.Contains(env, "AWS_") {
			os.Unsetenv(env)
		}
	}

	// DEBUG: Print entire config struct
	// res, _ := json.MarshalIndent(config, "", "  ")
	// fmt.Println(string(res))
	config := getConfig()
	profile := config[getProfileName()]
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile: profile.SourceProfile,
	}))

	// DEBUG: Print parent creds
	// val, _ := sess.Config.Credentials.Get()
	// fmt.Println(val)

	var creds *credentials.Credentials
	if role_arn := profile.RoleArn; len(role_arn) > 0 {
		serial := profile.MfaSerial
		creds = stscreds.NewCredentials(sess, role_arn, func(p *stscreds.AssumeRoleProvider) {
			p.SerialNumber = aws.String(serial)
			p.TokenProvider = stscreds.StdinTokenProvider
		})
	} else {
		creds = credentials.NewStaticCredentials(profile.Id, profile.Secret, "")
	}
	value, err := creds.Get()
	handleError(err)

	tmpl := `
export AWS_DEFAULT_REGION=%s
export AWS_ACCESS_KEY_ID=%s
export AWS_SECRET_ACCESS_KEY=%s
export AWS_SESSION_TOKEN=%s
`
	output := fmt.Sprintf("%s/.aws/env", os.Getenv("HOME"))
	content := fmt.Sprintf(tmpl, profile.Region, value.AccessKeyID, value.SecretAccessKey, value.SessionToken)
	err = ioutil.WriteFile(output, []byte(content), 0600)
	handleError(err)

	fmt.Printf("Env file written to `%s`\n", output)
}
