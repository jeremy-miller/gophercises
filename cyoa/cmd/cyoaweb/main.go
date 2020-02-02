package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jeremy-miller/gophercises/cyoa"
)

func main() {
	port := flag.Int("port", 3000, "port to start the Choose-Your-Own-Adventure web application on")
	filename := flag.String("file", "gopher.json", "JSON file with the Choose-Your-Own-Adventure story")
	flag.Parse()
	fmt.Printf("Using the story in %s.\n", *filename)

	f, err := os.Open(*filename)
	if err != nil {
		panic(err)
	}

	story, err := cyoa.ParseStory(f)
	if err != nil {
		panic(err)
	}

	h := cyoa.NewHandler(story)
	fmt.Printf("Starting the server at localhost:%d...\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), h))
}
