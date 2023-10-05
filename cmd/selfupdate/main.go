package main

// Inspired by ChatGPT

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

const (
	stylesSourceFile = "cmd/gendoc/styles.go"
	styleListWebPage = "https://github.com/alecthomas/chroma/tree/master/styles"
)

func fetchPage() (string, error) {
	// Fetch the webpage
	resp, err := http.Get(styleListWebPage)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// Read the body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func hasS(xs []string, x string) bool {
	for _, element := range xs {
		if element == x {
			return true
		}
	}
	return false
}

func extractStyleNames(htmlContent string) []string {
	// Use a regular expression to find all the links to XML files
	re := regexp.MustCompile(`styles/.*\.xml`)
	matches := re.FindAllString(htmlContent, -1)

	// Extract the style names from the matches
	var styleNames []string
	if len(matches) == 1 {
		matches = strings.Split(matches[0], "\",\"")
	}
	for _, match := range matches {
		if strings.Contains(match, "\"") {
			fields := strings.Split(match, "\"")
			if len(fields) > 0 {
				match = fields[len(fields)-1]
			}
		}
		if strings.Contains(match, "/") {
			fields := strings.Split(match, "/")
			if len(fields) > 0 {
				match = fields[len(fields)-1]
			}
		}
		if strings.Contains(match, ".") {
			fields := strings.Split(match, ".")
			if len(fields) > 0 {
				match = fields[0]
			}
		}
		styleName := match
		if !hasS(styleNames, styleName) {
			styleNames = append(styleNames, styleName)
		}
	}

	return styleNames
}

func generateGoSourceCode(styles []string) string {
	var buffer bytes.Buffer
	buffer.WriteString("package main\n\nvar styles = []string{\n")
	for _, style := range styles {
		buffer.WriteString("\t\"")
		buffer.WriteString(style)
		buffer.WriteString("\",\n")
	}
	buffer.WriteString("}\n")
	return buffer.String()
}

func main() {
	pageHTML, err := fetchPage()
	if err != nil {
		log.Fatalf("error fetching page: %v", err)
	}

	styleNames := extractStyleNames(pageHTML)

	sourceCode := generateGoSourceCode(styleNames)

	err = ioutil.WriteFile(stylesSourceFile, []byte(sourceCode), 0644)
	if err != nil {
		log.Fatalf("error writing source code to file: %v", err)
	}
}
