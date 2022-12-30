package main

// Inspired by ChatGPT

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	stylesSourceFile = "../gendoc/styles.go"
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

func writeStylesToFile(styleNames []string, filePath string) error {
	// Open the file for writing
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the double quoted style names to the file
	for _, styleName := range styleNames {
		if _, err := fmt.Fprintf(f, "\"%s\",\n", styleName); err != nil {
			return err
		}
	}

	return nil
}

func extractStyleNames(htmlContent string) []string {
	// Use a regular expression to find all the links to XML files
	re := regexp.MustCompile(`styles/.*\.xml`)
	matches := re.FindAllString(htmlContent, -1)

	// Extract the style names from the matches
	styleNames := make([]string, len(matches))
	for i, match := range matches {
		// Find the last '/' and the last '.'
		lastSlashIdx := strings.LastIndex(match, "/")
		lastDotIdx := strings.LastIndex(match, ".")
		// Extract the style name from the match
		styleName := match[lastSlashIdx+1 : lastDotIdx]
		// Remove everything after the last period
		styleName = strings.TrimSuffix(styleName, styleName[strings.LastIndex(styleName, "."):])
		styleNames[i] = styleName
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
