package main

// GetPhotosResponse describes VK API response for photos.get
type GetPhotosResponse struct {
	Response struct {
		Count int     `json:"count"`
		Items []Photo `json:"items"`
	} `json:"response"`
}

// DeletePhotoResponse describes VK API response for photos.delete
type DeletePhotoResponse struct {
	Response int `json:"response"`
}

// Photo is a mapping for VK API Photo object
// https://vk.com/dev/objects/photo
type Photo struct {
	ID      int64 `json:"id"`
	AlbumID int64 `json:"album_id"`
	OwnerID int64 `json:"owner_id"`
	Sizes   []struct {
		Type   string `json:"type"`
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"sizes"`
	Text string `json:"text"`
	Date int    `json:"date"`
}
