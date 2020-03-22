package main

import (
	"github.com/xyproto/splash"
	"io/ioutil"
)

func main() {
	// Read "input.html"
	inputHTML, err := ioutil.ReadFile("input.html")
	if err != nil {
		panic(err)
	}

	// Highlight the source code in the HTML document with the monokai style
	outputHTML, err := splash.Splash(inputHTML, "monokai")
	if err != nil {
		panic(err)
	}

	// Write the highlighted HTML to "output.html"
	if err := ioutil.WriteFile("output.html", outputHTML, 0644); err != nil {
		panic(err)
	}
}
