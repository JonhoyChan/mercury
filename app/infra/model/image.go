package model

type ProfileImage struct {
	Tiny     string `json:"tiny"`
	Small    string `json:"small"`
	Medium   string `json:"medium"`
	Large    string `json:"large"`
	Original string `json:"original"`
}

type ProfileVideo struct {
	Original string `json:"original"`
}
