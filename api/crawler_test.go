package api

import (
	"fmt"
	"testing"
)

func TestCrawler(t *testing.T) {
	c, err := NewCrawler(map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	u, err := c.search("9787514847642")
	if err != nil {
		t.Fatal(err)
	}
	bookInfo, err := c.getBookInfo(u)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(bookInfo)
}
