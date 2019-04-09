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
	"time"
)

//TODO Check for VK API errors when parsing JSON

// APIVersion is vk.com API version
const APIVersion = "5.92"
const retryLimit = 5
const defaultWaitTime = 5 * time.Second
const waitFactor = 3

// TODO: move to args/envvars
const ownerID = 59233038

var batchSize int

func apiRequest(url string, out interface{}) error {
	r, err := http.DefaultClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &out); err != nil {
		return err
	}
	return nil
}

func getPhotosList(album string, accessToken string, offset int) (*GetPhotosResponse, error) {
	url := fmt.Sprintf("https://api.vk.com/method/photos.get?access_token=%s&v=%s&album_id=%s&owner_id=%d&count=%d&offset=%d",
		accessToken, APIVersion, album, ownerID, batchSize, offset)
	var photos GetPhotosResponse
	if err := apiRequest(url, &photos); err != nil {
		return nil, err
	}
	return &photos, nil
}

func deletePhoto(accessToken string, photo Photo) error {
	url := fmt.Sprintf("https://api.vk.com/method/photos.delete?access_token=%s&v=%s&owner_id=%d&photo_id=%d",
		accessToken, APIVersion, photo.OwnerID, photo.ID)
	var parsedReponse DeletePhotoResponse
	if err := apiRequest(url, &parsedReponse); err != nil {
		return err
	}
	if parsedReponse.Response != 1 {
		return fmt.Errorf("photos.delete returned invaild result (not 0)")
	}
	return nil
}

func download(url, dest string) error {
	r, err := http.DefaultClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	f, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	io.Copy(f, r.Body)
	return nil
}

func downloadPhotos(accessToken string, photos *GetPhotosResponse, dest string) error {
	for i, photo := range photos.Response.Items {
		sort.Slice(photo.Sizes, func(i, j int) bool {
			return (photo.Sizes[i].Width + photo.Sizes[i].Height) >
				(photo.Sizes[j].Width + photo.Sizes[j].Height)
		})
		url := photo.Sizes[0].URL
		name := url[strings.LastIndex(url, "/")+1:]
		destFilename := filepath.Join(dest, name)
		fmt.Printf("[%d/%d] Downloading photo %s to %s\n", i+1,
			len(photos.Response.Items), url, destFilename)
		download(url, destFilename)
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
			if err = os.MkdirAll(dest, 0755); err != nil {
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

func deletePhotos(accessToken string, photos []Photo) error {
	errCount := 0
	count := len(photos)
	waitTime := defaultWaitTime
	if count == 0 {
		return nil
	}
	for i, photo := range photos {
		// VK API allows only 3 req/sec
		if (i+1)%3 == 0 {
			time.Sleep(1 * time.Second)
		}
		fmt.Printf("[%d/%d] deleting photo %d\n", i+1, count, photo.ID)
		if err := deletePhoto(accessToken, photo); err != nil {
			errCount++
			if errCount > retryLimit {
				return err
			}
			fmt.Printf("error %s when deleting photo %d, going to retry after %s\n",
				err, photo.ID, waitTime)
			time.Sleep(waitTime)
			waitTime *= waitFactor
		}
	}
	return nil
}

func main() {
	album := flag.String("album", "saved", "VK album name")
	dest := flag.String("dest", ".", "Destination folder for saved photos")
	dryRun := flag.Bool("dry-run", true, "If false deletes photos after save")
	batchSizeFlag := flag.Int("count", 10, "Batch size of photos processed in one step")
	flag.Parse()
	accessToken := os.Getenv("ACCESS_TOKEN")
	if accessToken == "" {
		panic("please set vk.com access token in ACCESS_TOKEN environment variable")
	}
	batchSize = *batchSizeFlag
	//TODO implement cycle because API returns only up to 1000 items per request
	processed := 0
	count := 1
	for processed < count {
		photos, err := getPhotosList(*album, accessToken, processed)
		if err != nil {
			panic(err)
		}
		count = photos.Response.Count
		absDest, err := prepareDestination(*dest)
		if err != nil {
			panic(err)
		}
		err = downloadPhotos(accessToken, photos, absDest)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%d photos have been saved successfully\n", len(photos.Response.Items))
		if !*dryRun {
			err = deletePhotos(accessToken, photos.Response.Items)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%d photos have been deleted successfully\n", len(photos.Response.Items))
		}
		processed += len(photos.Response.Items)
	}
}
