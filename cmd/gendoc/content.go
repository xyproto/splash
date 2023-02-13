package main

import (
	_ "embed"
)

//go:embed example/main.go
var exampleSourceCode string

var longerSampleContent = "<code><pre>" + exampleSourceCode + "</pre></code>"
