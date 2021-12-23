package app

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	router = mux.NewRouter()
	
)

func StartApplication(){

	fmt.Println("start application!")
	mapUrls()
	srv := &http.Server{
		Handler: router,
		Addr: "localhost:8080",
	}

	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}

}

