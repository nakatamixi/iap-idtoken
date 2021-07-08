package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"

	"io/ioutil"

	"github.com/kelseyhightower/envconfig"
	"github.com/manifoldco/promptui"
)

type Config struct {
	OauthClientID     string `required:"true" split_words:"true"`
	OauthClientSecret string `required:"true" split_words:"true"`
	IapOauthClientID  string `required:"true" split_words:"true"`
}

// print id_token of https://cloud.google.com/iap/docs/authentication-howto?hl=ja#authenticating_from_a_desktop_app
func main() {

	conf := MustGetConfig()
	authURL := fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&response_type=code&scope=openid%%20email&access_type=offline&redirect_uri=urn:ietf:wg:oauth:2.0:oob", conf.OauthClientID)
	if _, err := exec.LookPath("open"); err == nil {
		_, err := exec.Command("open", authURL).Output()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Sprintf("please access this url from user browser.\n%s\n", authURL)
	}
	authCode := MustPromptStr("paste AUTH_CODE")
	refreshToken := MustGetRefreshToken(conf, authCode)
	fmt.Print(MustGetIDToken(conf, refreshToken))
}

func MustPromptStr(l string) string {
	prompt := promptui.Prompt{
		Label: l,
	}
	result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}
	return result
}
func MustGetConfig() Config {
	var conf Config
	if err := envconfig.Process("", &conf); err != nil {
		log.Fatal(err)
	}
	return conf
}

func MustGetRefreshToken(conf Config, authCode string) string {
	resp, err := http.PostForm("https://oauth2.googleapis.com/token",
		url.Values{
			"client_id":     {conf.OauthClientID},
			"client_secret": {conf.OauthClientSecret},
			"code":          {authCode},
			"redirect_uri":  {"urn:ietf:wg:oauth:2.0:oob"},
			"grant_type":    {"authorization_code"},
		})
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var j struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(body, &j); err != nil {
		log.Fatalln("error:", err)
	}
	return j.RefreshToken
}

func MustGetIDToken(conf Config, refreshToken string) string {
	resp, err := http.PostForm("https://oauth2.googleapis.com/token",
		url.Values{
			"client_id":     {conf.OauthClientID},
			"client_secret": {conf.OauthClientSecret},
			"refresh_token": {refreshToken},
			"audience":      {conf.IapOauthClientID},
			"grant_type":    {"refresh_token"},
		})
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var j struct {
		IdToken string `json:"id_token"`
	}
	if err := json.Unmarshal(body, &j); err != nil {
		log.Fatalln("error:", err)
	}
	return j.IdToken
}
