package gurl

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// GetQueryParam retrieves the value of a specified query parameter from a URL.
//
// Parameters:
//
//	url: The URL from which to retrieve the query parameter.
//	param: The name of the query parameter to retrieve.
//
// Returns:
//
//	A string containing the value of the query parameter, and an error if any occurred.
//
// Example:
//
//	result, err := GetQueryParam("http://example.com/?p1=1&p2=2", "p1")
//	if err != nil {
//	  panic(err)
//	}
//	fmt.Println(result) // Output: "1"
func GetQueryParam(u, param string) (string, error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	values := parsedUrl.Query()
	return values.Get(param), nil
}

// SetQueryParam sets the value of a specified query parameter in a URL and returns the new URL.
//
// Parameters:
//
//	url: The URL in which to set the query parameter.
//	param: The name of the query parameter to set.
//	value: The value to set the query parameter to.
//
// Returns:
//
//	A string containing the new URL, and an error if any occurred.
//
// Example:
//
//	result, err := SetQueryParam("http://example.com/?p1=1&p2=2", "p1", "3")
//	if err != nil {
//	  panic(err)
//	}
//	fmt.Println(result) // Output: "http://example.com/?p1=3&p2=2"
func SetQueryParam(u, param, value string) (string, error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	values := parsedUrl.Query()
	values.Set(param, value)
	parsedUrl.RawQuery = values.Encode()
	return parsedUrl.String(), nil
}

// DelQueryParam deletes a specified query parameter from a URL and returns the new URL.
//
// Parameters:
//
//	url: The URL from which to delete the query parameter.
//	param: The name of the query parameter to delete.
//
// Returns:
//
//	A string containing the new URL, and an error if any occurred.
//
// Example:
//
//	result, err := DelQueryParam("http://example.com/?p1=1&p2=2", "p1")
//	if err != nil {
//	  panic(err)
//	}
//	fmt.Println(result) // Output: "http://example.com/?p2=2"
func DelQueryParam(u, param string) (string, error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	values := parsedUrl.Query()
	values.Del(param)
	parsedUrl.RawQuery = values.Encode()
	return parsedUrl.String(), nil
}

func parseFragment(fra string) (path string, query string) {
	if strings.Contains(fra, "?") {
		splitFra := strings.SplitN(fra, "?", 2)
		path = splitFra[0]
		query = splitFra[1]
	} else if strings.HasPrefix(fra, "?") {
		query = fra[1:]
	} else {
		path = fra
	}
	return
}

// GetHashParam retrieves the value of a specified hash parameter from a URL.
//
// Parameters:
//
//	url: The URL from which to retrieve the hash parameter.
//	param: The name of the hash parameter to retrieve.
//
// Returns:
//
//	A string containing the value of the hash parameter, and an error if any occurred.
//
// Example:
//
//	result, err := GetHashParam("http://example.com/#?t1=1&t2=2", "t1")
//	if err != nil {
//	  panic(err)
//	}
//	fmt.Println(result) // Output: "1"
func GetHashParam(u, param string) (string, error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	fra := parsedUrl.Fragment
	_, fraQuery := parseFragment(fra)
	hashParams := []string{}
	if fraQuery != "" {
		hashParams = strings.Split(fraQuery, "&")
	}
	for _, p := range hashParams {
		pair := strings.Split(p, "=")
		if pair[0] == param && len(pair) > 1 {
			return pair[1], nil
		}
	}
	return "", nil // fmt.Errorf("param %s not found in hash", param)
}

// SetHashParam sets the value of a specified hash parameter in a URL and returns the new URL.
//
// Parameters:
//
//	url: The URL in which to set the hash parameter.
//	param: The name of the hash parameter to set.
//	value: The value to set the hash parameter to.
//
// Returns:
//
//	A string containing the new URL, and an error if any occurred.
//
// Example:
//
//	result, err := SetHashParam("http://example.com/#?t1=1&t2=2", "t1", "3")
//	if err != nil {
//	  panic(err)
//	}
//	fmt.Println(result) // Output: "http://example.com/#?t1=3&t2=2"
func SetHashParam(u, param, value string) (string, error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	fra := parsedUrl.Fragment
	fraPath, fraQuery := parseFragment(fra)
	hashParams := []string{}
	if fraQuery != "" {
		hashParams = strings.Split(fraQuery, "&")
	}
	var newHashParams []string
	paramExists := false
	for _, p := range hashParams {
		pair := strings.Split(p, "=")
		if pair[0] == param {
			newHashParams = append(newHashParams, param+"="+value)
			paramExists = true
		} else {
			newHashParams = append(newHashParams, p)
		}
	}
	if !paramExists {
		newHashParams = append(newHashParams, param+"="+value)
	}
	newFraStr := fmt.Sprintf("%s?%s", fraPath, strings.Join(newHashParams, "&"))
	parsedUrl.Fragment = newFraStr
	return parsedUrl.String(), nil
}

// DelHashParam deletes a specified hash parameter from a URL and returns the new URL.
//
// Parameters:
//
//	url: The URL from which to delete the hash parameter.
//	param: The name of the hash parameter to delete.
//
// Returns:
//
//	A string containing the new URL, and an error if any occurred.
//
// Example:
//
//	result, err := DelHashParam("http://example.com/#?t1=1&t2=2", "t1")
//	if err != nil {
//	  panic(err)
//	}
//	fmt.Println(result) // Output: "http://example.com/#?t2=2"
func DelHashParam(u, param string) (string, error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	fra := parsedUrl.Fragment
	fraPath, fraQuery := parseFragment(fra)
	hashParams := []string{}
	if fraQuery != "" {
		hashParams = strings.Split(fraQuery, "&")
	}
	var newHashParams []string
	for _, p := range hashParams {
		pair := strings.Split(p, "=")
		if pair[0] != param {
			newHashParams = append(newHashParams, p)
		}
	}
	newFraStr := ""
	if len(newHashParams) == 0 {
		newFraStr = fraPath
	} else {
		newFraStr = fmt.Sprintf("%s?%s", fraPath, strings.Join(newHashParams, "&"))
	}
	parsedUrl.Fragment = newFraStr
	return parsedUrl.String(), nil
}

// GetPath retrieves the path from a URL.
//
// Parameters:
//
//	url: The URL from which to retrieve the path.
//
// Returns:
//
//	A string containing the path, and an error if any occurred.
//
// Example:
//
//	result, err := GetPath("http://example.com/path/to/resource")
//	if err != nil {
//	  panic(err)
//	}
//	fmt.Println(result) // Output: "/path/to/resource"
func GetPath(u string) (string, error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	return parsedUrl.Path, nil
}

// SetPath sets the path in a URL and returns the new URL.
//
// Parameters:
//
//	url: The URL in which to set the path.
//	newPath: The path to set.
//
// Returns:
//
//	A string containing the new URL, and an error if any occurred.
//
// Example:
//
//	result, err := SetPath("http://example.com/path/to/resource", "/new/path")
//	if err != nil {
//	  panic(err)
//	}
//	fmt.Println(result) // Output: "http://example.com/new/path"
func SetPath(u, newPath string) (string, error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	parsedUrl.Path = newPath
	return parsedUrl.String(), nil
}

// GetHost retrieves the host from a URL.
//
// Parameters:
//
//	url: The URL from which to retrieve the host.
//
// Returns:
//
//	A string containing the host, and an error if any occurred.
//
// Example:
//
//	result, err := GetHost("http://example.com:8080/path/to/resource")
//	if err != nil {
//	  panic(err)
//	}
//	fmt.Println(result) // Output: "example.com:8080"
func GetHost(u string) (string, error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	return parsedUrl.Host, nil
}

// SetHost sets the host in a URL and returns the new URL.
//
// Parameters:
//
//	url: The URL in which to set the host.
//	newHost: The host to set.
//
// Returns:
//
//	A string containing the new URL, and an error if any occurred.
//
// Example:
//
//	result, err := SetHost("http://example.com:8080/path/to/resource", "newhost.com:9090")
//	if err != nil {
//	  panic(err)
//	}
//	fmt.Println(result) // Output: "http://newhost.com:9090/path/to/resource"
func SetHost(u, newHost string) (string, error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	parsedUrl.Host = newHost
	return parsedUrl.String(), nil
}

// GetHostname retrieves the hostname from a URL.
//
// Parameters:
//
//	url: The URL from which to retrieve the hostname.
//
// Returns:
//
//	A string containing the hostname, and an error if any occurred.
//
// Example:
//
//	result, err := GetHostname("http://subdomain.example.com:8080/path/to/resource")
//	if err != nil {
//	  panic(err)
//	}
//	fmt.Println(result) // Output: "subdomain.example.com"
func GetHostname(u string) (string, error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	hostParts := strings.Split(parsedUrl.Host, ":")
	return hostParts[0], nil
}

// SetHostname sets the hostname in a URL and returns the new URL.
//
// Parameters:
//
//	url: The URL in which to set the hostname.
//	newHostname: The hostname to set.
//
// Returns:
//
//	A string containing the new URL, and an error if any occurred.
//
// Example:
//
//	result, err := SetHostname("http://subdomain.example.com:8080/path/to/resource", "newsubdomain.example.com")
//	if err != nil {
//	  panic(err)
//	}
//	fmt.Println(result) // Output: "http://newsubdomain.example.com:8080/path/to/resource"
func SetHostname(u, newHostname string) (string, error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	hostParts := strings.Split(parsedUrl.Host, ":")
	if len(hostParts) > 1 {
		parsedUrl.Host = newHostname + ":" + hostParts[1]
	} else {
		parsedUrl.Host = newHostname
	}
	return parsedUrl.String(), nil
}

// GetProtocol retrieves the protocol from a URL.
//
// Parameters:
//
//	url: The URL from which to retrieve the protocol.
//
// Returns:
//
//	A string containing the protocol, and an error if any occurred.
//
// Example:
//
//	result, err := GetProtocol("http://example.com/path/to/resource")
//	if err != nil {
//	  panic(err)
//	}
//	fmt.Println(result) // Output: "http"
func GetProtocol(u string) (string, error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	return strings.Split(parsedUrl.Scheme, ":")[0], nil
}

// SetProtocol sets the protocol in a URL and returns the new URL.
//
// Parameters:
//
//	url: The URL in which to set the protocol.
//	newProtocol: The protocol to set.
//
// Returns:
//
//	A string containing the new URL, and an error if any occurred.
//
// Example:
//
//	result, err := SetProtocol("http://example.com/path/to/resource", "https")
//	if err != nil {
//	  panic(err)
//	}
//	fmt.Println(result) // Output: "https://example.com/path/to/resource"
func SetProtocol(u, newProtocol string) (string, error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	parsedUrl.Scheme = newProtocol
	return parsedUrl.String(), nil
}

// CheckValid checks if a URL is valid.
//
// Parameters:
//
//	url: The URL to check.
//
// Returns:
//
//	A boolean indicating whether the URL is valid.
//
// Example:
//
//	result := CheckValid("http://example.com/path/to/resource")
//	fmt.Println(result) // Output: true
func CheckValid(u string) bool {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return false
	}
	return parsedUrl.Scheme != "" && parsedUrl.Host != ""
}

// CheckValidHttpUrl checks if a URL is a valid HTTP or HTTPS URL.
//
// Parameters:
//
//	url: The URL to check.
//
// Returns:
//
//	A boolean indicating whether the URL is a valid HTTP or HTTPS URL.
//
// Example:
//
//	result := CheckValidHttpUrl("http://example.com/path/to/resource")
//	fmt.Println(result) // Output: true
func CheckValidHttpUrl(u string) bool {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return false
	}
	return parsedUrl.Scheme == "http" || parsedUrl.Scheme == "https"
}

// GetUrlFileType retrieves the file type from a URL.
//
// Parameters:
//
//	url: The URL from which to retrieve the file type.
//
// Returns:
//
//	A string containing the file type, and an error if any occurred.
//
// Example:
//
//	result, err := GetUrlFileType("https://example.com/a/b/c.png")
//	if err != nil {
//	  panic(err)
//	}
//	fmt.Println(result) // Output: "png"
func GetUrlFileType(u string) (string, error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(filepath.Ext(parsedUrl.Path), "."), nil
}
