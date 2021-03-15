package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"
)

const (
	VK_APP_ID  = "7065390"
	VK_APP_KEY = "cQZe3Vvo4mHotmetUdXK"
	// куда идти с токеном за информацией
	VK_API_URL = "https://api.vk.com/method/users.get?fields=photo_50&access_token=%s&v=5.52"
	// куда идти для получения токена
	VK_AUTH_URL = "https://oauth.vk.com/authorize?client_id=7065390&redirect_uri=http://localhost:8080/user/login_oauth&response_type=code&scope=email"
)

type Response struct {
	Response []struct {
		FirstName string `json:"first_name"`
		Photo     string `json:"photo_50"`
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		code := r.FormValue("code")

		if code == "" {
			w.Write([]byte(`<div><a href="` + VK_AUTH_URL + `">authorize</a></div>`))
			return
		}

		conf := oauth2.Config{
			ClientID:     VK_APP_ID,
			ClientSecret: VK_APP_KEY,
			RedirectURL:  "http://localhost:8080/user/login_oauth",
			Endpoint:     vk.Endpoint,
		}

		token, err := conf.Exchange(ctx, code)
		if err != nil {
			http.Error(w, "cannot exchange "+err.Error(), 500)
			return
		}

		email := token.Extra("email").(string)
		userIDraw := token.Extra("user_id").(float64)
		userID := int(userIDraw)

		w.Write([]byte(`
		<div> Oauth token:<br>
			` + fmt.Sprintf("%#v", token) + `
		</div>
		<div>Email: ` + email + `</div>
		<div>UserID: ` + strconv.Itoa(userID) + `</div>
		<br>
		`))

		client := conf.Client(ctx, token)
		resp, err := client.Get(fmt.Sprintf(VK_API_URL, token.AccessToken))
		if err != nil {
			http.Error(w, "request error "+err.Error(), 500)
			return
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "buf read err "+err.Error(), 500)
			return
		}

		data := &Response{}
		json.Unmarshal(body, data)

		w.Write([]byte(`
		<div>
			<img src="` + data.Response[0].Photo + `"/>
			` + data.Response[0].FirstName + `
		</div>
		<br>
		<div> User info:<br>
			` + string(body) + `
		</div>
		`))
	})

	http.ListenAndServe(":8080", nil)
}
