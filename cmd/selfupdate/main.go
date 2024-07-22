package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
)

const (
	stylesSourceFile = "cmd/gendoc/styles.go"
)

func fetchStyles() (string, error) {
	cmd := exec.Command("sh", "-c", `curl -s https://github.com/alecthomas/chroma/tree/master/styles | grep "application/json" | tr '{' '\n' | grep styles | cut -d'"' -f4 | grep \.xml | cut -d. -f1`)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func extractStyleNames(cmdOutput string) []string {
	styles := strings.Split(cmdOutput, "\n")
	var styleNames []string
	for _, style := range styles {
		if style != "" {
			styleNames = append(styleNames, style)
		}
	}
	return styleNames
}

func generateGoSourceCode(styles []string) string {
	var buffer bytes.Buffer
	buffer.WriteString("package main\n\nvar styles = []string{\n")
	for _, style := range styles {
		buffer.WriteString(fmt.Sprintf("\t\"%s\",\n", style))
	}
	buffer.WriteString("}\n")
	return buffer.String()
}

func main() {
	cmdOutput, err := fetchStyles()
	if err != nil {
		log.Fatalf("error fetching styles: %v", err)
	}

	styleNames := extractStyleNames(cmdOutput)

	sourceCode := generateGoSourceCode(styleNames)

	fmt.Println(sourceCode)

	err = ioutil.WriteFile(stylesSourceFile, []byte(sourceCode), 0644)
	if err != nil {
		log.Fatalf("error writing source code to file: %v", err)
	}
}
