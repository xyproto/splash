package splash

import (
	"bytes"
	"fmt"
	"github.com/russross/blackfriday"
	"strings"
	"testing"
)

const (
	simpleCSS = "body { font-family: sans-serif; margin: 4em; } .chroma { padding: 1em; }"

	simpleMarkdown = `Hi there

This is some code:

    abc 123

And some more:

    package main

    import "fmt"

    func main() {
      fmt.Println("ost")
    }

And some more:

    print("hi")
    l = [x for x in range(10) if x > 5]

Like that.

And now some ` + "`inline stuff`."
)

func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

func TestTags(t *testing.T) {

	// Generate a HTML document for the github style
	styleName := "github"
	title := "Testing"

	var inputBuffer bytes.Buffer
	inputBuffer.WriteString("<!doctype html><html><head><title>")
	inputBuffer.WriteString(title)
	inputBuffer.WriteString("</title><style>")
	inputBuffer.WriteString(simpleCSS)
	inputBuffer.WriteString("</style></head><body><h1>")
	inputBuffer.WriteString(title)
	inputBuffer.WriteString("</h1><code><pre>")
	inputBuffer.WriteString(`
package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`)
	inputBuffer.WriteString("</pre></code></body></html>")

	// Highlight the source code in the HTML with the current style
	htmlBytes, err := Splash(inputBuffer.Bytes(), styleName)
	if err != nil {
		panic(err)
	}

	input := inputBuffer.String()
	output := string(htmlBytes)

	inputPreCount := strings.Count(input, "<pre")
	inputCodeCount := strings.Count(input, "<code")

	outputPreCount := strings.Count(output, "<pre")
	outputCodeCount := strings.Count(output, "<code")

	assertEqual(t, inputPreCount, outputPreCount, "<pre count differs")
	assertEqual(t, inputCodeCount, outputCodeCount, "<code count differs")
}

func TestMarkdown(t *testing.T) {
	var inputBuffer bytes.Buffer

	inputBuffer.WriteString("<!doctype html><html><head><title>Markdown</title></head><body>")

	// Convert Markdown to HTML
	inputBuffer.Write(blackfriday.MarkdownCommon([]byte(simpleMarkdown)))

	inputBuffer.WriteString("</body></html>")

	// Highlight the source code in the HTML with the current style
	htmlBytes, err := Splash(inputBuffer.Bytes(), "github")
	if err != nil {
		panic(err)
	}

	input := inputBuffer.String()
	output := string(htmlBytes)

	fmt.Println("--- INPUT ---")
	fmt.Println(input)
	fmt.Println("--- OUTPUT ---")
	fmt.Println(output)

	//inputPreCount := strings.Count(input, "<pre")
	//inputCodeCount := strings.Count(input, "<code")

	//outputPreCount := strings.Count(output, "<pre")
	//outputCodeCount := strings.Count(output, "<code")

	//assertEqual(t, inputPreCount, outputPreCount, "<pre count differs")
	//assertEqual(t, inputCodeCount, outputCodeCount, "<code count differs")
}
