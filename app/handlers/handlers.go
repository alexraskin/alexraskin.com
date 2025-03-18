package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/alexraskin/alexraskin.com/app/utils"
)

func MainHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	userAgent := r.Header.Get("User-Agent")
	if strings.Contains(strings.ToLower(userAgent), "curl") || strings.Contains(strings.ToLower(userAgent), "httpie") {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, utils.GetCard())
		return
	}

	htmlContent, err := os.ReadFile("public/index.html")
	if err != nil {
		log.Printf("Error reading index.html: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, string(htmlContent))
}
