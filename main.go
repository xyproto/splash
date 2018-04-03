package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/yhat/scrape"
	gohtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// Placeholder function, should use the function in Algernon instead
func AddCSSToHTML(htmlData, cssData []byte) ([]byte, error) {
	if !bytes.Contains(htmlData, []byte("</head>")) {
		return []byte{}, errors.New("HTML must contain </head> when adding CSS")
	}
	var buf bytes.Buffer
	buf.WriteString("<style>")
	buf.Write(cssData)
	buf.WriteString("</style></head>")
	return bytes.Replace(htmlData, []byte("</head>"), buf.Bytes(), 1), nil
}

func TextJoiner(s []string) string {
	return strings.Join(s, "")
}

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

	root, err := gohtml.Parse(fileReader2)
	if err != nil {
		log.Fatal(err)
	}

	matcher := func(n *gohtml.Node) bool {
		if n.DataAtom == atom.Code {
			fmt.Println("FOUND CODE")
			return true
		}
		if n.DataAtom == atom.Pre {
			fmt.Println("FOUND PRE")
			return true
		}
		return false
	}

	var buf bytes.Buffer    // tmp buf for generated syntax highlighted HTML
	var cssbuf bytes.Buffer // tmp buf for generated CSS

	allCodeTags := scrape.FindAll(root, matcher)
	for i, codeTag := range allCodeTags {
		fmt.Printf("--- NUM: %d ---\n\n", i+1)
		sourceCode := scrape.TextJoin(codeTag, TextJoiner)

		lexer := lexers.Analyse(sourceCode)
		if lexer == nil {
			// Could not identify the language
			lexer = lexers.Fallback
		}

		style := styles.Get("swapoff") // monokai
		if style == nil {
			// Could not use the given style
			style = styles.Fallback
		}

		formatter := html.New(html.WithClasses())
		if formatter == nil {
			log.Fatal(err)
		}

		err = formatter.WriteCSS(&cssbuf, style)
		if err != nil {
			log.Fatal(err)
		}

		iterator, err := lexer.Tokenise(nil, sourceCode)

		err = formatter.Format(&buf, style, iterator)
		if err != nil {
			log.Fatal(err)
		}

		//fmt.Println("--- FROM ---\n" + sourceCode)
		//fmt.Println("--- TO ---\n" + buf.String())

		mutableBytes = bytes.Replace(mutableBytes, []byte(sourceCode), buf.Bytes(), 1)
	}

	htmlBytes, err := AddCSSToHTML(mutableBytes, cssbuf.Bytes())
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
