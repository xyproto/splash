package main

import (
	"github.com/xyproto/splash"
	"io/ioutil"
)

func main() {
	htmlData, err := ioutil.ReadFile("input.html")
	if err != nil {
		panic(err)
	}

	// Highlight the source code in the HTML with the monokai style
	htmlBytes, err := splash.Splash(htmlData, "monokai")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("output.html", htmlBytes, 0644)
	if err != nil {
		panic(err)
	}
}
