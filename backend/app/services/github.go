package services

import (
	"log"
	"net/url"
)

func BuildUrl(clientid, redirecturl, state string) (string, error) {
	BaseUrl, err := url.Parse("https://github.com/login/oauth/authorize")
	if err != nil {
		log.Fatal(err)
	}
	params := url.Values{}
	params.Add("client_id", clientid)
	params.Add("redirect_uri", redirecturl)
	params.Add("scope", "read:user user:mail")
	params.Add("state", state)
	BaseUrl.RawQuery = params.Encode()
	return BaseUrl.String(), nil
}
