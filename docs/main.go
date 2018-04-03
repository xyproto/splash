package main

import (
	"bytes"
	"github.com/xyproto/splash"
	"io/ioutil"
)

var styles = []string{
	"abap",
	"algol",
	"algol_nu",
	"api",
	"arduino",
	"autumn",
	"borland",
	"bw",
	"colorful",
	"dracula",
	"emacs",
	"friendly",
	"fruity",
	"github",
	"igor",
	"lovelace",
	"manni",
	"monokai",
	"monokailight",
	"murphy",
	"native",
	"paraiso-dark",
	"paraiso-light",
	"pastie",
	"perldoc",
	"pygments",
	"rainbow_dash",
	"rrt",
	"swapoff",
	"tango",
	"trac",
	"vim",
	"vs",
	"xcode",
}

const title = "Chroma Style Gallery"
const simpleCSS = "body { font-family: sans-serif; margin: 4em; }"

func main() {

	for _, styleName := range styles {

		// Generate a HTML document for the current style name
		var inputBuffer bytes.Buffer
		inputBuffer.WriteString("<!doctype html><html><head><title>")
		inputBuffer.WriteString(styleName)
		inputBuffer.WriteString("</title><style>")
		inputBuffer.WriteString(simpleCSS)
		inputBuffer.WriteString("</style></head><body><h1>")
		inputBuffer.WriteString(styleName)
		inputBuffer.WriteString("</h1><code><pre>")
		inputBuffer.WriteString(`
package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`)
		inputBuffer.WriteString("</pre></code>")
		inputBuffer.WriteString("<button onClick=\"history.go(-1)\">Back</button>")
		inputBuffer.WriteString("</body></html>")

		// Highlight the source code in the HTML with the monokai style
		htmlBytes, err := splash.Splash(inputBuffer.Bytes(), styleName)
		if err != nil {
			panic(err)
		}

		// Write the HTML sample for this style name
		err = ioutil.WriteFile(styleName+".html", htmlBytes, 0644)
		if err != nil {
			panic(err)
		}

	}

	// Generate an Index file for viewing the different styles
	var buf bytes.Buffer
	buf.WriteString("<!doctype html><html><head><title>")
	buf.WriteString(title)
	buf.WriteString("</title><style>")
	buf.WriteString(simpleCSS)
	buf.WriteString("</style></head><body><h1>")
	buf.WriteString(title)
	buf.WriteString("</h1><ul>")
	for _, styleName := range styles {
		buf.WriteString("<li><a href=\"" + styleName + ".html\">" + styleName + "</a></li>")
	}
	buf.WriteString("</ul></body></html>")
	err := ioutil.WriteFile("index.html", buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}

}
