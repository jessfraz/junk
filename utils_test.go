package main

import "testing"

func testIsSigned(t *testing.T) {
	commits := map[string]bool{
		"https://github.com/jfrazelle/docker/commit/51589a13dcb14f8618610880b693fbcafcc6b08c.patch": true,
		"https://github.com/docker/docker/commit/0277dcdacabf03820b0a544a1fcdc5971ca89667.patch":    true,
		"https://github.com/docker/docker/commit/f8dcb983e11044450a40e00aa001b109f6c187c2.patch":    true,
	}

	for url, expected := range commits {
		if isSigned(url) != expected {
			t.Fatalf("Expected %s isSigned to be %#v & was not", url, expected)
		}
	}
}
