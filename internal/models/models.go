package models

type Song struct {
	ID          uint   `json:"id" binding:"omitempty"`
	Group       string `json:"group" binding:"required"`
	Song        string `json:"song" binding:"required"`
	ReleaseDate string `json:"release_date" binding:"omitempty"`
	Text        string `json:"text" binding:"omitempty"`
	Link        string `json:"link" binding:"omitempty"`
}

type SongTextResp struct {
	ID    uint     `json:"id" binding:"omitempty"`
	Group string   `json:"group"`
	Song  string   `json:"song"`
	Text  []string `json:"text"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

//Форма для БД

type CreateSongReq struct {
	Group string `json:"group" binding:"required"`
	Song  string `json:"song" binding:"required"`
}

type CreateSongResp struct {
	ID    uint   `json:"id" binding:"required"`
	Group string `json:"group" binding:"required"`
	Song  string `json:"song" binding:"required"`
}

type SongDetailResp struct {
	ReleaseDate string `json:"release_date"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

type SongTextResponse struct {
	Group string   `json:"group"`
	Song  string   `json:"song"`
	Text  []string `json:"text"`
}
