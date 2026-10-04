package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/cjvnjde/chirpy/internal/auth"
	"github.com/cjvnjde/chirpy/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	jwtSecret      string
}

func (cfg *apiConfig) middlewareMetricInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	hits := cfg.fileserverHits.Load()
	content := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, hits)
	w.Write([]byte(content))
}

func (c *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("PLATFORM") != "dev" {
		w.WriteHeader(403)
		return
	}
	err := c.db.DeleteUsers(r.Context())
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	c.fileserverHits.Store(0)
	hits := strconv.Itoa(int(c.fileserverHits.Load()))
	w.Write([]byte(string(hits)))
}

func (cfg *apiConfig) healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (c *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	type userBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	body := userBody{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&body)
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	pswHash, err := auth.HashPassword(body.Password)
	if err != nil {
		somethingWentWrong(w, err)
		return
	}
	user, err := c.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Email:     body.Email,
		HashedPassword: sql.NullString{
			String: pswHash,
			Valid:  pswHash != "",
		},
	})
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json")

	d, err := json.Marshal(UserItemResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	w.Write(d)
}

func (c *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	userUUID, err := c.checkAuth(w, r)
	if err != nil {
		unauthorized(w, err)
		return
	}

	type userBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	body := userBody{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err = decoder.Decode(&body)
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	pswHash, err := auth.HashPassword(body.Password)
	if err != nil {
		somethingWentWrong(w, err)
		return
	}
	user, err := c.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:        userUUID,
		UpdatedAt: time.Now().UTC(),
		Email:     body.Email,
		HashedPassword: sql.NullString{
			String: pswHash,
			Valid:  pswHash != "",
		},
	})
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	d, err := json.Marshal(UserItemResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(d)
}

func (c *apiConfig) chirpHandler(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		unauthorized(w, err)
		return
	}
	userUuid, err := auth.ValidateJWT(bearerToken, c.jwtSecret)
	if err != nil {
		unauthorized(w, err)
		return
	}

	type chirpBody struct {
		Body string `json:"body"`
	}
	params := chirpBody{}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if len(params.Body) > 140 {
		w.WriteHeader(400)
		dat, err := NewJSONError("Chirp is too long")
		if err != nil {
			w.Write([]byte("Erorr"))
			return
		}
		w.Write(dat)
		return
	}

	chirp, err := c.db.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Body:      censorWords(params.Body),
		UserID:    userUuid,
	})
	if err != nil {
		somethingWentWrong(w, err)
		return
	}
	d, err := json.Marshal(ChirpItemResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
	if err != nil {
		somethingWentWrong(w, err)
		return
	}
	w.WriteHeader(201)
	w.Write(d)
}

func (c *apiConfig) allChirpsHandler(w http.ResponseWriter, r *http.Request) {
	data, err := c.db.GetAllChirps(r.Context())
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	allChirps := make([]ChirpItemResponse, len(data))

	for i, chirp := range data {
		allChirps[i] = ChirpItemResponse(chirp)
	}

	d, err := json.Marshal(allChirps)
	if err != nil {
		somethingWentWrong(w, err)
		return
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(d)
}

func (c *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		somethingWentWrong(w, err)
		return
	}
	userUUID, err := c.checkAuth(w, r)
	if err != nil {
		unauthorized(w, err)
		return
	}
	chirp, err := c.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte{})
		return
	}
	if chirp.UserID != userUUID {
		w.WriteHeader(403)
		w.Write([]byte{})
		return
	}

	_, err = c.db.DeleteChirpByID(r.Context(), chirpID)
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	w.WriteHeader(204)
	w.Write([]byte{})
}

func (c *apiConfig) getChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		somethingWentWrong(w, err)
		return
	}
	chirp, err := c.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte{})
		return
	}
	d, err := json.Marshal(ChirpItemResponse(chirp))
	if err != nil {
		somethingWentWrong(w, err)
		return
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(d)
}

func (c *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	type body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	params := body{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	user, err := c.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte{})
		return
	}
	isOk, err := auth.CheckPasswordHash(params.Password, user.HashedPassword.String)

	if !isOk {
		w.WriteHeader(401)
		w.Write([]byte{})
		return
	}

	expiresIn := time.Duration(1 * time.Hour)

	refreshToken := auth.MakeRefreshToken()

	rt, err := c.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(60 * 24 * time.Hour),
	})
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	jwt, err := auth.MakeJWT(user.ID, c.jwtSecret, expiresIn)
	if err != nil {
		somethingWentWrong(w, err)
		return
	}
	d, err := json.Marshal(UserItemResponse{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        jwt,
		RefreshToken: rt.Token,
	})
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(d)
}

func (c *apiConfig) checkAuth(w http.ResponseWriter, r *http.Request) (userUUID uuid.UUID, err error) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		return uuid.UUID{}, err
	}
	userUUID, err = auth.ValidateJWT(bearerToken, c.jwtSecret)
	if err != nil {
		return uuid.UUID{}, err
	}

	return userUUID, nil
}

func (c *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	rt, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte{})
		return
	}

	refreshToken, err := c.db.GetRefreshToken(r.Context(), rt)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte{})
		return
	}
	if refreshToken.RevokedAt.Valid {
		if !refreshToken.RevokedAt.Time.After(time.Now().UTC()) {
			w.WriteHeader(401)
			w.Write([]byte{})
			return
		}
	}

	if !refreshToken.ExpiresAt.After(time.Now().UTC()) {
		w.WriteHeader(401)
		w.Write([]byte{})
		return
	}

	user, err := c.db.GetUserFromRefreshToken(r.Context(), refreshToken.Token)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte{})
		return
	}

	newToken, err := auth.MakeJWT(user.ID, c.jwtSecret, time.Duration(1*time.Hour))
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	type token struct {
		Token string `json:"token"`
	}

	d, err := json.Marshal(token{
		Token: newToken,
	})
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(d)
}

func (c *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	rt, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte{})
		return
	}

	refreshToken, err := c.db.GetRefreshToken(r.Context(), rt)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte{})
		return
	}

	_, err = c.db.RevokeRefreshToken(r.Context(), database.RevokeRefreshTokenParams{
		Token: refreshToken.Token,
		RevokedAt: sql.NullTime{
			Time:  time.Now().UTC(),
			Valid: true,
		},
	})
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte{})
		return
	}

	w.WriteHeader(204)
	w.Write([]byte{})
}
