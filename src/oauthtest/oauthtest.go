// Package main provides a simple sketch of calling API endpoints after
// obtaining OAuth tokens via OAuth2 Client-Credentials Flow
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"oauthtest/helper"
)

// makeAPICalls takes
// TODO threads not used - use goroutine
// TODO algo: review the algorithm used to make number of calls
func makeAPICalls(maxcalls, threads int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	count := r.Intn(maxcalls)
	log.Printf("Choosing %v of max %v calls.", count, maxcalls)
	clients := configuration.Clients

	// TODO algorithm change: this always makes calls to all clients.
	// - may need a different strategy, i.e. randomly choose one
	// or, given a total number of calls, random between clients and endpoints

	for i := 0; i < count; i++ {
		log.Printf("Client: %s", clients[r.Intn(len(clients))].Name)
		for _, v := range clients {
			// Make a call to an API endpoint
			if v.AccessToken == (helper.TokenResponse{}) {
				log.Printf("No access token for '%s'.", v.Name)
				return
			}
			// random endpoint
			endpoint := configuration.GenerateRandomURL()
			//log.Printf("Calling %s ... with %s %v", endpoint, v.Name, v.AccessToken)
			err := callEndpointWithAuthzHeader(endpoint, v.AccessToken.AccessToken)
			if err != nil {
				log.Fatalln("Can't call ", endpoint)
				return
			}
		}
	}
}

// Make a call to an API Endpoint using the OAuth2 token provided
func callEndpointWithAuthzHeader(endpoint, token string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", endpoint, nil)
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/json")
	if debug {
		log.Println("URL:", req.URL)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln("Can't GET", err)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	log.Printf("%s %v %s\n", resp.Status, len(body), endpoint)
	//log.Printf("%s", body)
	return nil
}

// Given a set of OAuth Clients, obtain access tokens
func obtainOAuthTokens(config helper.Configuration) {
	// Obtain tokens by calling token endpoint for OAuth2 2-legged client credentials
	// A few TODOs here
	// * check to see if a token is present and
	//   that the token haven't expired prior to obtaining another one
	for k, v := range config.Clients {

		// TODO: Incomplete - check to see if access token present, make call if expires is over
		if v.AccessToken != (helper.TokenResponse{}) {
			log.Printf("Access token present, expires: %v", v.AccessToken.ExpiresIn)
		}

		//log.Printf("Obtaining token from %s%s", config.OAuth.BaseURI, config.OAuth.TokenURI)
		resp, err := http.Get(config.OAuth.BaseURI + config.OAuth.TokenURI +
			"?client_id=" + v.AppKey +
			"&client_secret=" + v.AppSecret +
			"&scope=" + config.OAuth.Scope +
			"&grant_type=client_credentials")
		if err != nil {
			log.Fatalln(err)
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		//log.Printf("%s", body)
		var token helper.TokenResponse
		err = json.Unmarshal(body, &token)
		if err != nil {
			log.Fatalln("Can't unmarshal json response from token endpoint", err)
			return
		}
		log.Printf("Token obtained (%s): %s", v.Name, token.AccessToken)
		// get the original array, the one in this loop is a copy
		config.Clients[k].AccessToken = token

	}
}

// control and api
var (
	apiBaseURI string
	//apis          []string
	//locations     []string
	//scope         string
	maxcalls      int
	threads       int
	clients       []helper.Client
	configfile    string
	configuration helper.Configuration
	debug         bool
)

// parseConfig parses a json file
func parseConfig(configfile string) (helper.Configuration, error) {

	config := helper.Configuration{}

	configBytes, err := ioutil.ReadFile(configfile)
	if err != nil {
		fmt.Printf("Error opening config file '%s'.\n", err)
		return config, err
	}

	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		fmt.Printf("Unable to parse the configuration file '%s'.\n", err)
		return config, err
	}

	return config, nil
}

func init() {

	//helpflag := flag.String("help", "", "Display help info.")
	maxflag := flag.Int("max", 20, "Maximum amount API calls to make")
	threadsflag := flag.Int("threads", 2, "Maximum amount of threads to use")
	configflag := flag.String("config", "", "json config file to use")
	debugflag := flag.Bool("debug", false, "show debug statements")
	flag.Parse()
	maxcalls = int(*maxflag)
	threads = int(*threadsflag)
	configfile = string(*configflag)
	debug = bool(*debugflag)

	/*
		// Defaults
		// API Info
		apiBaseURI = "https://nd.akana.dev:9982"
		apis = []string{
			"/v4/geocode/city?address=",
			"/v4/geocode/address?address=",
		}
		locations = []string{"Boulder, CO", "Fort Collins, CO", "Chicago, IL", "Los Angeles, CA", "Ventura, CA", "Santa Monica, CA"}
		scope = "Public"
		//maxcalls = 20
		//threads = 2

		// Client Info
		clients = []helper.Client{
			{Name: "License-restricted, T3 Gold", AppKey: "enterpriseapi-AnBefYqhBHF76Onxl6CLjD5z", AppSecret: "cbaaac6e80d07961612971fd257bfd27f8954c70"},
			{Name: "License-restricted, T2 Silver", AppKey: "enterpriseapi-6XWQII2hKOJAvbGzsqdxjy7E", AppSecret: "d5be0d59c2a5abef7fb13711c908ccd900339b06"},
			{Name: "License-restricted, T1 Bronze", AppKey: "enterpriseapi-4yPiqABBNwO59QssNRjyYJ7C", AppSecret: "4f6ac688fdfb10fabc4f12b77fa536bf276efd96"},
			{Name: "Internal Tester", AppKey: "enterpriseapi-1eVguH9dlsb7rKEnQ1Baof7R", AppSecret: "06c6dc757e0a600513f6ee11c4bf142532b58d62"},
		}
	*/

	// Config file, override defaults
	// if configfile exists and is parsable, override defaults with config info
	// determine existence, open, parse
	if len(configfile) > 0 {
		var err error
		configuration, err = parseConfig(configfile)
		if err != nil {
			log.Println(err)
			flag.Usage()
			os.Exit(1)
		}
		maxcalls = configuration.Max
		threads = configuration.Threads
		if debug {
			log.Println("Config", configuration)
		}

		// DEBUG starts here
		/*
			configbytes, _ := json.Marshal(configuration)
			fmt.Printf("%s\n", string(configbytes))
			fmt.Printf("Max %v, Threads %v\n", maxcalls, threads)
			for i := 0; i < 10; i++ {
				fmt.Printf("%s\n", configuration.GenerateRandomURL())
			}
			os.Exit(1)
		*/
		// DEBUG ends here
	}
}

// TODO clarify code, then incorporate into apipong
func main() {

	obtainOAuthTokens(configuration)
	makeAPICalls(maxcalls, threads)

}
