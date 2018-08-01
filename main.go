package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//TODO Check for VK API errors when parsing JSON

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

func download(url, dest string) error {
	r, err := http.DefaultClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	name := url[strings.LastIndex(url, "/")+1:]
	f, err := os.OpenFile(filepath.Join(dest, name), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	io.Copy(f, r.Body)
	return nil
}

func downloadPhotos(accessToken string, photos *GetPhotosResponse, dest string) error {
	for _, photo := range photos.Response.Items {
		sort.Slice(photo.Sizes, func(i, j int) bool {
			return (photo.Sizes[i].Width + photo.Sizes[i].Height) >
				(photo.Sizes[j].Width + photo.Sizes[j].Height)
		})
		url := photo.Sizes[0].URL
		download(url, dest)
	}
	return nil
}

func prepareDestination(dest string) (string, error) {
	dest, err := filepath.Abs(dest)
	if err != nil {
		return "", err
	}
	stat, err := os.Stat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(dest, 0744); err != nil {
				return "", err
			}
			return dest, nil
		}
		return "", err
	}
	if !stat.IsDir() {
		return "", fmt.Errorf("Destination should be a directory")
	}
	return dest, nil
}

func main() {
	album := flag.String("album", "saved", "Album name")
	dest := flag.String("dest", ".", "Destination folder for saving photos")
	flag.Parse()
	accessToken := os.Getenv("ACCESS_TOKEN")
	if accessToken == "" {
		panic("please set vk.com access token in ACCESS_TOKEN environment variable")
	}
	photos, err := getPhotosList(*album, accessToken)
	if err != nil {
		panic(err)
	}
	absDest, err := prepareDestination(*dest)
	if err != nil {
		panic(err)
	}
	err = downloadPhotos(accessToken, photos, absDest)
	if err != nil {
		panic(err)
	}
}
