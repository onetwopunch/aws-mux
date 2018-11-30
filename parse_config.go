package main

import (
	"regexp"
	"strings"
)

type Config map[string]*Profile

type Profile struct {
	RoleArn       string
	MfaSerial     string
	Id            string
	Secret        string
	SourceProfile string
	Region        string
}

type Parser struct {
	RegexSection  *regexp.Regexp
	RegexKeyValue *regexp.Regexp
	RegexAccount  *regexp.Regexp
}

func NewParser() *Parser {
	parser := &Parser{}
	parser.RegexAccount = regexp.MustCompile(`\[([a-z0-9_-]+)\]`)
	parser.RegexSection = regexp.MustCompile(`\[profile ([a-z0-9_-]+)\]`)
	parser.RegexKeyValue = regexp.MustCompile(`([a-z0-9_-]+) = (.+?)$`)
	return parser
}

func (this *Parser) Parse(data []byte) Config {
	result := make(Config)
	lines := strings.Split(string(data), "\n")
	var currentSection string
	var currentProfile *Profile

	for _, line := range lines {
		if this.RegexAccount.MatchString(line) {
			match := this.RegexAccount.FindStringSubmatch(line)
			currentSection = match[1]
			currentProfile = &Profile{SourceProfile: currentSection}
			result[currentSection] = currentProfile
		} else if this.RegexSection.MatchString(line) {
			match := this.RegexSection.FindStringSubmatch(line)
			currentSection = match[1]
			currentProfile = &Profile{SourceProfile: "default"}
			result[currentSection] = currentProfile
		} else if this.RegexKeyValue.MatchString(line) {
			match := this.RegexKeyValue.FindStringSubmatch(line)
			key := match[1]
			value := match[2]
			switch key {
			case "role_arn":
				currentProfile.RoleArn = value
			case "mfa_serial":
				currentProfile.MfaSerial = value
			case "aws_access_key_id":
				currentProfile.Id = value
			case "aws_secret_access_key":
				currentProfile.Secret = value
			case "source_profile":
				currentProfile.SourceProfile = value
			case "region":
				currentProfile.Region = value
			}
		}
	}
	return result
}
