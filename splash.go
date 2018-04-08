// Package splash adds a dash of color to embedded source code in HTML
package splash

import (
	"bytes"
	"errors"
	"html"
	"regexp"
	"unicode"

	"github.com/alecthomas/chroma"
	chromaHTML "github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/lexers/c"
	"github.com/alecthomas/chroma/styles"
)

var (
	errHEAD = errors.New("HTML should contain </head> or <html> when adding CSS")
)

// AddCSSToHTML takes htmlData and adds cssData in a <style> tag.
// Returns an error if </head> or <html> does not already exists.
// Tries to add CSS as late in <head> as possible.
func AddCSSToHTML(htmlData, cssData []byte) ([]byte, error) {
	if bytes.Contains(htmlData, []byte("</head>")) {
		var buf bytes.Buffer
		buf.WriteString("<style>")
		buf.Write(cssData)
		buf.WriteString("\n    </style>\n  </head>")
		return bytes.Replace(htmlData, []byte("</head>"), buf.Bytes(), 1), nil
	} else if bytes.Contains(htmlData, []byte("<html>")) {
		var buf bytes.Buffer
		buf.WriteString("<html>\n  <head>\n    <style>")
		buf.Write(cssData)
		buf.WriteString("\n    </style>\n  </head>")
		return bytes.Replace(htmlData, []byte("<html>"), buf.Bytes(), 1), nil
	} else {
		return []byte{}, errHEAD
	}
}

// Splash takes HTML code as bytes and tries to syntax highlight code between
// <pre> and </pre> tags.
//
// "style" is a syntax highlight style, like "monokai".
//
// Full style list here: https://github.com/alecthomas/chroma/tree/master/styles
//
// Returns the modified HTML source code with embedded CSS as a <style> tag.
// Requires the given HTML to contain </head> or <html>.
//
// unescape can be set to true for unescaping already escaped code in <pre> tags,
// which can be useful when highlighting code in newly rendered markdown.
func Splash(htmlData []byte, styleName string, unescape bool) ([]byte, error) {

	// Try to use the given style name
	style := styles.Get(styleName)
	if style == nil {
		// Could not use the given style name
		style = styles.Fallback
	}

	// Create a HTML formatter
	formatter := chromaHTML.New(chromaHTML.WithClasses())
	if formatter == nil {
		return []byte{}, errors.New("Unable to instanciate chroma HTML formatter")
	}

	var (
		cssBuf       bytes.Buffer // buffer for generated CSS
		mutableBytes = htmlData[:]
		outerErr     error
	)

	// Replace the non-highlighted code with highlighted code
	re := regexp.MustCompile(`(?m)(?s)(<pre>|<pre .*?chroma.*?>)(.*?)(</pre>)`)
	mutableBytes = re.ReplaceAllFunc(mutableBytes, func(preSource []byte) []byte {

		strippedPreTag1 := false
		if bytes.HasPrefix(preSource, []byte("<pre>")) && bytes.HasSuffix(preSource, []byte("</pre>")) {
			// Remove leading and trailing pre tags
			preSource = preSource[5 : len(preSource)-6]
			strippedPreTag1 = true
		}

		strippedCodeTag := false
		if bytes.HasPrefix(preSource, []byte("<code>")) && bytes.HasSuffix(preSource, []byte("</code>")) {
			// Remove leading and trailing pre tags
			preSource = preSource[6 : len(preSource)-7]
			strippedCodeTag = true
		}

		strippedPreTag2 := false
		if bytes.HasPrefix(preSource, []byte("<pre>")) && bytes.HasSuffix(preSource, []byte("</pre>")) {
			// Remove leading and trailing pre tags
			preSource = preSource[5 : len(preSource)-6]
			strippedPreTag2 = true
		}

		// Check if something like <code class="language-c"> has been specified
		language := ""
		strippedLongerCodeTag := false
		if bytes.HasPrefix(preSource, []byte(`<code class="language-`)) && bytes.Count(preSource, []byte(`"`)) >= 2 {
			language = string(bytes.SplitN((bytes.SplitN(preSource, []byte(`"`), 3)[1]), []byte("-"), 2)[1])
			// Then strip the longer tag, if possible
			if bytes.HasPrefix(preSource, []byte(`<code class="language-`+language+`">`)) && bytes.HasSuffix(preSource, []byte("</code>")) {
				// Remove leading and trailing pre tags
				preSource = preSource[len(language)+24 : len(preSource)-7]
				strippedLongerCodeTag = true
			}
		}

		// From bytes to string, while trimming away whitespace from only the end of the string.
		// There may be wanted indentation at the beginning of the string.
		preSourceString := string(bytes.TrimRightFunc(preSource, unicode.IsSpace))

		// Unescape HTML, like &amp;, if this has already been done by ie. a Markdown renderer
		if unescape {
			preSourceString = html.UnescapeString(preSourceString)
		}

		// Try to find a suitable lexer
		var lexer chroma.Lexer
		if language != "" {
			// Try to use the specified language
			lexer = lexers.Get(language)
		}
		if lexer == nil {
			// Try to identify the language based on the source code that is to be highlighted
			lexer = lexers.Analyse(preSourceString)
		}
		if lexer == nil {
			// Could not identify the language, use the lexer for C by default (the laternative is lexers.Fallback)
			lexer = c.C
		}

		// Combine token runs
		lexer = chroma.Coalesce(lexer)

		// Prepare to iterate over the tokens in the source code
		iterator, err := lexer.Tokenise(nil, preSourceString)
		if err != nil {
			outerErr = err
			return []byte{}
		}

		// Write the needed CSS to cssBuf
		err = formatter.WriteCSS(&cssBuf, style)
		if err != nil {
			outerErr = err
			return []byte{}
		}

		// Write the highlightet HTML to the hiBuf buffer
		var hiBuf bytes.Buffer
		err = formatter.Format(&hiBuf, style, iterator)
		if err != nil {
			outerErr = err
			return []byte{}
		}

		// Check that the highlighted bytes have a minimum of information
		hiBytes := hiBuf.Bytes()

		//fmt.Println("GOT POST SOURCE:\n" + string(hiBytes) + "\n\n")

		// <code><pre class="chroma"> gives the best results. <pre> and then <code> renders things wrong here.

		if !strippedPreTag2 {
			// Remove the <pre> tag that was added by chroma
			hlen := len(hiBytes)
			if bytes.HasPrefix(hiBytes, []byte(`<pre class="chroma">`)) && bytes.HasSuffix(hiBytes, []byte("</pre>")) {
				// Remove the leading <pre class="chroma"> and the trailing </pre> tag
				hiBytes = hiBytes[len(`<pre class="chroma">`) : hlen-len("</pre>")]
			}
		}

		if strippedCodeTag || strippedLongerCodeTag {
			// Add the <code> tag again
			hiBytes = []byte("<code>" + string(hiBytes) + "</code>")
		}

		if strippedPreTag1 {
			// Add the <pre> tag
			hiBytes = []byte(`<pre class="chroma">` + string(hiBytes) + "</pre>")
		}

		//fmt.Println("GOT OUTPUT SOURCE:\n" + string(hiBytes) + "\n\n")

		return hiBytes
	})

	if outerErr != nil {
		return []byte{}, outerErr
	}

	// Add all the generated CSS to a <style> tag in the generated HTML
	htmlBytes, err := AddCSSToHTML(mutableBytes, cssBuf.Bytes())
	if err != nil {
		return []byte{}, err
	}

	return htmlBytes, nil
}
