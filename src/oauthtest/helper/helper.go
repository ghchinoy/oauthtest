package helper

import (
	"math/rand"
	"net/url"
	"regexp"
	"time"
)

// Configuration is the struct for the config file
type Configuration struct {
	URI           string         `json:"uri"`
	Clients       []Client       `json:"clients"`
	Max           int            `json:"max"`
	Threads       int            `json:"threads"`
	Substitutions []Substitution `json:"substitutions"`
	OAuth         OAuth          `json:"oauth"`
}

// Substitution is used within the configuration structure
type Substitution struct {
	Name  string   `json:"name"`
	Array []string `json:"array"`
}

type OAuth struct {
	BaseURI  string `json:"baseuri"`
	TokenURI string `json:"tokenuri"`
	Scope    string `json:"scope"`
}

// Client is a data structure to represent an OAuth Client.
// Also used in configuration struct / file.
type Client struct {
	Name        string `json:"name"`
	AppKey      string `json:"appkey"`
	AppSecret   string `json:"appsecret"`
	AccessToken TokenResponse
}

// TokenResponse represents the OAuth 2 token response object
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (config Configuration) GenerateRandomURL() string {

	rand.Seed(time.Now().UTC().UnixNano())

	re, _ := regexp.Compile("({([a-zA-Z]+)})")

	vars := re.FindAllStringSubmatch(config.URI, -1)

	// Random URL
	randomURL := config.URI
	for q := 0; q < len(vars); q++ {
		param := vars[q][2]

		// get the appropriate Substitution
		var values []string
		for _, v := range config.Substitutions {
			if v.Name == param {
				values = v.Array
			}
		}
		value := url.QueryEscape(values[rand.Intn(len(values))])

		exp := regexp.MustCompile("{" + param + "}")
		randomURL = exp.ReplaceAllLiteralString(randomURL, value)
	}
	//fmt.Printf("%s\n", randomURL)

	return randomURL
}
