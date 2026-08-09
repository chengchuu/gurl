package gurl_test

import (
	"fmt"

	"github.com/chengchuu/gurl"
)

func ExampleGetQueryParam() {
	value, err := gurl.GetQueryParam("https://example.com/search?q=go&lang=en", "q")
	if err != nil {
		panic(err)
	}
	fmt.Println(value)
	// Output: go
}

func ExampleCheckValidHTTPURL() {
	fmt.Println(gurl.CheckValidHTTPURL("https://example.com"))
	fmt.Println(gurl.CheckValidHTTPURL("https:"))
	// Output:
	// true
	// false
}
