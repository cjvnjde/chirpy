package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type httpError struct {
	Error string `json:"error"`
}

func NewJSONError(s string) ([]byte, error) {
	dat, err := json.Marshal(httpError{
		Error: s,
	})
	if err != nil {
		return []byte{}, err
	}

	return dat, nil
}

func somethingWentWrong(w http.ResponseWriter, err error) {
	fmt.Println(err)
	w.WriteHeader(500)
	dat, err := NewJSONError("Something went wrong")
	if err != nil {
		w.Write([]byte("Erorr"))
		return
	}
	w.Write(dat)
}
