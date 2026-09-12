package books

import (
	"encoding/json"
	"net/http"
)

func ListenAndServe(addr string, catalog *Catalog) error {
	return http.ListenAndServe(addr, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			books := catalog.GetAllBooks()
			err := json.NewEncoder(w).Encode(books)
			if err != nil {
				panic(err)
			}
		}))
}
