package splash

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/russross/blackfriday"
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
      fmt.Println("ost", 3 > 1)
    }

And some more:

    print("hi")
    l = [x for x in range(10) if x > 5]

Like that.

And now some ` + "`inline stuff`."

	languageBlock = `<pre><code class="language-c">// Return the version string for the server.
version() -&gt; string

// Sleep the given number of seconds (can be a float).
sleep(number)

// Log the given strings as information. Takes a variable number of strings.
log(...)

// Log the given strings as a warning. Takes a variable number of strings.
warn(...)

// Log the given strings as an error. Takes a variable number of strings.
err(...)

// Return the number of nanoseconds from 1970 (&quot;Unix time&quot;)
unixnano() -&gt; number

// Convert Markdown to HTML
markdown(string) -&gt; string

// Return the directory where the REPL or script is running. If a filename (optional) is given, then the path to where the script is running, joined with a path separator and the given filename, is returned.
scriptdir([string]) -&gt; string

// Return the directory where the server is running. If a filename (optional) is given, then the path to where the server is running, joined with a path separator and the given filename, is returned.
serverdir([string]) -&gt; string
</code></pre>`
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

func TestSimple(t *testing.T) {

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
	fmt.Println("Hello, World!", 3 > 1)
}
`)
	inputBuffer.WriteString("</pre></code></body></html>")

	// Highlight the source code in the HTML with the current style
	htmlBytes, err := Splash(inputBuffer.Bytes(), styleName, false)
	if err != nil {
		panic(err)
	}

	input := inputBuffer.String()
	output := string(htmlBytes)

	//fmt.Println("SIMPLE OUTPUT\n", output)

	err = ioutil.WriteFile("output_test_simple.html", htmlBytes, 0644)
	if err != nil {
		panic(err)
	}

	inputPreCount := strings.Count(input, "<pre")
	inputCodeCount := strings.Count(input, "<code")

	outputPreCount := strings.Count(output, "<pre")
	outputCodeCount := strings.Count(output, "<code")

	assertEqual(t, inputPreCount, outputPreCount, "<pre count differs")
	assertEqual(t, inputCodeCount, outputCodeCount, "<code count differs")
}

func TestMarkdown(t *testing.T) {
	var inputBuffer bytes.Buffer

	//fmt.Println("--- Markdown INPUT ---")
	//fmt.Println(simpleMarkdown)

	inputBuffer.WriteString("<!doctype html><html><head><title>Markdown</title></head><body>")

	// Convert Markdown to HTML
	inputBuffer.Write(blackfriday.MarkdownCommon([]byte(simpleMarkdown)))

	inputBuffer.WriteString("</body></html>")

	// Highlight the source code in the HTML with the current style
	htmlBytes, err := Splash(inputBuffer.Bytes(), "monokai", true)
	if err != nil {
		panic(err)
	}

	input := inputBuffer.String()
	output := string(htmlBytes)

	//fmt.Println("--- HTML INPUT ---")
	//fmt.Println(input)

	//fmt.Println("--- HTML OUTPUT ---")
	//fmt.Println(output)

	err = ioutil.WriteFile("output_test_markdown.html", htmlBytes, 0644)
	if err != nil {
		panic(err)
	}

	inputPreCount := strings.Count(input, "<pre")
	//fmt.Println("INPUT PRE:", inputPreCount)
	inputCodeCount := strings.Count(input, "<code")

	outputPreCount := strings.Count(output, "<pre")
	//fmt.Println("OUTPUT PRE:", outputPreCount)
	outputCodeCount := strings.Count(output, "<code")

	assertEqual(t, inputPreCount, outputPreCount, "<pre count differs")
	assertEqual(t, inputCodeCount, outputCodeCount, "<code count differs")
}

func TestExistingPre(t *testing.T) {

	// Generate a HTML document for the github style
	styleName := "github"
	title := "Testing2"

	var inputBuffer bytes.Buffer
	inputBuffer.WriteString("<!doctype html><html><head><title>")
	inputBuffer.WriteString(title)
	inputBuffer.WriteString("</title><style>")
	inputBuffer.WriteString(simpleCSS)
	inputBuffer.WriteString("</style></head><body><h1>")
	inputBuffer.WriteString(title)
	inputBuffer.WriteString(`</h1><code><pre class="existing">`)
	inputBuffer.WriteString(`
package main

import "fmt"

func main() {
	fmt.Println("Hello, World!", 3 > 1)
}
`)
	inputBuffer.WriteString("</pre></code></body></html>")

	// Highlight the source code in the HTML with the current style
	htmlBytes, err := Splash(inputBuffer.Bytes(), styleName, false)
	if err != nil {
		panic(err)
	}

	output := string(htmlBytes)

	//fmt.Println("TEST OUTPUT\n", output)

	if strings.Contains(output, `pre class="chroma"`) {
		t.Fatal("pre code with existing class should not be formatted with chroma!")
	}
}

func TestLanguageBlock(t *testing.T) {

	// Generate a HTML document for the github style
	styleName := "dracula"
	title := "Language Test"

	var inputBuffer bytes.Buffer
	inputBuffer.WriteString("<!doctype html><html><head><title>")
	inputBuffer.WriteString(title)
	inputBuffer.WriteString("</title><style>")
	inputBuffer.WriteString(simpleCSS)
	inputBuffer.WriteString("</style></head><body><h1>")
	inputBuffer.WriteString(title)
	inputBuffer.WriteString("</h1>")
	inputBuffer.WriteString(languageBlock)
	inputBuffer.WriteString("</body></html>")

	// Highlight the source code in the HTML with the current style
	htmlBytes, err := Splash(inputBuffer.Bytes(), styleName, true)
	if err != nil {
		panic(err)
	}

	input := inputBuffer.String()
	output := string(htmlBytes)

	//fmt.Println("LANGUAGE OUTPUT\n", output)

	err = ioutil.WriteFile("output_test_language.html", htmlBytes, 0644)
	if err != nil {
		panic(err)
	}

	inputPreCount := strings.Count(input, "<pre")
	inputCodeCount := strings.Count(input, "<code")

	outputPreCount := strings.Count(output, "<pre")
	outputCodeCount := strings.Count(output, "<code")

	assertEqual(t, inputPreCount, outputPreCount, "<pre count differs")
	assertEqual(t, inputCodeCount, outputCodeCount, "<code count differs")
}
