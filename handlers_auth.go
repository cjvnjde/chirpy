package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/cjvnjde/chirpy/internal/auth"
	"github.com/cjvnjde/chirpy/internal/database"
)

func (a *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	type body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	params := body{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("decode response: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	user, err := a.db.GetUserByEmail(r.Context(), params.Email)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		{
			writeError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
	case err != nil:
		{
			log.Printf("get user by email response: %v", err)
			writeError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
	}
	isOk, err := auth.CheckPasswordHash(params.Password, user.HashedPassword.String)
	if err != nil {
		log.Printf("check password hash response: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if !isOk {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	expiresIn := time.Duration(1 * time.Hour)

	refreshToken := auth.MakeRefreshToken()

	rt, err := a.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(60 * 24 * time.Hour),
	})
	if err != nil {
		log.Printf("create refresh token response: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	jwt, err := auth.MakeJWT(user.ID, a.jwtSecret, expiresIn)
	if err != nil {
		log.Printf("make jwt response: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, loginResponse{
		userResponse: newUserResponse(user),
		Token:        jwt,
		RefreshToken: rt.Token,
	})
}

func (a *application) refreshHandler(w http.ResponseWriter, r *http.Request) {
	rt, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	refreshToken, err := a.db.GetRefreshToken(r.Context(), rt)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if refreshToken.RevokedAt.Valid {
		if !refreshToken.RevokedAt.Time.After(time.Now().UTC()) {
			writeError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
	}

	if !refreshToken.ExpiresAt.After(time.Now().UTC()) {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := a.db.GetUserFromRefreshToken(r.Context(), refreshToken.Token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	newToken, err := auth.MakeJWT(user.ID, a.jwtSecret, time.Duration(1*time.Hour))
	if err != nil {
		log.Printf("make jwt response: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	type token struct {
		Token string `json:"token"`
	}

	writeJSON(w, http.StatusOK, token{
		Token: newToken,
	})
}

func (a *application) revokeHandler(w http.ResponseWriter, r *http.Request) {
	rt, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	refreshToken, err := a.db.GetRefreshToken(r.Context(), rt)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	_, err = a.db.RevokeRefreshToken(r.Context(), database.RevokeRefreshTokenParams{
		Token: refreshToken.Token,
		RevokedAt: sql.NullTime{
			Time:  time.Now().UTC(),
			Valid: true,
		},
	})
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	w.WriteHeader(204)
}
