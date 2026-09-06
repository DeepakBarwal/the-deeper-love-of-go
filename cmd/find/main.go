package main

import (
	"books/books"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: find <BOOK ID>")
		return
	}
	ID := os.Args[1]
	catalog := books.GetCatalog()
	book, ok := catalog.GetBook(ID)
	if !ok {
		fmt.Println("Sorry, I couldn't find that book in the catalog.")
		return
	}
	fmt.Println(book)
}
