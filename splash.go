// splash adds a dash of color to embedded source code in HTML
package splash

import (
	"bytes"
	"errors"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/yhat/scrape"
	gohtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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

// Style can be "swapoff" or "monokai",
// full style list here: https://github.com/alecthomas/chroma/tree/master/styles
func Splash(contents []byte, style string) ([]byte, error) {
	// TODO: copy?
	mutableBytes := contents

	contentReader := bytes.NewReader(contents)
	root, err := gohtml.Parse(contentReader)
	if err != nil {
		return []byte{}, err
	}

	matcher := func(n *gohtml.Node) bool {
		return n.DataAtom == atom.Code || n.DataAtom == atom.Pre
	}

	var buf bytes.Buffer    // tmp buf for generated syntax highlighted HTML
	var cssbuf bytes.Buffer // tmp buf for generated CSS

	allCodeTags := scrape.FindAll(root, matcher)
	for _, codeTag := range allCodeTags {
		sourceCode := scrape.TextJoin(codeTag, TextJoiner)

		lexer := lexers.Analyse(sourceCode)
		if lexer == nil {
			// Could not identify the language
			lexer = lexers.Fallback
		}

		style := styles.Get(style)
		if style == nil {
			// Could not use the given style
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

		iterator, err := lexer.Tokenise(nil, sourceCode)

		err = formatter.Format(&buf, style, iterator)
		if err != nil {
			return []byte{}, err
		}

		mutableBytes = bytes.Replace(mutableBytes, []byte(sourceCode), buf.Bytes(), 1)
	}

	htmlBytes, err := AddCSSToHTML(mutableBytes, cssbuf.Bytes())
	if err != nil {
		return []byte{}, err
	}

	return htmlBytes, nil
}
