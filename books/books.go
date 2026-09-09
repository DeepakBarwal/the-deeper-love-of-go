package books

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
)

type Book struct {
	Title  string
	Author string
	Copies int
	ID     string
}

func (book Book) String() string {
	return fmt.Sprintf("%v by %v - %v copies",
		book.Title, book.Author, book.Copies)
}

type Catalog map[string]Book

func (catalog Catalog) GetAllBooks() []Book {
	books := maps.Values(catalog)
	return slices.Collect(books)
}

func (catalog Catalog) GetBook(ID string) (Book, bool) {
	book, ok := catalog[ID]
	return book, ok
}

func (catalog Catalog) AddBook(book Book) {
	catalog[book.ID] = book
}

func (book *Book) SetCopies(copies int) error {
	if copies < 0 {
		return fmt.Errorf("begative number of copies: %d", copies)
	}
	book.Copies = copies // actually does (*book).Copies = copies
	return nil
}

func OpenCatalog(path string) (Catalog, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	catalog := Catalog{}
	err = json.NewDecoder(file).Decode(&catalog)
	if err != nil {
		return nil, err
	}
	return catalog, nil
}
