package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
)

type httpError struct {
	Error string `json:"error"`
}

func NewJSONError(s string) ([]byte, error) {
	dat, err := json.Marshal(httpError{
		Error: "Error decoding parameters",
	})
	if err != nil {
		return []byte{}, err
	}

	return dat, nil
}

func somethingWentWrong(w http.ResponseWriter, err error) {
	log.Printf("Error decoding parameters: %s", err)
	w.WriteHeader(500)
	dat, err := NewJSONError("Something went wrong")
	if err != nil {
		w.Write([]byte("Erorr"))
		return
	}
	w.Write(dat)
}

func replaceWords(body string) string {
	bannedWords := []string{"kerfuffle", "sharbert", "fornax"}

	words := strings.Split(body, " ")
	newStr := make([]string, 0, len(words))

	for _, word := range words {
		loverWord := strings.ToLower(word)
		shouldSkip := false
		for _, banned := range bannedWords {
			if loverWord == banned {
				shouldSkip = true
			}
		}
		if !shouldSkip {
			newStr = append(newStr, word)
		} else {
			newStr = append(newStr, "****")
		}
	}

	return strings.Join(newStr, " ")
}

func main() {
	serveMux := http.NewServeMux()
	server := http.Server{
		Handler: serveMux,
		Addr:    ":8080",
	}

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}
	apiCfg.fileserverHits.Store(0)

	serveMux.Handle("/app/", apiCfg.middlewareMetricInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	serveMux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	serveMux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)

	serveMux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
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
			CleanedBody: replaceWords(params.Body),
		}
		d, err := json.Marshal(v)
		if err != nil {
			somethingWentWrong(w, err)
			return
		}

		w.Write(d)
	})

	server.ListenAndServe()
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(200)
	hits := cfg.fileserverHits.Load()
	content := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, hits)
	rw.Write([]byte(content))
}

func (cfg *apiConfig) resetHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(200)
	cfg.fileserverHits.Store(0)
	hits := strconv.Itoa(int(cfg.fileserverHits.Load()))
	rw.Write([]byte(string(hits)))
}
