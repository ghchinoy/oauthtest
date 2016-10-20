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

	"github.com/fatih/color"
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
	servererr := color.New(color.FgRed).SprintFunc()
	if resp.StatusCode != 200 {
		log.Printf("%s %v %s\n", servererr(resp.Status), len(body), endpoint)
	} else {
		log.Printf("%s %v %s\n", resp.Status, len(body), endpoint)
	}

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
		info := color.New(color.FgGreen).SprintFunc()
		log.Printf("Token obtained (%s): %s", v.Name, info(token.AccessToken))
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
	// max is 0
	maxflag := flag.Int("max", 0, "Maximum amount API calls to make")
	threadsflag := flag.Int("threads", 0, "Maximum amount of threads to use")
	configflag := flag.String("config", "", "json config file to use")
	debugflag := flag.Bool("debug", false, "show debug statements")
	flag.Parse()
	configfile = string(*configflag)
	debug = bool(*debugflag)
	maxcalls = int(*maxflag)
	threads = int(*threadsflag)

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
		// if command-line max or threads flag > 0, ignore configuration file
		if maxcalls == 0 {
			maxcalls = configuration.Max
		}
		if threads == 0 {
			threads = configuration.Threads
		}
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
	} else {
		fmt.Println("Can't find config file. Please use the --config flag to specify the location of a configuration json file.")
		flag.Usage()
		os.Exit(1)
	}

	// set defaults
	if maxcalls == 0 {
		maxcalls = 20
	}
	if threads == 0 {
		threads = 2
	}

}

// TODO clarify code, then incorporate into apipong
func main() {

	obtainOAuthTokens(configuration)
	makeAPICalls(maxcalls, threads)

}
