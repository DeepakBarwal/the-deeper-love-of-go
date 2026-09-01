package main

import "fmt"

type Book struct {
	Title  string
	Author string
	Copies int
}

func main() {
	book := Book{
		Title:  "Engineering in Plain Sight",
		Author: "Grady Hillhouse",
		Copies: 2,
	}
	fmt.Println(BookToString(book))
}

func BookToString(book Book) string {
	return fmt.Sprintf("%v by %v - %v copies",
		book.Title, book.Author, book.Copies)
}
