package main

import (
  "os"
  "log"
  "fmt"
  "io/ioutil"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/aws/credentials"
  "github.com/aws/aws-sdk-go/aws/credentials/stscreds"
)
func handleError(err error) {
  if err != nil {
    log.Fatal(err)
  }
}
func main() {
  profileName := os.Args[1]

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
  config := parser.Parse([]byte(data))

  // DEBUG: Print entire config struct
  // res, _ := json.MarshalIndent(config, "", "  ")
  // fmt.Println(string(res))
  profile := config[profileName]
  sess := session.Must(session.NewSessionWithOptions(session.Options{
  	 Profile: profile.SourceProfile,
  }))

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
