package main

import (
	"net/http"
	"strconv"
	"sync/atomic"
)

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

	serveMux.HandleFunc("GET /healthz", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.WriteHeader(200)
		rw.Write([]byte("OK"))
	})

	serveMux.HandleFunc("GET /metrics", apiCfg.metricsHandler)
	serveMux.HandleFunc("POST /reset", apiCfg.resetHandler)

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
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(200)
	hits := cfg.fileserverHits.Load()
	str := "Hits: " + strconv.Itoa(int(hits))
	rw.Write([]byte(str))
}

func (cfg *apiConfig) resetHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(200)
	cfg.fileserverHits.Store(0)
	hits := strconv.Itoa(int(cfg.fileserverHits.Load()))
	rw.Write([]byte(string(hits)))
}
