package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"mikromolekula2002/music_library_ver1.0/internal/models"
)

func (r *Repository) SaveSongInfo(group, song, releaseDate, link string) (int, error) {
	op := "repository.SaveSongInfo"

	var id int
	query := `INSERT INTO song_info (group_name, song, release_date, link) VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.db.QueryRow(query, group, song, releaseDate, link).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %v", op, err)
	}
	return id, nil
}

func (r *Repository) SaveSongText(songID uint, verse string) error {
	op := "repository.SaveSongText"

	query := `INSERT INTO song_text (song_id, verse) VALUES ($1, $2)`
	_, err := r.db.Exec(query, songID, verse)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	return nil
}

func (r *Repository) GetSongs(filter map[string]string, limit, offset int) ([]models.Song, error) {
	op := "repository.GetSongs"

	query := `SELECT id, group_name, song, release_date, link FROM song_info WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	if group, ok := filter["group_name"]; ok && group != "" {
		query += fmt.Sprintf(" AND group_name = $%d", argIndex)
		args = append(args, group)
		argIndex++
	}
	if song, ok := filter["song"]; ok && song != "" {
		query += fmt.Sprintf(" AND song = $%d", argIndex)
		args = append(args, song)
		argIndex++
	}
	if link, ok := filter["link"]; ok && link != "" {
		query += fmt.Sprintf(" AND link = $%d", argIndex)
		args = append(args, link)
		argIndex++
	}
	if releaseDate, ok := filter["releaseDate"]; ok && releaseDate != "" {
		query += fmt.Sprintf(" AND release_date = $%d", argIndex)
		args = append(args, releaseDate)
		argIndex++
	}
	if startDate, ok := filter["startDate"]; ok && startDate != "" {
		query += fmt.Sprintf(" AND release_date >= $%d", argIndex)
		args = append(args, startDate)
		argIndex++
	}
	if endDate, ok := filter["endDate"]; ok && endDate != "" {
		query += fmt.Sprintf(" AND release_date <= $%d", argIndex)
		args = append(args, endDate)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY release_date DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}
	defer rows.Close()

	var songs []models.Song
	for rows.Next() {
		var song models.Song
		if err := rows.Scan(&song.ID, &song.Group, &song.Song, &song.ReleaseDate, &song.Link); err != nil {
			return nil, fmt.Errorf("%s: %v", op, err)
		}
		songs = append(songs, song)
	}

	return songs, nil
}

func (r *Repository) GetSongTextByGroup(groupName, songName string, limit, offset int) ([]string, error) {
	op := "repository.GetSongTextByGroup"

	query := `
	SELECT sl.verse 
	FROM song_text sl
	JOIN song_info si ON si.id = sl.song_id
	WHERE si.group_name = $1 AND si.song = $2
	ORDER BY sl.id
	LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(query, groupName, songName, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}
	defer rows.Close()

	var verses []string
	for rows.Next() {
		var verse string
		if err := rows.Scan(&verse); err != nil {
			return nil, fmt.Errorf("%s: %v", op, err)
		}
		verses = append(verses, verse)
	}

	return verses, nil
}

func (r *Repository) DeleteSong(groupName, songName string) error {
	op := "repository.DeleteSong"

	query := `DELETE FROM song_info WHERE group_name = $1 AND song = $2`

	result, err := r.db.Exec(query, groupName, songName)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %v", op, sql.ErrNoRows)
	}

	return nil
}

func (r *Repository) UpdateSongInfo(groupName, songName, newReleaseDate, newLink string) error {
	op := "repository.UpdateSongInfo"

	query := `UPDATE song_info SET updated_at = CURRENT_TIMESTAMP`
	var args []interface{}
	argCount := 1

	if newReleaseDate != "" {
		query += `, release_date = $` + fmt.Sprintf("%d", argCount)
		args = append(args, newReleaseDate)
		argCount++
	}

	if newLink != "" {
		query += `, link = $` + fmt.Sprintf("%d", argCount)
		args = append(args, newLink)
		argCount++
	}

	query += ` WHERE group_name = $` + fmt.Sprintf("%d", argCount)
	args = append(args, groupName)
	argCount++

	query += ` AND song = $` + fmt.Sprintf("%d", argCount)
	args = append(args, songName)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %v", op, sql.ErrNoRows)
	}

	return nil
}

func (r *Repository) UpdateSongText(groupName, songName string, newVerses []string) error {
	op := "repository.UpdateSongText"

	var songID int
	query := `SELECT id FROM song_info WHERE group_name = $1 AND song = $2`
	err := r.db.QueryRow(query, groupName, songName).Scan(&songID)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	_, err = r.db.Exec(`DELETE FROM song_text WHERE song_id = $1`, songID)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	for _, verse := range newVerses {
		_, err := r.db.Exec(`INSERT INTO song_text (song_id, verse) VALUES ($1, $2)`, songID, verse)
		if err != nil {
			return fmt.Errorf("%s: %v", op, err)
		}
	}

	return nil
}

func (r *Repository) UpdateSong(groupName, songName, newReleaseDate, newLink string, newVerses []string) error {
	op := "repository.UpdateSong"

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	var songID int
	query := `SELECT id FROM song_info WHERE group_name = $1 AND song = $2`
	err = tx.QueryRow(query, groupName, songName).Scan(&songID)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows
		}
		return fmt.Errorf("%s: %v", op, err)
	}

	err = r.UpdateSongInfo(groupName, songName, newReleaseDate, newLink)
	if err != nil {
		tx.Rollback()
		return err
	}

	if len(newVerses) > 0 {
		err = r.UpdateSongText(groupName, songName, newVerses)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	return nil
}
