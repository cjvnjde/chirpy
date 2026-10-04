package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/cjvnjde/chirpy/internal/database"
	"github.com/google/uuid"
)

func (a *application) chirpHandler(w http.ResponseWriter, r *http.Request, userUUID uuid.UUID) {
	type body struct {
		Body string `json:"body"`
	}
	params := body{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("decode response: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if len(params.Body) > 140 {
		writeError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	chirp, err := a.db.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Body:      censorWords(params.Body),
		UserID:    userUUID,
	})
	if err != nil {
		log.Printf("create chirp response: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, ChirpItemResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

type ByUpdatedAt []database.Chirp

func (a ByUpdatedAt) Len() int {
	return len(a)
}

func (a ByUpdatedAt) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByUpdatedAt) Less(i, j int) bool {
	return a[i].UpdatedAt.After(a[j].UpdatedAt)
}

func (a *application) allChirpsHandler(w http.ResponseWriter, r *http.Request) {
	authorID := r.URL.Query().Get("author_id")
	sortOrder := r.URL.Query().Get("sort")
	var data []database.Chirp
	var err error
	var authorUUID uuid.UUID

	if authorID != "" {
		authorUUID, err = uuid.Parse(authorID)
		if err == nil {
			data, err = a.db.GetUserChirps(r.Context(), authorUUID)
		}
	} else {
		data, err = a.db.GetAllChirps(r.Context())
	}

	if sortOrder == "desc" {
		fmt.Println(sortOrder)
		sort.Sort(ByUpdatedAt(data))
	}

	if err != nil {
		log.Printf("db response: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	allChirps := make([]ChirpItemResponse, len(data))

	for i, chirp := range data {
		allChirps[i] = ChirpItemResponse(chirp)
	}

	writeJSON(w, http.StatusOK, allChirps)
}

func (a *application) deleteChirpHandler(w http.ResponseWriter, r *http.Request, userUUID uuid.UUID) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("parse response: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	chirp, err := a.db.GetChirpByID(r.Context(), chirpID)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		{
			writeError(w, http.StatusNotFound, "Chirp not found")
			return
		}
	case err != nil:
		{
			log.Printf("get chirp by id response: %v", err)
			writeError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
	case chirp.UserID != userUUID:
		{
			writeError(w, http.StatusForbidden, "Used does not have permissions to delete this chirp")
			return
		}
	}

	_, err = a.db.DeleteChirpByID(r.Context(), chirpID)
	if err != nil {
		log.Printf("delete chirp response: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *application) getChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("parse response: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	chirp, err := a.db.GetChirpByID(r.Context(), chirpID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		{
			writeError(w, http.StatusNotFound, "Chirp not found")
			return
		}
	case err != nil:
		{
			log.Printf("get chirp by id response: %v", err)
			writeError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
	}
	writeJSON(w, http.StatusOK, ChirpItemResponse(chirp))
}
