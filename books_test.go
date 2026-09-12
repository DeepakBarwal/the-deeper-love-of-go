package books_test

import (
	"books/books"
	"cmp"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"slices"
	"testing"
)

func TestBookToString_FormatsBookInfoAsString(t *testing.T) {
	t.Parallel()
	input := books.Book{
		Title:  "Sea Room",
		Author: "Adam Nicolson",
		Copies: 2,
	}
	want := "Sea Room by Adam Nicolson - 2 copies"
	got := input.String()
	if want != got {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func getTestCatalog() *books.Catalog {
	catalog := books.NewCatalog()
	err := catalog.AddBook(books.Book{
		ID:     "abc",
		Title:  "In the Company of Cheerful Ladies",
		Author: "Alexander McCall Smith",
		Copies: 1,
	})
	if err != nil {
		panic(err)
	}
	err = catalog.AddBook(books.Book{
		Title:  "White Heat",
		Author: "Dominic Sandbrook",
		Copies: 2,
		ID:     "xyz",
	})
	if err != nil {
		panic(err)
	}
	return catalog
}

func TestGetAllBooks_ReturnsAllBooks(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalog()
	bookList := catalog.GetAllBooks()
	assertTestBooks(t, bookList)
}

func TestGetBook_FindsBookInCatalogByID(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalog()
	want := books.Book{
		ID:     "abc",
		Title:  "In the Company of Cheerful Ladies",
		Author: "Alexander McCall Smith",
		Copies: 1,
	}
	got, ok := catalog.GetBook("abc")
	if !ok {
		t.Fatal("book not found")
	}
	if want != got {
		t.Fatalf("want %#v, got %#v", want, got)
	}
}

func TestGetBook_ReturnsFalseWhenBookNotFound(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalog()
	_, ok := catalog.GetBook("nonexistent ID")
	if ok {
		t.Fatal("want false for nonexistent ID, got true")
	}
}

func TestAddBook_AddsGivenBookToCatalog(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalog()
	_, ok := catalog.GetBook("123")
	if ok {
		t.Fatal("book already present")
	}
	err := catalog.AddBook(books.Book{
		ID:     "123",
		Title:  "The Prize of all the Oceans",
		Author: "Glyn Williams",
		Copies: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, ok = catalog.GetBook("123")
	if !ok {
		t.Fatal("added book not found")
	}
}

func TestSetCopies_SetsNumberOfCopiesToGivenValue(t *testing.T) {
	t.Parallel()
	book := books.Book{
		Copies: 5,
	}
	err := book.SetCopies(12)
	if err != nil {
		t.Fatal(err)
	}
	if book.Copies != 12 {
		t.Errorf("want 12 copies, got %d", book.Copies)
	}
}

func TestSetCopies_ReturnsErrorIfCopiesNegative(t *testing.T) {
	t.Parallel()
	book := books.Book{}
	err := book.SetCopies(-1)
	if err == nil {
		t.Error("want error for negative copies, got nil")
	}
}

func TestOpenCatalog_ReadsSameDataWrittenBySync(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalog()
	path := t.TempDir() + "/catalog"
	catalog.Path = path
	err := catalog.Sync()
	if err != nil {
		t.Fatal(err)
	}
	newCatalog, err := books.OpenCatalog(path)
	if err != nil {
		t.Fatal(err)
	}
	bookList := newCatalog.GetAllBooks()
	assertTestBooks(t, bookList)
}

func assertTestBooks(t *testing.T, got []books.Book) {
	t.Helper()
	want := []books.Book{
		{
			Title:  "In the Company of Cheerful Ladies",
			Author: "Alexander McCall Smith",
			Copies: 1,
			ID:     "abc",
		},
		{
			Title:  "White Heat",
			Author: "Dominic Sandbrook",
			Copies: 2,
			ID:     "xyz",
		},
	}
	slices.SortFunc(got, func(a, b books.Book) int {
		return cmp.Compare(a.Author, b.Author)
	})
	if !slices.Equal(want, got) {
		t.Fatalf("want %#v, got %#v", want, got)
	}
}

func TestSetCopies_OnCatalogModifiesSpecifiedBook(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalog()
	book, ok := catalog.GetBook("abc")
	if !ok {
		t.Fatal("book not found")
	}
	if book.Copies != 1 {
		t.Fatalf("want 1 copy before change, got %d", book.Copies)
	}
	err := catalog.SetCopies("abc", 2)
	if err != nil {
		t.Fatal(err)
	}
	book, ok = catalog.GetBook("abc")
	if !ok {
		t.Fatal("book not found")
	}
	if book.Copies != 2 {
		t.Fatalf("want 2 copies after change, got %d", book.Copies)
	}
}

func TestAddBook_ReturnsErrorIfIDExists(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalog()
	_, ok := catalog.GetBook("abc")
	if !ok {
		t.Fatal("book not present")
	}
	err := catalog.AddBook(books.Book{
		ID:     "abc",
		Title:  "In the Company of Cheerful Ladies",
		Author: "Alexander McCall Smith",
		Copies: 1,
	})
	if err == nil {
		t.Fatal("want error for duplicate ID, got nil")
	}
}

func TestSetCopies_IsRaceFree(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalog()
	go func() {
		for range 100 {
			err := catalog.SetCopies("abc", 0)
			if err != nil {
				panic(err)
			}
		}
	}()
	for range 100 {
		_, err := catalog.GetCopies("abc")
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestNewCatalog_CreatesEmptyCatalog(t *testing.T) {
	t.Parallel()
	catalog := books.NewCatalog()
	books := catalog.GetAllBooks()
	if len(books) > 0 {
		t.Errorf("want empty catalog, got %#v", books)
	}
}

func TestServer_ListsAllBooks(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalog()
	catalog.Path = t.TempDir() + "/catalog"
	addr := randomLocalAddr(t)
	go func() {
		err := books.ListenAndServe(addr, catalog)
		if err != nil {
			panic(err)
		}
	}()
	resp, err := http.Get("http://" + addr + "/v1/list")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d", resp.StatusCode)
	}
	bookList := []books.Book{}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(data, &bookList)
	if err != nil {
		t.Fatalf("%v in %q", err, data)
	}
	assertTestBooks(t, bookList)
}

func randomLocalAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().String()
}

func TestServer_FindsBookByID(t *testing.T) {
	t.Parallel()
	addr := randomLocalAddr(t)
	catalog := getTestCatalog()
	catalog.Path = t.TempDir() + "/catalog"
	go func() {
		err := books.ListenAndServe(addr, catalog)
		if err != nil {
			panic(err)
		}
	}()
	resp, err := http.Get("http://" + addr + "/v1/find/abc")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d", resp.StatusCode)
	}
	got := books.Book{}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(data, &got)
	if err != nil {
		t.Fatalf("%v in %q", err, data)
	}
	want := books.Book{
		Title:  "In the Company of Cheerful Ladies",
		Author: "Alexander McCall Smith",
		Copies: 1,
		ID:     "abc",
	}
	if want != got {
		t.Fatalf("want %#v, got %#v", want, got)
	}
}

func TestServer_FindReturnsNotFoundWhenBookNotFound(t *testing.T) {
	t.Parallel()
	addr := randomLocalAddr(t)
	catalog := getTestCatalog()
	catalog.Path = t.TempDir() + "/catalog"
	go func() {
		err := books.ListenAndServe(addr, catalog)
		if err != nil {
			panic(err)
		}
	}()
	resp, err := http.Get("http://" + addr + "/v1/find/bogus")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected status %d", resp.StatusCode)
	}
}

func TestGetBook_OnClientFindsBookByID(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	got, err := client.GetBook("abc")
	if err != nil {
		t.Fatal(err)
	}
	if got != ABC {
		t.Fatalf("want %#v, got %#v", ABC, got)
	}
}

func getTestClient(t *testing.T) *books.Client {
	t.Helper()
	addr := randomLocalAddr(t)
	catalog := getTestCatalog()
	catalog.Path = t.TempDir() + "/catalog"
	go func() {
		err := books.ListenAndServe(addr, catalog)
		if err != nil {
			panic(err)
		}
	}()
	return books.NewClient(addr)
}

var (
	ABC = books.Book{
		Title:  "In the Company of Cheerful Ladies",
		Author: "Alexander McCall Smith",
		Copies: 1,
		ID:     "abc",
	}

	XYZ = books.Book{
		Title:  "White Heat",
		Author: "Dominic Sandbrook",
		Copies: 2,
		ID:     "xyz",
	}
)

func TestGetBook_FindReturnsErrorWhenBookNotFound(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	_, err := client.GetBook("bogus")
	if err == nil {
		t.Error("want error when book not found, got nil")
	}
}

func TestGetAllBooks_OnClientListsAllBooks(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	bookList, err := client.GetAllBooks()
	if err != nil {
		t.Fatal(err)
	}
	assertTestBooks(t, bookList)
}

func TestGetCopies_OnClientReturnsCopiesForBook(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	copies, err := client.GetCopies("abc")
	if err != nil {
		t.Fatal(err)
	}
	if copies != 1 {
		t.Fatalf("want 1 copy, got %d", copies)
	}
}

func TestGetCopies_OnClientErrorsWhenBookNotFound(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	_, err := client.GetCopies("bogus")
	if err == nil {
		t.Error("want error when book not found, got nil")
	}
}

func TestAddCopies_CorrectlyUpdatesStockLevel(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	copies, err := client.GetCopies("abc")
	if err != nil {
		t.Fatal(err)
	}
	if copies != 1 {
		t.Fatalf("want 1 copy before change, got %d", copies)
	}
	stock, err := client.AddCopies("abc", 2)
	if err != nil {
		t.Fatal(err)
	}
	if stock != 3 {
		t.Fatalf("want 3 copies after change, got %d", copies)
	}
}

func TestAddCopies_OnClientErrorsWhenBookNotFound(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	_, err := client.AddCopies("bogus", 1)
	if err == nil {
		t.Error("want error when book not found, got nil")
	}
}

func TestSubCopies_CorrectlyUpdatesStockLevel(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	copies, err := client.GetCopies("abc")
	if err != nil {
		t.Fatal(err)
	}
	if copies != 1 {
		t.Fatalf("want 1 copy before change, got %d", copies)
	}
	stock, err := client.SubCopies("abc", 1)
	if err != nil {
		t.Fatal(err)
	}
	if stock != 0 {
		t.Fatalf("want 0 copies after change, got %d", copies)
	}
}

func TestSubCopies_FailsIfStockLevelTooLow(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	copies, err := client.GetCopies("abc")
	if err != nil {
		t.Fatal(err)
	}
	if copies != 1 {
		t.Fatalf("want 1 copy before change, got %d", copies)
	}
	_, err = client.SubCopies("abc", 2)
	if err == nil {
		t.Error("want error when stock level too low, got nil")
	}
}
