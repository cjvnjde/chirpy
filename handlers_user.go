package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/cjvnjde/chirpy/internal/auth"
	"github.com/cjvnjde/chirpy/internal/database"
	"github.com/google/uuid"
)

func (a *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	type userBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	params := userBody{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("decode response: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	pswHash, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("hash password response: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	user, err := a.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Email:     params.Email,
		HashedPassword: sql.NullString{
			String: pswHash,
			Valid:  pswHash != "",
		},
	})
	if err != nil {
		log.Printf("create user response: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, newUserResponse(user))
}

func (a *application) updateUserHandler(w http.ResponseWriter, r *http.Request, userUUID uuid.UUID) {
	type userBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	params := userBody{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("decode response: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	pswHash, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("hash password response: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	user, err := a.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:        userUUID,
		UpdatedAt: time.Now().UTC(),
		Email:     params.Email,
		HashedPassword: sql.NullString{
			String: pswHash,
			Valid:  pswHash != "",
		},
	})
	if err != nil {
		log.Printf("update user response: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, newUserResponse(user))
}
