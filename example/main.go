package main

import (
	"github.com/xyproto/splash"
	"io/ioutil"
	"log"
	"os"
)

func process(filename string) {
	fileReader1, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer fileReader1.Close()

	fileReader2, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer fileReader2.Close()

	mutableBytes, err := ioutil.ReadAll(fileReader1)
	if err != nil {
		log.Fatal(err)
	}

	htmlBytes, err := splash.Splash(mutableBytes, "monokai")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("output_"+filename, htmlBytes, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	process("a.html")
	process("b.html")
	process("c.html")
	process("d.html")
}
