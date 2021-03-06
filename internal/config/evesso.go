package config

import (
	"net/url"
	"os"
)

const appPrefix = "EVE_VULCAN_"
const ssoPrefix = appPrefix + "EVE_SSO_"

func EveSSODomainName() string {
	return os.Getenv(ssoPrefix + "DOMAIN_NAME")
}

type ClientAuth struct {
	ID     string
	Secret string
}

func EveSSOClientAuth() ClientAuth {
	return ClientAuth{
		ID:     os.Getenv(ssoPrefix + "CLIENT_ID"),
		Secret: os.Getenv(ssoPrefix + "CLIENT_SECRET"),
	}
}

func EveSSORedirectURL() *url.URL {
	redirectURL, err := url.Parse(os.Getenv(ssoPrefix + "REDIRECT_URL"))
	if err != nil {
		panic(err)
	}
	return redirectURL
}
