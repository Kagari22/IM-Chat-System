package main

import (
	"log"
	"net/http"

	"IM_Chat_System/internal/app"
)

// .\scripts\Start-Dev.ps1 -WithInfra

func main() {
	a, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	log.Println("server listening on", a.Config.HTTPAddr)
	log.Fatal(http.ListenAndServe(a.Config.HTTPAddr, a.Router()))
}
