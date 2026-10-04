package main

import (
	"net/http"
)

func initRouts(app *application, serveMux *http.ServeMux) {
	serveMux.Handle("/app/", app.middlewareMetricInc(http.StripPrefix("/app", http.FileServer(http.Dir("./public")))))

	serveMux.HandleFunc("GET /api/healthz", app.healthzHandler)
	// admin
	serveMux.HandleFunc("GET /admin/metrics", app.metricsHandler)
	serveMux.HandleFunc("POST /admin/reset", app.resetHandler)
	// chirps
	serveMux.HandleFunc("POST /api/chirps", app.requireAuth(app.chirpHandler))
	serveMux.HandleFunc("GET /api/chirps", app.allChirpsHandler)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", app.getChirpHandler)
	serveMux.HandleFunc("DELETE /api/chirps/{chirpID}", app.requireAuth(app.deleteChirpHandler))
	// webhooks
	serveMux.HandleFunc("POST /api/polka/webhooks", app.requirePolkaKey(app.webhooksHandler))
	// users
	serveMux.HandleFunc("POST /api/users", app.createUserHandler)
	serveMux.HandleFunc("PUT /api/users", app.requireAuth(app.updateUserHandler))
	// auth
	serveMux.HandleFunc("POST /api/login", app.loginHandler)
	serveMux.HandleFunc("POST /api/refresh", app.refreshHandler)
	serveMux.HandleFunc("POST /api/revoke", app.revokeHandler)
}
