package main

import (
	"goweb/app"
	"log"

	"github.com/joho/godotenv"
)

func main(){
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	godotenv.Load(".env")
	app.StartApplication()
}