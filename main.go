package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/alexraskin/alexraskin.com/app/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", handlers.MainHandler)

	fmt.Printf("Server started on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
