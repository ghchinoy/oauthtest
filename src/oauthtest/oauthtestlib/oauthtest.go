// Package oauthtestlib provides methods for calling API endpoints after
// obtaining OAuth tokens via OAuth2 Client-Credentials Flow
package oauthtestlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"oauthtest/helper"

	"github.com/fatih/color"
)

// MakeAPICalls takes
// TODO threads not used - use goroutine
// TODO algo: review the algorithm used to make number of calls
func MakeAPICalls(config helper.Configuration, debug bool) {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	count := r.Intn(config.Max)
	log.Printf("Choosing %v of max %v calls.", count, config.Max)
	clients := config.Clients

	if debug {
		fmt.Println("Clients", len(clients))
	}
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
			endpoint := config.GenerateRandomURL()
			//log.Printf("Calling %s ... with %s %v", endpoint, v.Name, v.AccessToken)
			err := CallEndpointWithAuthzHeader(endpoint, v.AccessToken.AccessToken, debug)
			if err != nil {
				log.Fatalln("Can't call ", endpoint)
				return
			}
		}
	}
}

// CallEndpointWithAuthzHeader Makes a call to an API Endpoint using the OAuth2 token provided
func CallEndpointWithAuthzHeader(endpoint, token string, debug bool) error {
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

// ObtainOAuthTokens Given a set of OAuth Clients, obtain access tokens
func ObtainOAuthTokens(config helper.Configuration, debug bool) helper.Configuration {
	// Obtain tokens by calling token endpoint for OAuth2 2-legged client credentials
	// A few TODOs here
	// * check to see if a token is present and
	//   that the token haven't expired prior to obtaining another one
	for k, v := range config.Clients {

		// TODO: Incomplete - check to see if access token present, make call if expires is over
		if v.AccessToken != (helper.TokenResponse{}) {
			log.Printf("Access token present, expires: %v", v.AccessToken.ExpiresIn)
		}

		uristr := config.OAuth.BaseURI + config.OAuth.TokenURI + "?grant_type=client_credentials"

		var resp *http.Response
		var err error

		if config.AccessTokenMethod != "post" {
			config.AccessTokenMethod = "get"
		}

		if debug {
			log.Println("Access Token Method", config.AccessTokenMethod)
		}

		if config.AccessTokenMethod == "post" {
			data := url.Values{}
			data.Set("client_secret", v.AppSecret)
			data.Set("client_id", v.AppKey)

			client := &http.Client{}
			req, _ := http.NewRequest("POST", uristr, bytes.NewBufferString(data.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
			resp, err = client.Do(req)
			if err != nil {
				log.Fatalln(err)
				return config
			}
		} else {
			if debug {
				log.Printf("Obtaining token from %s%s", config.OAuth.BaseURI, config.OAuth.TokenURI)
			}
			uristr = uristr + "&client_id=" + v.AppKey +
				"&client_secret=" + v.AppSecret +
				"&scope=" + config.OAuth.Scope
			if debug {
				log.Println(uristr)
			}
			resp, err = http.Get(uristr)
			if err != nil {
				log.Fatalln(err)
				return config
			}
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if debug {
			log.Printf("%s", body)
		}
		var token helper.TokenResponse
		err = json.Unmarshal(body, &token)
		if err != nil {
			log.Fatalln("Can't unmarshal json response from token endpoint", err)
			return config
		}
		info := color.New(color.FgGreen).SprintFunc()
		log.Printf("Token obtained (%s): %s", v.Name, info(token.AccessToken))
		// get the original array, the one in this loop is a copy
		config.Clients[k].AccessToken = token

	}

	return config
}

// parseConfig parses a json file
// No longer used
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
