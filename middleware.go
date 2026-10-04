package main

import (
	"net/http"

	"github.com/cjvnjde/chirpy/internal/auth"
	"github.com/google/uuid"
)

type authenticatedHandler func(
	http.ResponseWriter,
	*http.Request,
	uuid.UUID,
)

func (a *application) requireAuth(next authenticatedHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearerToken, err := auth.GetBearerToken(r.Header)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		userUUID, err := auth.ValidateJWT(bearerToken, a.jwtSecret)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		next(w, r, userUUID)
	})
}

func (a *application) middlewareMetricInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (a *application) requirePolkaKey(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		polkaKey, err := auth.GetAPIKey(r.Header)
		if err != nil || polkaKey != a.polkaKey {
			writeError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		next(w, r)
	})
}
