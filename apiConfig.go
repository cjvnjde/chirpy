package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/cjvnjde/chirpy/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
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

func (cfg *apiConfig) validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		somethingWentWrong(w, err)
		return
	}

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

	type valid struct {
		CleanedBody string `json:"cleaned_body"`
	}

	v := valid{
		CleanedBody: cencorWords(params.Body),
	}
	d, err := json.Marshal(v)
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	w.Write(d)
}

func (cfg *apiConfig) healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (c *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	type userBody struct {
		Email string `json:"email"`
	}
	body := userBody{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&body)
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	user, err := c.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     body.Email,
	})
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	type userCreated struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json")

	d, err := json.Marshal(userCreated{
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
