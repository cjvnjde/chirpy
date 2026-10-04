package main

import (
	"sync/atomic"

	"github.com/cjvnjde/chirpy/internal/database"
)

type application struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	jwtSecret      string
	polkaKey       string
}
