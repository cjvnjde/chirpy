package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

func (a *application) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	hits := a.fileserverHits.Load()
	content := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, hits)
	w.Write([]byte(content))
}

func (a *application) resetHandler(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("PLATFORM") != "dev" {
		writeError(w, http.StatusForbidden, "No access")
		return
	}
	err := a.db.DeleteUsers(r.Context())
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	a.fileserverHits.Store(0)
	hits := strconv.Itoa(int(a.fileserverHits.Load()))
	w.Write([]byte(string(hits)))
}
