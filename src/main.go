package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	router := SetupRouter()
	log.Printf("Starting server on :%d\n", PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", PORT), router))
}
