// splash adds a dash of color to embedded source code in HTML
package splash

import (
	"bytes"
	"errors"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/yhat/scrape"
	gohtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"strings"
	"unicode"
)

var (
	errHEAD = errors.New("HTML should contain <head> or <html> when adding CSS")
)

// AddCSStoHTML takes htmlData and adds cssData in a <style> tag.
// Returns an error if <head> or <html> does not already exists.
func AddCSSToHTML(htmlData, cssData []byte) ([]byte, error) {
	if bytes.Contains(htmlData, []byte("</head>")) {
		var buf bytes.Buffer
		buf.WriteString("<head>\n    <style>")
		buf.Write(cssData)
		buf.WriteString("    </style>")
		return bytes.Replace(htmlData, []byte("<head>"), buf.Bytes(), 1), nil
	} else if bytes.Contains(htmlData, []byte("<html>")) {
		var buf bytes.Buffer
		buf.WriteString("<html>\n  <head>\n  <style>")
		buf.Write(cssData)
		buf.WriteString("    </style>\n  </head>")
		return bytes.Replace(htmlData, []byte("<html>"), buf.Bytes(), 1), nil
	} else {
		return []byte{}, errHEAD
	}
}

// Splash takes HTML code as bytes and tries to syntax highlight
// code between <pre></pre>, <pre><code></code></pre>, <code><pre></pre></code> or <code></code>.
// "style" is a syntax highlight style, like "monokai".
// Full style list here: https://github.com/alecthomas/chroma/tree/master/styles
// Returns the modified HTML source code with embedded CSS as a <style> tag.
// Requires the given HTML to contain </head> or <html>.
func Splash(htmlData []byte, styleName string) ([]byte, error) {

	// Create a byte slice used for changing the HTML code when adding
	// syntax highlight tags and style
	mutableBytes := make([]byte, len(htmlData))
	copy(mutableBytes, htmlData)

	// Parse the given HTML
	root, err := gohtml.Parse(bytes.NewReader(htmlData))
	if err != nil {
		return []byte{}, err
	}

	// Find all <code> and <pre> tags
	matcher := func(n *gohtml.Node) bool {
		return n.DataAtom == atom.Code || n.DataAtom == atom.Pre
	}
	allCodeTags := scrape.FindAll(root, matcher)

	var cssbuf bytes.Buffer // buffer for generated CSS

	// Extract, syntax highlight and place back all snippets of code in the given HTML data
	for _, codeTag := range allCodeTags {

		sourceCode := scrape.TextJoin(codeTag, func(s []string) string {
			return strings.TrimRightFunc(strings.Join(s, ""), unicode.IsSpace)
		})

		// Try to identify the language
		lexer := lexers.Analyse(sourceCode)
		if lexer == nil {
			// Could not identify the language
			lexer = lexers.Fallback
		}
		// Combine token runs
		lexer = chroma.Coalesce(lexer)

		// Try to use the given style name
		style := styles.Get(styleName)
		if style == nil {
			// Could not use the given style name
			style = styles.Fallback
		}

		formatter := html.New(html.WithClasses())
		if formatter == nil {
			return []byte{}, err
		}

		err = formatter.WriteCSS(&cssbuf, style)
		if err != nil {
			return []byte{}, err
		}

		var highlightedHTML bytes.Buffer // tmp buf for the generated syntax highlighted HTML

		// Format the sourceCode and write the new markup to highlightedHTML
		iterator, err := lexer.Tokenise(nil, sourceCode)
		if err != nil {
			return []byte{}, err
		}
		err = formatter.Format(&highlightedHTML, style, iterator)
		if err != nil {
			return []byte{}, err
		}

		// Replace the non-highlighted code with highlighted code
		mutableBytes = bytes.Replace(mutableBytes, []byte(sourceCode), highlightedHTML.Bytes(), 1)
	}

	// Add all the generated CSS to a <style> tag in the generated HTML
	htmlBytes, err := AddCSSToHTML(mutableBytes, cssbuf.Bytes())
	if err != nil {
		return []byte{}, err
	}

	return htmlBytes, nil
}
