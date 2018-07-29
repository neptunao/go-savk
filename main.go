package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const APIVersion = "5.80"

func getPhotosList(album string, accessToken string) (*GetPhotosResponse, error) {
	client := http.DefaultClient
	r, err := client.Get(fmt.Sprintf("https://api.vk.com/method/photos.get?access_token=%s&v=%s&album_id=%s",
		accessToken, APIVersion, album))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var photos GetPhotosResponse
	if err = json.Unmarshal(data, &photos); err != nil {
		return nil, err
	}
	return &photos, nil
}

func main() {
	album := flag.String("album", "saved", "Album name")
	accessToken := os.Getenv("ACCESS_TOKEN")
	if accessToken == "" {
		panic("please set vk.com access token in ACCESS_TOKEN environment variable")
	}
	photos, err := getPhotosList(*album, accessToken)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v", photos.Response.Items)
}
