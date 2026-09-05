package main

import (
	"books/books"
	"fmt"
)

func main() {
	book := books.Book{
		Title:  "Engineering in Plain Sight",
		Author: "Grady Hillhouse",
		Copies: 2,
	}
	fmt.Println(books.BookToString(book))
}
