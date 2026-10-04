package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

func (a *application) webhooksHandler(w http.ResponseWriter, r *http.Request) {
	type body struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}
	decoder := json.NewDecoder(r.Body)
	params := body{}
	defer r.Body.Close()
	err := decoder.Decode(&params)
	if err != nil {
		somethingWentWrong(w, err)
		return
	}

	switch params.Event {
	case "user.upgraded":
		{
			_, err := a.db.UpgradeUserChirpToRed(r.Context(), params.Data.UserID)

			switch {
			case errors.Is(err, sql.ErrNoRows):
				{
					writeError(w, http.StatusNotFound, "Chirp not found")
					return
				}
			case err != nil:
				{
					writeError(w, http.StatusInternalServerError, "Internal server error")
				}
			}

			w.WriteHeader(http.StatusNoContent)
			return
		}
	default:
		{
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
}
