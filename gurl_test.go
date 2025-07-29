package gurl

import (
	"testing"
)

func TestGetQueryParam(t *testing.T) {
	result, err := GetQueryParam("http://example.com/?t1=1&t2=2", "t1")
	if err != nil || result != "1" {
		t.Errorf("GetQueryParam was incorrect, got: %s, want: %s.", result, "1")
	}
}

func TestSetQueryParam(t *testing.T) {
	result, err := SetQueryParam("http://example.com/?t1=1&t2=2", "t1", "3")
	if err != nil || result != "http://example.com/?t1=3&t2=2" {
		t.Errorf("SetQueryParam was incorrect, got: %s, want: %s.", result, "http://example.com/?t1=3&t2=2")
	}
}

func TestDelQueryParam(t *testing.T) {
	result, err := DelQueryParam("http://example.com/?t1=1&t2=2", "t1")
	if err != nil || result != "http://example.com/?t2=2" {
		t.Errorf("DelQueryParam was incorrect, got: %s, want: %s.", result, "http://example.com/?t2=2")
	}
}

func TestGetHashParam(t *testing.T) {
	type HashTest struct {
		url    string
		param  string
		result string
	}
	tests := []HashTest{
		{"http://example.com/#?t1=1&t2=2", "t1", "1"},
		{"http://example.com/?t1=1&t2=2", "t1", ""},
		{"http://example.com/p/?id=3#?t1=1&t2=2", "t1", "1"},
		{"http://example.com/#?t1=1&t2=2", "t3", ""},
		{"http://example.com/?p1=233#path", "t3", ""},
		{"http://example.com/path/subp?233#?p1=1&t3=333", "t3", "333"},
		{"http://example.com/?233#?t3=3&t4=&t5", "t4", ""},
		{"http://example.com/?233#path?t3=3", "t3", "3"},
	}
	for _, test := range tests {
		result, err := GetHashParam(test.url, test.param)
		if err != nil || result != test.result {
			t.Errorf("GetHashParam was incorrect, got: %s, want: %s.", result, test.result)
		}
	}
}

func TestSetHashParam(t *testing.T) {
	type HashTest struct {
		url    string
		param  string
		value  string
		result string
	}
	tests := []HashTest{
		{"http://example.com/#?ssss&p1=1&p2=2&000", "p1", "3", "http://example.com/#?ssss&p1=3&p2=2&000"},
		{"http://example.com/#?ssss&p1=&p2=2&000", "p1", "3", "http://example.com/#?ssss&p1=3&p2=2&000"},
		{"http://example.com/#?t1=1&t2=2", "t1", "3", "http://example.com/#?t1=3&t2=2"},
		{"http://example.com/?t1=1&t2=2", "t1", "3", "http://example.com/?t1=1&t2=2#?t1=3"},
		{"http://example.com/p/?id=3#?t1=1&t2=2", "t1", "3", "http://example.com/p/?id=3#?t1=3&t2=2"},
		{"http://example.com/#?t1=1&t2=2", "t3", "3", "http://example.com/#?t1=1&t2=2&t3=3"},
		{"http://example.com/?p1=233", "t3", "3", "http://example.com/?p1=233#?t3=3"},
		{"http://example.com/path/subp?233", "t3", "3", "http://example.com/path/subp?233#?t3=3"},
		{"http://example.com/?233#?t3=3", "t4", "4", "http://example.com/?233#?t3=3&t4=4"},
		{"http://example.com/?233#path?t3=3", "t3", "4", "http://example.com/?233#path?t3=4"},
		{"http://example.com/?233#?p3=3&p4=4", "p3", "4", "http://example.com/?233#?p3=4&p4=4"},
		{"http://example.com/?233#p3=3&p4=4", "p3", "4", "http://example.com/?233#p3=3&p4=4?p3=4"},
	}
	for _, test := range tests {
		result, err := SetHashParam(test.url, test.param, test.value)
		if err != nil || result != test.result {
			t.Errorf("SetHashParam was incorrect, got: %s, want: %s.", result, test.result)
		}
	}
}

func TestDelHashParam(t *testing.T) {
	type HashTest struct {
		url    string
		param  string
		result string
	}
	tests := []HashTest{
		{"http://example.com/#?p1=1", "p1", "http://example.com/"},
		{"http://example.com/#?t1=1&t2=2", "t1", "http://example.com/#?t2=2"},
		{"http://example.com/?t1=1&t2=2", "t1", "http://example.com/?t1=1&t2=2"},
		{"http://example.com/p/?id=3#?t1=1&t2=2", "t1", "http://example.com/p/?id=3#?t2=2"},
		{"http://example.com/#?t1=1&t2=2", "t3", "http://example.com/#?t1=1&t2=2"},
		{"http://example.com/?p1=233", "t3", "http://example.com/?p1=233"},
		{"http://example.com/path/subp?233", "t3", "http://example.com/path/subp?233"},
		{"http://example.com/?233#?t3=3", "t4", "http://example.com/?233#?t3=3"},
		{"http://example.com/?233#path?t3=3", "t3", "http://example.com/?233#path"},
		{"http://example.com/?233#?p3=3&p4=4", "p3", "http://example.com/?233#?p4=4"},
		{"http://example.com/?233#?p3=3&p4=4", "p3", "http://example.com/?233#?p4=4"},
	}
	for _, test := range tests {
		result, err := DelHashParam(test.url, test.param)
		if err != nil || result != test.result {
			t.Errorf("DelHashParam was incorrect, got: %s, want: %s.", result, test.result)
		}
	}
}

func TestGetPath(t *testing.T) {
	result, err := GetPath("http://example.com/path/to/resource")
	if err != nil || result != "/path/to/resource" {
		t.Errorf("GetPath was incorrect, got: %s, want: %s.", result, "/path/to/resource")
	}
}

func TestSetPath(t *testing.T) {
	result, err := SetPath("http://example.com/path/to/resource", "/new/path")
	if err != nil || result != "http://example.com/new/path" {
		t.Errorf("SetPath was incorrect, got: %s, want: %s.", result, "http://example.com/new/path")
	}
}

func TestGetHost(t *testing.T) {
	result, err := GetHost("http://example.com/path/to/resource")
	if err != nil || result != "example.com" {
		t.Errorf("GetHost was incorrect, got: %s, want: %s.", result, "example.com")
	}
}

func TestSetHost(t *testing.T) {
	result, err := SetHost("http://example.com/path/to/resource", "newhost.com")
	if err != nil || result != "http://newhost.com/path/to/resource" {
		t.Errorf("SetHost was incorrect, got: %s, want: %s.", result, "http://newhost.com/path/to/resource")
	}
}

func TestGetHostname(t *testing.T) {
	result, err := GetHostname("http://subdomain.example.com/path/to/resource")
	if err != nil || result != "subdomain.example.com" {
		t.Errorf("GetHostname was incorrect, got: %s, want: %s.", result, "subdomain")
	}
}

func TestSetHostname(t *testing.T) {
	result, err := SetHostname("http://subdomain.example.com/path/to/resource", "newsubdomain.example.com")
	if err != nil || result != "http://newsubdomain.example.com/path/to/resource" {
		t.Errorf("SetHostname was incorrect, got: %s, want: %s.", result, "http://newsubdomain.example.com/path/to/resource")
	}
}

func TestGetProtocol(t *testing.T) {
	result, err := GetProtocol("http://example.com/path/to/resource")
	if err != nil || result != "http" {
		t.Errorf("GetProtocol was incorrect, got: %s, want: %s.", result, "http")
	}
}

func TestSetProtocol(t *testing.T) {
	result, err := SetProtocol("http://example.com/path/to/resource", "https")
	if err != nil || result != "https://example.com/path/to/resource" {
		t.Errorf("SetProtocol was incorrect, got: %s, want: %s.", result, "https://example.com/path/to/resource")
	}
}

func TestCheckValid(t *testing.T) {
	if !CheckValid("http://example.com/?#2333?t1=1&t2=2&t3=3&t4=4") {
		t.Errorf("CheckValid was incorrect, got: false, want: true.")
	}
}

func TestCheckValidHTTPURL(t *testing.T) {
	if !CheckValidHTTPURL("http://example.com/?#2333?t1=1&t2=2&t3=3&t4=4") {
		t.Errorf("CheckValidHTTPURL was incorrect, got: false, want: true.")
	}
}

func TestGetURLFileType(t *testing.T) {
	result, err := GetURLFileType("https://example.com/a/b/c.png")
	if err != nil || result != "png" {
		t.Errorf("GetURLFileType was incorrect, got: %s, want: %s.", result, "png")
	}
}

func TestGetBaseURL(t *testing.T) {
	type BaseUrlTest struct {
		url    string
		result string
	}
	tests := []BaseUrlTest{
		{"https://example.com/path/to/resource?param=value#fragment", "https://example.com/path/to/resource"},
		{"http://example.com/?t1=1&t2=2", "http://example.com/"},
		{"https://subdomain.example.com:8080/path?query=test#hash", "https://subdomain.example.com:8080/path"},
		{"http://example.com/path/to/resource", "http://example.com/path/to/resource"},
		{"https://example.com/#fragment", "https://example.com/"},
		{"http://example.com/?query=only", "http://example.com/"},
		{"https://example.com/deep/nested/path?multiple=params&another=value#section", "https://example.com/deep/nested/path"},
		{"http://example.com/path?a=1&b=2#a=1&b=2", "http://example.com/path"},
	}
	for _, test := range tests {
		result, err := GetBaseURL(test.url)
		if err != nil || result != test.result {
			t.Errorf("GetBaseURL was incorrect, got: %s, want: %s.", result, test.result)
		}
	}
}
