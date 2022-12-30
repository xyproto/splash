package main

import (
	"bytes"
	"fmt"
	"github.com/xyproto/splash"
	"io/ioutil"
	"os"
	"time"
)

const (
	title         = "Chroma Style Gallery"
	simpleCSS     = "body { font-family: sans-serif; margin: 4em; } .chroma { padding: 1em; } #main-headline { border-bottom: 3px solid red; margin-bottom: 2em; } a { color: #1E385B; } a:visited { color: #1E385B; } a:hover { color: #4682B4; }"
	sampleContent = `<code><pre>
package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
</pre></code>`
)

func footer() string {
	var buf bytes.Buffer
	buf.WriteString("<small>Generated ")
	buf.WriteString(time.Now().UTC().Format(time.RFC3339)[:10])
	buf.WriteString(" by <a href=\"https://github.com/xyproto\">xyproto</a> using <a href=\"https://github.com/xyproto/splash\">splash</a>.")
	buf.WriteString("</small></body></html>")
	return buf.String()
}

func generateGallery(sampleContent, dirname string) {

	// Generate a HTML page per style
	for i, styleName := range styles {

		// Generate a HTML document for the current style name
		var inputBuffer bytes.Buffer
		inputBuffer.WriteString("<!doctype html><html><head><title>")
		inputBuffer.WriteString(styleName)
		inputBuffer.WriteString("</title><style>")
		inputBuffer.WriteString(simpleCSS + " a { text-decoration: none; }  a:hover { color: #4682B4; }")
		inputBuffer.WriteString("</style></head><body>")
		inputBuffer.WriteString("<h1>")
		inputBuffer.WriteString("<a alt='View " + styleName + " on a page with all the styles' href='all.html#" + styleName + "'>" + styleName + "</a>")
		inputBuffer.WriteString("</h1>")
		inputBuffer.WriteString(sampleContent)

		// Button to the previous style, if possible
		if i > 0 {
			prevName := styles[i-1]
			inputBuffer.WriteString("<button onClick=\"location.href='" + prevName + ".html'\">Prev</button>")
		} else {
			inputBuffer.WriteString("<button disabled='true'>Prev</button>")
		}

		// Button to the next style, if possible
		if i < (len(styles) - 1) {
			nextName := styles[i+1]
			inputBuffer.WriteString("<button onClick=\"location.href='" + nextName + ".html'\">Next</button>")
		} else {
			inputBuffer.WriteString("<button disabled='true'>Next</button>")
		}

		// Button to the to of the single page with all styles
		inputBuffer.WriteString("<button onClick=\"location.href='all.html'\">All</button>")

		// Button to the overview
		inputBuffer.WriteString("<button onClick=\"location.href='index.html'\">Overview</button>")

		inputBuffer.WriteString("</body></html>")

		// Highlight the source code in the HTML with the current style
		htmlBytes, err := splash.Splash(inputBuffer.Bytes(), styleName)
		if err != nil {
			panic(err)
		}

		// Write the HTML sample for this style name
		err = ioutil.WriteFile(dirname+"/"+styleName+".html", htmlBytes, 0644)
		if err != nil {
			panic(err)
		}

	}

	// Generate an Index file for listing the names of all the different styles
	var buf bytes.Buffer
	buf.WriteString("<!doctype html><html><head><title>")
	buf.WriteString(title)
	buf.WriteString("</title><style>")
	buf.WriteString(simpleCSS)
	buf.WriteString("</style></head><body><h1 id='main-headline'>")
	buf.WriteString(title)
	buf.WriteString("</h1><ul>")
	for _, styleName := range styles {
		buf.WriteString("<li><a href=\"" + styleName + ".html\" alt=\"" + styleName + " style\">" + styleName + "</a></li>")
	}
	buf.WriteString("</ul>")
	buf.WriteString("<p><a href='all.html' alt='All styles on one page'>All styles on one page</a></p>")
	if dirname == "." {
		buf.WriteString("<p><a href='longer/index.html' alt='Gallery with longer code samples'>Gallery with longer code samples</a></p>")
	} else {
		buf.WriteString("<p><a href='../index.html' alt='Gallery with shorter code samples'>Gallery with shorter code samples</a></p>")
	}
	buf.WriteString(footer())
	err := ioutil.WriteFile(dirname+"/"+"index.html", buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}

	// Generate a single page with all the styles
	buf = bytes.Buffer{}
	buf.WriteString("<!doctype html><html><head><title>")
	buf.WriteString(title)
	buf.WriteString("</title><style>")
	css := "body { font-family: sans-serif; margin: 4em; } h1 { border-bottom: 3px solid red; margin-bottom: 2em; } pre { display: inline-block; margin: 2em; padding: 3em 5em 3em 2em; box-shadow: 5px 5px 5px rgba(68, 68, 68, 0.6); border-radius: 7px; border: 2px solid black;} #stylelink:link { text-decoration: none; color: black; } #stylelink:visited { color: black; } #stylelink:hover { color: #4682B4; } a { color: #1E385B; }"
	buf.WriteString(css)
	buf.WriteString("</style></head><body><h1>")
	buf.WriteString(title)
	buf.WriteString("</h1>")
	for _, styleName := range styles {
		buf.WriteString("<a name='" + styleName + "'>") // HTML anchor
		buf.WriteString("<h2>")
		buf.WriteString("<a id='stylelink' href='" + styleName + ".html' alt='View only " + styleName + "'>" + styleName + "</a>")
		buf.WriteString("</h2>")

		htmlBytes, cssBytes, err := splash.Highlight([]byte(sampleContent), styleName, false)
		if err != nil {
			panic(err)
		}
		cssElements := bytes.Split(cssBytes, []byte("}"))
		for _, element := range cssElements {
			elements := bytes.Split(bytes.Replace(element, []byte(".chroma"), []byte(""), -1), []byte("{"))
			if len(elements) == 2 {
				className := bytes.TrimSpace(elements[0])
				if bytes.HasPrefix(className, []byte(".")) {
					className = className[1:]
				}
				style := bytes.TrimSpace(elements[1])
				if len(className) == 0 {
					htmlBytes = bytes.Replace(htmlBytes, []byte("<pre "), []byte("<pre style=\""+string(style)+"\" "), -1)
				} else {
					htmlBytes = bytes.Replace(htmlBytes, []byte("<span class=\""+string(className)+"\""), []byte("<span style=\""+string(style)+"\""), -1)
				}
			}
		}
		buf.WriteString("</a>") // HTML anchor
		buf.Write(htmlBytes)
	}
	buf.WriteString("<p><a id='bottom' href='index.html' alt='Back to overview'>Back to overview</a></p>")
	if dirname == "." {
		buf.WriteString("<p><a id='bottom' href='longer/index.html' alt='Go to the other gallery'>Gallery with longer code samples</a></p>")
	} else {
		buf.WriteString("<p><a id='bottom' href='../index.html' alt='Go to the other gallery'>Gallery with shorter code samples</a></p>")
	}
	buf.WriteString(footer())
	err = ioutil.WriteFile(dirname+"/all.html", buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}

}

func main() {
	if err := os.Chdir("../../docs"); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	generateGallery(sampleContent, ".")
	// TODO: Create directory first
	generateGallery(longerSampleContent, "longer")
}
