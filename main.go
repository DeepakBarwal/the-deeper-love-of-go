package main

import "fmt"

type Book struct {
	Title  string
	Author string
	Copies int
}

func main() {
	fmt.Println("It's all good!")
}

func BookToString(book Book) string {
	return fmt.Sprintf("%v by %v - %v copies",
		book.Title, book.Author, book.Copies)
}
