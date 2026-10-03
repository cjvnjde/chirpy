package main

import "strings"

func censorWords(body string) string {
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
