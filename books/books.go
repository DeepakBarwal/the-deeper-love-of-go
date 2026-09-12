package books

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"sync"
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

type Catalog struct {
	mu   *sync.RWMutex
	data map[string]Book
	Path string
}

func (catalog *Catalog) GetAllBooks() []Book {
	books := maps.Values(catalog.data)
	return slices.Collect(books)
}

func (catalog *Catalog) GetBook(ID string) (Book, bool) {
	book, ok := catalog.data[ID]
	return book, ok
}

func (catalog *Catalog) AddBook(book Book) error {
	_, ok := catalog.data[book.ID]
	if ok {
		return fmt.Errorf("ID %q already exists", book.ID)
	}
	catalog.data[book.ID] = book
	return nil
}

func (book *Book) SetCopies(copies int) error {
	if copies < 0 {
		return fmt.Errorf("negative number of copies: %d", copies)
	}
	book.Copies = copies // actually does (*book).Copies = copies
	return nil
}

func OpenCatalog(path string) (*Catalog, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	catalog := NewCatalog()
	err = json.NewDecoder(file).Decode(&catalog.data)
	if err != nil {
		return nil, err
	}
	catalog.Path = path
	return catalog, nil
}

func (catalog *Catalog) Sync() error {
	file, err := os.Create(catalog.Path)
	if err != nil {
		return err
	}
	defer file.Close()
	err = json.NewEncoder(file).Encode(catalog.data)
	if err != nil {
		return err
	}
	return nil
}

func (catalog *Catalog) SetCopies(ID string, copies int) error {
	book, ok := catalog.data[ID]
	if !ok {
		return fmt.Errorf("ID %q not found", ID)
	}
	err := book.SetCopies(copies)
	if err != nil {
		return err
	}
	catalog.data[ID] = book
	return nil
}

func (catalog *Catalog) GetCopies(ID string) (int, error) {
	book, ok := catalog.data[ID]
	if !ok {
		return 0, fmt.Errorf("ID %q not found", ID)
	}
	return book.Copies, nil
}

func NewCatalog() *Catalog {
	return &Catalog{
		mu:   &sync.RWMutex{},
		data: map[string]Book{},
	}
}
