# OAuth Client-Credentials test

This small program calls API endpoints after obtaining an OAuth2 access token via the OAuth2 2-Legged Client-Credentials Grant Type Flow.

## Types of OAuth Grant Types

* AC: 3-Legged Authorization Code
* I: 3-Legged Implicit
* *CC: 2-Legged Client Credentials* - app implements this
* ROC: 2-Legged Resource Owner Credentials

## Usage of oauthtest

flags:

* `max` - optional, max amt of calls to make, default `20`
* `threads` - optional, threads to use, default `2` (not implemented at this time)
* `config` - optional path to config file, defaults to `$HOME/.akana/oauthtest.json`
* `debug` - optional, output debug statements

### Example usage

```
bin/oauthtest --config config/local.json
```

## Configuration file

_Please note, the config file format has changed slightly between v.1.x and v.2.x, with the major change that a single file can now refer to multiple profiles._

Example of configuration file, below, containing two profiles, `default` and `saas`.

The default path for a config file is `$HOME/.akana/oauthtest.json`


```
{
    "default":
        {
            "uri": "https://nd.akana.dev:9982/v4/geocode/{endpoint}?address={address}",
            "clients": [
                {"name":"Client 1",
                "appkey":"enterpriseapi-AnBefYqhBHF76Onxl6CLjD5z",
                "appsecret":"cbaaac6e80d07961612971fd257bfd27f8954c70"},
                {"name":"Client 2",
                "appkey":"enterpriseapi-6XWQII2hKOJAvbGzsqdxjy7E",
                "appsecret":"d5be0d59c2a5abef7fb13711c908ccd900339b06"},
            ],
            "substitutions": [
                {
                    "name": "address",
                    "array": [
                        "Boulder, CO",
                        "Fort Collins, CO",
                    ]
                },
                {
                    "name": "endpoint",
                    "array": [
                        "city",
                        "address"
                    ]
                }
            ],
            "threads": 2,
            "max": 10,
            "scope": "Public",
            "oauth": {
            "baseuri": "http://portal.akana.dev:9980",
            "tokenuri": "/oauth/oauth20/token"
            }
        },
    "saas": {
        ...
    }
}
```

Sample config files are in the `config` directory.

## Known Issues

* `threads` flag not used at this time - threading not implemented.


## Build and Install

Written in [Go](https://golang.org/) and uses [gb](https://getgb.io/) to build. 

To get gb, with golang already installed:

```
go get github.com/constabulary/gb/...
```

To build, be in this directory, `oauthtest`:

```
gb build
```

This will create the `oauthtest` binary in the bin dir (`oauthtest/bin`).
