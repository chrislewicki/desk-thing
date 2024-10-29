package main

import (
	"fmt"
	"net/http"
)

func dataHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This will be our data endpoint.")
}

func main() {
	http.HandleFunc("/data", dataHandler)

	fmt.Println("Server is running on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
