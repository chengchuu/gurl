// Command basic demonstrates query inspection with the public gurl package.
package main

import (
	"fmt"
	"log"

	"github.com/chengchuu/gurl"
)

func main() {
	value, err := gurl.GetQueryParam("https://example.com/search?q=go&lang=en", "q")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(value)
	fmt.Println(gurl.CheckValidHTTPURL("https://example.com"))
	fmt.Println(gurl.CheckValidHTTPURL("https:"))
}
