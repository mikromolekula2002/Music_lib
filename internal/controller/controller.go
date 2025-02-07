package controller

import (
	"database/sql"
	"errors"
	"mikromolekula2002/music_library_ver1.0/internal/models"
	"mikromolekula2002/music_library_ver1.0/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type MusicLibController struct {
	service *service.MusicLibService
}

func NewMusicLibController(service *service.MusicLibService) *MusicLibController {
	return &MusicLibController{service: service}
}

// @Summary Save song data
// @Description Save group and song from the request and fetch additional text from an external API
// @Tags sav song
// @Accept json
// @Produce json
// @Param song body models.CreateSongReq true "Song data"
// @Success 201 {object} models.CreateSongReq "Song successfully saved"
// @Failure 400 {object} models.ErrorResponse "Invalid request format"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /create-song [post]
func (m *MusicLibController) SaveSong(ctx *gin.Context) {
	m.service.Logger.Info("Handling request", logrus.Fields{
		"method": ctx.Request.Method,
		"url":    ctx.Request.URL.String(),
	})

	var song models.CreateSongReq
	var songResp models.CreateSongResp

	if err := ctx.ShouldBindJSON(&song); err != nil {
		m.service.Logger.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	m.service.Logger.Debug("save song with parameters:", logrus.Fields{
		"group": song.Group,
		"song":  song.Song,
	})

	songData, err := m.service.GetSongDetailsFromAPI(song.Group, song.Song, ctx)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	if err := m.service.SaveSong(songData); err != nil {
		ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	m.service.Logger.Info("Song successfully saved", logrus.Fields{
		"id":    songData.ID,
		"group": songData.Group,
		"song":  songData.Song,
	})

	songResp.ID = songData.ID
	songResp.Group = song.Group
	songResp.Song = song.Song

	ctx.JSON(201, songResp)
}

// @Summary Get song text by group and song name
// @Description Fetches the lyrics of a song from a specific group with pagination.
// @Tags song text
// @Accept json
// @Produce json
// @Param group query string true "Group name"
// @Param song query string true "Song name"
// @Param limit query int false "Number of lines to return" default(10)
// @Param offset query int false "Offset from the beginning" default(0)
// @Success 200 {object} models.SongTextResponse "Successful response"
// @Failure 400 {object} models.ErrorResponse "Bad Request: Invalid parameters"
// @Failure 404 {object} models.ErrorResponse "Not Found: Song text not found"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /song [get]
func (m *MusicLibController) GetSongTextByGroup(ctx *gin.Context) {
	m.service.Logger.Info("Handling request", logrus.Fields{
		"method": ctx.Request.Method,
		"url":    ctx.Request.URL.String(),
	})

	limit, offset, err := parseLimitOffset(ctx)
	if err != nil {
		m.service.Logger.Error("GetSongTextByGroup: ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	groupName := ctx.Query("group")
	songName := ctx.Query("song")
	if groupName == "" || songName == "" {
		m.service.Logger.Error("GetSongTextByGroup: invalid parameters")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameters: group or song"})
		return
	}

	m.service.Logger.Debug("get song text by group with parameters:", logrus.Fields{
		"group": groupName,
		"song":  songName,
	})

	song, err := m.service.GetSongTextByGroup(groupName, songName, limit, offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(404, gin.H{"error": "Song text not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	ctx.JSON(200, song)
}

// @Summary Update an existing song
// @Description Updates a song's details such as release date, text, and link.
// @Tags update song
// @Accept json
// @Produce json
// @Param song body models.Song true "Song data to update"
// @Success 200 {object} models.ErrorResponse "Song updated successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 404 {object} models.ErrorResponse "Song not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /song [put]
func (m *MusicLibController) UpdateSong(ctx *gin.Context) {
	m.service.Logger.Info("Handling request", logrus.Fields{
		"method": ctx.Request.Method,
		"url":    ctx.Request.URL.String(),
	})

	var song models.Song
	if err := ctx.ShouldBindJSON(&song); err != nil {
		m.service.Logger.Error("UpdateSong: invalid parameters")
		ctx.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	m.service.Logger.Debug("update song with parameters:", logrus.Fields{
		"group":        song.Group,
		"song":         song.Song,
		"release_date": song.ReleaseDate,
		"text":         song.Text,
		"link":         song.Link,
	})

	if song.ReleaseDate != "" && !m.service.IsValidDate(song.ReleaseDate) {
		m.service.Logger.Error("GetSongTextByGroup: invalid parameter release date")
		ctx.JSON(400, gin.H{"error": "Invalid request body"})
	}

	if err := m.service.UpdateSong(&song); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(404, gin.H{"error": "Song not found"})
			return
		}
		ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	m.service.Logger.Info("Song data updated successfully", logrus.Fields{
		"group": song.Group,
		"song":  song.Song,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "Song updated successfully"})
}

// @Summary Delete a song by group and song name
// @Description Deletes a song from the library based on the provided group and song name.
// @Tags delete song
// @Accept json
// @Produce json
// @Param group query string true "Group name"
// @Param song query string true "Song name"
// @Success 200 {object} models.ErrorResponse "Song deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Bad Request: Missing required parameters"
// @Failure 404 {object} models.ErrorResponse "Not Found: Song not found"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /song [delete]
func (m *MusicLibController) DeleteSong(ctx *gin.Context) {
	m.service.Logger.Info("Handling request", logrus.Fields{
		"method": ctx.Request.Method,
		"url":    ctx.Request.URL.String(),
	})

	groupName := ctx.Query("group")
	songName := ctx.Query("song")
	if groupName == "" || songName == "" {
		ctx.JSON(400, gin.H{"error": "Missing required parameter: group or song"})
		return
	}

	m.service.Logger.Debug("delete song with parameters:", logrus.Fields{
		"group": groupName,
		"song":  songName,
	})

	if err := m.service.DeleteSong(groupName, songName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(404, gin.H{"error": "Song not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	ctx.JSON(200, gin.H{"message": "Song deleted successfully"})
}

// @Summary Get all songs with optional filters
// @Description Retrieves a list of songs based on optional filters, with pagination support.
// @Tags songs
// @Accept json
// @Produce json
// @Param group query string false "Filter by group name"
// @Param song query string false "Filter by song name"
// @Param link query string false "Filter by link"
// @Param releaseDate query string false "Filter by exact release date (YYYY-MM-DD)"
// @Param startDate query string false "Filter by release date range start (YYYY-MM-DD)"
// @Param endDate query string false "Filter by release date range end (YYYY-MM-DD)"
// @Param limit query int false "Number of results to return" default(10)
// @Param offset query int false "Offset from the beginning" default(0)
// @Success 200 {array} models.Song "Successful response with list of songs"
// @Failure 400 {object} models.ErrorResponse "Bad Request: Invalid parameters"
// @Failure 404 {object} models.ErrorResponse "Not Found: No songs found"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /songs [get]
func (m *MusicLibController) GetAllSongs(ctx *gin.Context) {
	m.service.Logger.Info("Handling request", logrus.Fields{
		"method": ctx.Request.Method,
		"url":    ctx.Request.URL.String(),
	})

	limit, offset, err := parseLimitOffset(ctx)
	if err != nil {
		m.service.Logger.Error("GetAllSongs: ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := make(map[string]string)
	group := ctx.Query("group")
	if group != "" {
		filter["group_name"] = group
	}
	song := ctx.Query("song")
	if song != "" {
		filter["song"] = song
	}
	link := ctx.Query("link")
	if link != "" {
		filter["link"] = link
	}
	releaseDate := ctx.Query("releaseDate")
	if releaseDate != "" {
		filter["releaseDate"] = releaseDate
	}
	startDate := ctx.Query("startDate")
	if startDate != "" {
		filter["startDate"] = startDate
	}
	endDate := ctx.Query("endDate")
	if endDate != "" {
		filter["endDate"] = endDate
	}

	m.service.Logger.Debug("get songs with parameters:", logrus.Fields{
		"limit":        limit,
		"offset":       offset,
		"group":        group,
		"song":         song,
		"release date": releaseDate,
		"start date":   startDate,
		"end date":     endDate,
	})

	songs, err := m.service.GetAllSongs(filter, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	if len(songs) == 0 {
		ctx.JSON(404, gin.H{"error": "No songs found"})
		return
	}

	ctx.JSON(200, songs)
}

func parseLimitOffset(ctx *gin.Context) (int, int, error) {
	limitStr := ctx.DefaultQuery("limit", "15")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		return 0, 0, errors.New("invalid 'limit' parameter")
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		return 0, 0, errors.New("invalid 'offset' parameter")
	}

	return limit, offset, nil
}
