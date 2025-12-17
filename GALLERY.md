For manually running the code that generates the Style Gallery, try:

    go run cmd/selfupdate/*.go
    go run cmd/gendoc/*.go

`selfupdate` updates `cmd/gendoc/style.go` and `gendoc` generates the HTML gallery in `docs`.
