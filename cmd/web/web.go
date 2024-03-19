package main

import (
	"net/http"

	"github.com/datasektionen/GOrdian/internal/web"
)

func main() {
	if err := web.Mount(http.DefaultServeMux); err != nil {
		panic(err)
	}
	panic(http.ListenAndServe("0.0.0.0:3000", nil))
}
