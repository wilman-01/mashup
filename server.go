package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"fmt"
	"os"
	"database/sql"
	_"github.com/lib/pq"
)

// creates and returns router
func newRouter() *mux.Router {
	r := mux.NewRouter()

	// Declare the static file directory
	staticFileDirectory := http.Dir("./pages/")

	// "/pages/index.html" --> the file server
	// will look for only "index.html" inside the directory declared above.
	staticFileHandler := http.StripPrefix("/pages/", http.FileServer(staticFileDirectory))

	// The "PathPrefix" method acts as a matcher, and matches all routes starting
	// with "/pages/", instead of the absolute route itself
	r.PathPrefix("/pages/").Handler(staticFileHandler).Methods("GET")

	r.HandleFunc("/location/", getLocationHandler).Methods("GET")
	r.HandleFunc("/location/", createLocationHandler).Methods("POST")

	r.HandleFunc("/itineraryoptions/", getItineraryOptionsHandler).Methods("GET")
	r.HandleFunc("/itineraryoptions/", createItineraryOptionsHandler).Methods("POST")

	r.HandleFunc("/itinerary/", getItineraryHandler).Methods("GET")
	r.HandleFunc("/itinerary/", createItineraryHandler).Methods("POST")

	r.HandleFunc("/finalitinerary/", createFinalizedItineraryHandler).Methods("POST")

	r.HandleFunc("/deleteitinerary/", deleteItineraryHandler).Methods("POST")

	r.HandleFunc("/itinerarymatchesnull/", nullItineraryMatchesHandler).Methods("POST")

	r.HandleFunc("/itineraryspecs/", getUserItineraryHandler).Methods("GET")
	r.HandleFunc("/itineraryspecs/", createUserItineraryHandler).Methods("POST")

	r.HandleFunc("/itinerarylocation/", getItineraryLocationHandler).Methods("GET")
	r.HandleFunc("/itinerarylocation/", createItineraryLocationHandler).Methods("POST")

	r.HandleFunc("/itinerarycities/", getCitiesHandler).Methods("GET")
	r.HandleFunc("/itinerarytags/", getTagsHandler).Methods("GET")

	r.HandleFunc("/newitinerary/", createNewItineraryHandler).Methods("POST")

	r.HandleFunc("/changetagsitinerary/", createChangeTagsItinerary).Methods("POST")

	r.HandleFunc("/itinerarygetnulllocation/", getLocationNullHandler).Methods("POST")
	r.HandleFunc("/itinerarycreatenulllocation/", createLocationNullHandler).Methods("POST")

	return r
}

// get a port for server
func getPort() string {
	port := os.Getenv("PORT")

	// Set a default port if there is nothing in the environment
	if port == "" {
		port = "8080"
		fmt.Println("Navigate to http://localhost:8080/pages/")
	}
	return ":" + port
}

func main() {
	// The router is now formed by calling the `newRouter` constructor function
	r := newRouter()
	connection := os.Getenv("DATABASE_URL")
	if connection == "" {
		connection = "dbname=d63ln1arsv4c0t host=ec2-75-101-128-10.compute-1.amazonaws.com port=5432 user=wzojagsjltnhmc password=0b7ca45fa04ae1b6f450607d704eec9fbbd722c991de119849a18ec021305bfe sslmode=require"
	}

	var err error
	db, err = sql.Open("postgres", connection)
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		panic(err)
	}

	if err = db.Ping(); err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		panic(err)
	}
	fmt.Println("opened")
	defer db.Close()

	fill()
	http.ListenAndServe(getPort(), r)
}