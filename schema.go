package main

// GetPhotosResponse describes VK API response for photos.get
type GetPhotosResponse struct {
	Response struct {
		Count int `json:"count"`
		Items []struct {
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
		} `json:"items"`
	} `json:"response"`
}
