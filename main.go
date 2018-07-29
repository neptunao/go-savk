package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
)

// APIVersion is vk.com API version
const APIVersion = "5.80"

func getPhotosList(album string, accessToken string) (*GetPhotosResponse, error) {
	client := http.DefaultClient
	r, err := client.Get(fmt.Sprintf("https://api.vk.com/method/photos.get?access_token=%s&v=%s&album_id=%s",
		accessToken, APIVersion, album))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("photos.get returned %d : %s", r.StatusCode, r.Status)
	}
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

func download(url string) error {
	r, err := http.DefaultClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	name := url[strings.LastIndex(url, "/")+1:]
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	io.Copy(f, r.Body)
	return nil
}

func downloadPhotos(accessToken string, photos *GetPhotosResponse) error {
	for _, photo := range photos.Response.Items {
		sort.Slice(photo.Sizes, func(i, j int) bool {
			return (photo.Sizes[i].Width + photo.Sizes[i].Height) >
				(photo.Sizes[j].Width + photo.Sizes[j].Height)
		})
		url := photo.Sizes[0].URL
		download(url)
	}
	return nil
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
	err = downloadPhotos(accessToken, photos)
	if err != nil {
		panic(err)
	}
}
