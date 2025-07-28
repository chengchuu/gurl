# GURL: Go URL Manipulation Library

GURL is a Go library for URL manipulation and validation. It provides functions for getting and setting URL parameters, paths, hosts, and protocols, as well as checking URL validity and getting file types from URLs.

## Installation

To install GURL, execute the following command:

```bash
go get github.com/mazeychuu/gurl
```

## Usage

Here's a quick example of how to use GURL:

```go
package main

import (
 "fmt"
 "github.com/mazeychuu/gurl"
)

func main() {
 result, err := gurl.GetQueryParam("http://example.com/?t1=1&t2=2", "t1")
 if err != nil {
  panic(err)
 }
 fmt.Println(result) // Output: "1"
}
```

## Functions

| Function | Parameter | Return Value | Description |
|----------|------------|--------------|-------------|
| `GetQueryParam` | `url, param string` | `string, error` | Get the value of a query parameter from a URL |
| `SetQueryParam` | `url, param, value string` | `string, error` | Set the value of a query parameter in a URL |
| `DelQueryParam` | `url, param string` | `string, error` | Delete a query parameter from a URL |
| `GetHashParam` | `url, param string` | `string, error` | Get the value of a query parameter from the URL fragment |
| `SetHashParam` | `url, param, value string` | `string, error` | Set the value of a query parameter in the URL fragment |
| `DelHashParam` | `url, param string` | `string, error` | Delete a query parameter from the URL fragment |
| `GetPath` | `url string` | `string, error` | Get the path from a URL |
| `SetPath` | `url, newPath string` | `string, error` | Set the path in a URL |
| `GetHost` | `url string` | `string, error` | Get the host from a URL |
| `SetHost` | `url, newHost string` | `string, error` | Set the host in a URL |
| `GetHostname` | `url string` | `string, error` | Get the hostname from a URL |
| `SetHostname` | `url, newHostname string` | `string, error` | Set the hostname in a URL |
| `GetProtocol` | `url string` | `string, error` | Get the protocol from a URL |
| `SetProtocol` | `url, newProtocol string` | `string, error` | Set the protocol in a URL |
| `CheckValid` | `url string` | `bool` | Check if a URL is valid |
| `CheckValidHttpUrl` | `url string` | `bool` | Check if a URL is valid and uses either the HTTP or HTTPS scheme |
| `GetUrlFileType` | `url string` | `string, error` | Get the file type of a URL |
| `GetBaseUrl` | `url string` | `string, error` | Get the base URL without query parameters and fragment |

## Examples

### Basic URL Manipulation

```go
// Get base URL without query parameters and fragment
result, err := gurl.GetBaseUrl("https://example.com/path?param=value#fragment")
if err != nil {
    panic(err)
}
fmt.Println(result) // Output: "https://example.com/path"
```

### Hash Parameters

The typical hash history mode URLs are like this: `https://example.com/path1#path2?p1=1&p2=2`.

```go
link := "https://example.com/path1#path2?p1=1&p2=2"

gurl.GetHashParam(link, "p1") // "1"
gurl.SetHashParam(link, "p1", "3") // "https://example.com/path1#path2?p1=3&p2=2"
gurl.DelHashParam(link, "p1") // "https://example.com/path1#path2?p2=2"
```

## Contributing

Contributions to GURL are welcome! Please submit a pull request or open an issue on [GitHub repository](https://github.com/mazeychuu/gurl).

## License

GURL is licensed under the MIT License. For more details, please refer to the LICENSE file in this repository.
