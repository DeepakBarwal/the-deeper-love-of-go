package books

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	addr string
}

func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
	}
}

func (client *Client) GetBook(ID string) (Book, error) {
	resp, err := http.Get("http://" + client.addr + "/v1/find/" + ID)
	if err != nil {
		return Book{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Book{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	book := Book{}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Book{}, err
	}
	err = json.Unmarshal(data, &book)
	if err != nil {
		return Book{}, fmt.Errorf("%v in %q", err, data)
	}
	return book, nil
}
