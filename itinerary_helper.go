package main

import (
	"github.tamu.edu/rmyoung/mashup/keys"
	"database/sql"
)

// DarkSkyForecast from Darksky API
type DarkSkyForecast struct {
	Currently	Currently `json:"currently"`
	Daily     	Daily     `json:"daily"`
	Hourly    	Hourly    `json:"hourly"`
}

// Currently is current weather forecast
type Currently struct {
	Temperature float64 `json:"temperature"`
}

// Hourly is hourly weather forecast
type Hourly struct {
	Summary string `json:"summary"`
}

// Daily is daily weather forecast
type Daily struct {
	Data []DailyData `json:"data"`
}

// DailyData forecast for each individual day within a forecast
type DailyData struct {
	TemperatureHigh	float64 	`json:"temperatureHigh"`
	TemperatureLow 	float64 	`json:"temperatureLow"`
	Summary 				string	`json:"summary"`
}

// Yelp is information returned from Yelp search
type Yelp struct {
	Total			int			`json:"total"`
	Businesses 	[]Business 	`json:"businesses"`
}

// Business info from Yelp Search
type Business struct {
	ImageURL		string		`json:"image_url"`
	IsClosed		bool			`json:"is_closed"`
	Location		Location		`json:"location"`
	Name			string		`json:"name"`
	Price			string		`json:"price"`
	Rating		float64		`json:"rating"`
}

// Location is the address of a business
type Location struct {
	DisplayAddress	[]string	`json:"display_address"`
}

// Bing API Response
type Bing struct {
	ResourceSets		[]ResourceSet 	`json:"resourcesets"`
	ErrorDetails 		[]string			`json:"errordetails"`
}

// ResourceSet returned from Bing API
type ResourceSet struct {
	Resources 	[]Resource 	`json:"resources"`
}

// Resource in list of ResourceSets
type Resource struct {
	Point 		Point			`json:"point"`
}

// Point contains latitude and longitutde within Bing API response
type Point struct {
	Coordinates		[]float64 	`json:"coordinates"`
}


const (
	yelpURL = "https://api.yelp.com/v3/"
	darkSkyURL = "https://api.darksky.net/forecast/"
	bingURL = "http://dev.virtualearth.net/REST/v1/Locations?query="
	bearer = "Bearer " + keys.YelpKey
)

var (
	// ItineraryPlanner is a list of businesses generated from user search
	ItineraryPlanner []Business

	// Itinerary is the user's selected option from ItineraryPlanner results
	Itinerary []string

	// city user wants to create an itinerary for
	city string

	// UserItinerary is itinerary to be returned from database
	UserItinerary []string

	db *sql.DB

	// city user wants to find itineraries for
	itineraryCity string

	// CityAutoComplete is list of cities that have itineraries
	CityAutoComplete []string

	// TagsAutoComplete is list of tags for a specific city
	TagsAutoComplete []string

	// hash map storing list of cities in database
	citiesHashMap = make(map[string] string)
)
