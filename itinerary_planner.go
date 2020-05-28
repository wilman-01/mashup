package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"strconv"
	"strings"
	"github.tamu.edu/rmyoung/mashup/keys"
	_"github.com/lib/pq"
	"unicode"
)

func getLocationHandler(w http.ResponseWriter, r *http.Request) {
	cityBytes, err := json.Marshal(city)

	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}	

	w.Write(cityBytes)
}

func createLocationHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	city = r.Form.Get("destination")
	if !isAlphabetic(city) {
		city = "incorrect format"
	}

	http.Redirect(w, r, "/pages/createItinerary.html", http.StatusFound)
}

func getItineraryOptionsHandler(w http.ResponseWriter, r *http.Request) {
	itineraryPlannerListBytes, err := json.Marshal(ItineraryPlanner)

	// check for errors
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(itineraryPlannerListBytes)
}

func createItineraryOptionsHandler(w http.ResponseWriter, r *http.Request) {
	// All data is sent as HTML form
	// the `ParseForm` method of the request, parses the form's values
	err := r.ParseForm()

	// check for errors
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//location := r.Form.Get("location") + " " + city
	location := city
	latitude, longitude := getLatitudeLongitude(location)
	
	search := r.Form.Get("search")
	if notEmpty(search) {
		searchStrings := strings.Fields(search)
		search = ""
		for i := 0; i < len(searchStrings)-1; i++ {
			search += string(searchStrings[i]) + "%20"
		}
		search += string(searchStrings[len(searchStrings)-1])
	} else {
		search = ""
	}

	var tempSearch string
	for _, val := range search {
		if val == '\'' || val == 8217 {
			tempSearch += "%27"
		} else {
			tempSearch += string(val)
		}
	}
	
	search = tempSearch
	fmt.Println(search)

	getItinerary(latitude, longitude, search, location)

	// redirect the user to create itinerary page (located at `/pages/createItinerary.html`)
	http.Redirect(w, r, "/pages/createItinerary.html", http.StatusFound)
}

func getItineraryHandler(w http.ResponseWriter, r *http.Request) {
	itineraryBytes, err := json.Marshal(Itinerary)

	// check for errors
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(itineraryBytes)
}

func createItineraryHandler(w http.ResponseWriter, r *http.Request) {
	// All data is sent as HTML form
	// the `ParseForm` method of the request, parses the form's values
	err := r.ParseForm()

	// check for errors
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userPick := r.Form.Get("userChoice")
	userNotes := r.Form.Get("userNotes")

	if userNotes != "" {
		userPick += ", Notes: " + userNotes
	}

	Itinerary = append(Itinerary, userPick)
	// append to global image urls
	ItineraryPlanner = nil

	// redirect the user to home page (located at `/pages/`)
	http.Redirect(w, r, "/pages/createItinerary.html", http.StatusFound)
}

func createFinalizedItineraryHandler(w http.ResponseWriter, r *http.Request) {
	// All data is sent as HTML form
	// the `ParseForm` method of the request, parses the form's values
	err := r.ParseForm()

	// check for errors
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	finalItinerary := r.Form["userItinerary"]
	tags := r.Form.Get("tags")

	if notEmpty(tags) {
		if tags[0] != ' ' {
			tags = " " + tags
		}

		if tags[len(tags)-1] != ' ' {
			tags = tags + " "
		}
	} else {
		tags = ""
	}

	var itinerary string

	for i:=0; i < len(finalItinerary); i++ {
		name,stop := substring(finalItinerary[i], 0, ',')

		var price string
		if finalItinerary[i][stop+2] == 'P' {
			price,stop = substring(finalItinerary[i], stop+9, ',')
		} else {
			price = "no price info"
		}

		rating,stop := substring(finalItinerary[i], stop+10, ',')
		address,stop := substring(finalItinerary[i], stop+11, ':')

		notes := ""
		if stop < len(finalItinerary[i]) {
			address = address[:len(address)-7]
			notes = finalItinerary[i][stop+2:]
		}

		itinerary += name + " <br> " + "Price: " + price + " <br> " + "Rating: " + rating + " <br> " + "Address: " + address + " <br> "
		if notes != "" {
			itinerary += "Notes: " + notes + " <br>"
		}
		if i < len(finalItinerary)-1 {
			itinerary += " <br> "
		}
	}

	fmt.Println(itinerary)
	fmt.Println(tags)

	// send finalItinerary info to database
	sqlStatement := `INSERT INTO itinerary_table(itinerary, city, tags) VALUES($1, $2, $3);`
	_, err = db.Exec(sqlStatement, itinerary, city, tags)
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, ok := citiesHashMap[city]; !ok {
		CityAutoComplete = append(CityAutoComplete, city)
		citiesHashMap[city] = city
	}

	// clear itineraries
	ItineraryPlanner = nil
	Itinerary = nil
	city = ""

	// redirect the user to home page (located at `/pages/`)
	http.Redirect(w, r, "/pages/Index.html", http.StatusFound)
}

func deleteItineraryHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	city = ""
	ItineraryPlanner = nil
	Itinerary = nil

	http.Redirect(w, r, "/pages/createItinerary.html", http.StatusFound)
}

func createLocationNullHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	city = ""

	http.Redirect(w, r, "/pages/createItinerary.html", http.StatusFound)
}

func nullItineraryMatchesHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ItineraryPlanner = nil

	http.Redirect(w, r, "/pages/createItinerary.html", http.StatusFound)
}

func getItinerary(latitude string, longitude string, search string, address string) {
	ItineraryPlanner = nil
	
	if latitude == "" || longitude == "" || search == "" {
		noBusiness := Business{}
		noBusiness.Name = "nil"
		ItineraryPlanner = append(ItineraryPlanner, noBusiness)
	} else {
		yelp := getYelpSearch(latitude, longitude, search, 0)

		for _, Business := range yelp.Businesses {
			if Business.IsClosed {
				Business.Name += " (CLOSED)"
			}
			ItineraryPlanner = append(ItineraryPlanner, Business)
		}

		if ItineraryPlanner == nil {
			noBusiness := Business{}
			noBusiness.Name = "nil"
			ItineraryPlanner = append(ItineraryPlanner, noBusiness)
		}
	}
}

func getYelpSearch(latitude string, longitude string, search string, offset int) Yelp {
	getYelpURL := yelpURL + "businesses/search?term=" + search + "&latitude=" + latitude + "&longitude=" + longitude + "&limit=50&offset=" + strconv.Itoa(offset)

	req, err := http.NewRequest("GET", getYelpURL, nil)
	req.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	respRaw, _ := ioutil.ReadAll(resp.Body)
	yelp := Yelp{}
	json.Unmarshal(respRaw, &yelp)
	
	return yelp
}

func getLatitudeLongitude(Address string) (string, string) {
	arguments := strings.Fields(Address)
	mapURL := bingURL + string(arguments[0])
	for i := 1; i < len(arguments); i++ {
		mapURL += "%20"+string(arguments[i])
	}
	mapURL += "?o=json&key="+keys.BingKey

	req, err := http.NewRequest("GET", mapURL, nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	respRaw, _ := ioutil.ReadAll(resp.Body)
	bing := Bing{}
	json.Unmarshal(respRaw, &bing)

	if len(bing.ResourceSets) <= 0 || len(bing.ResourceSets[0].Resources) <= 0 {
		return "",""
	}
	
	return strconv.FormatFloat(bing.ResourceSets[0].Resources[0].Point.Coordinates[0], 'f', -1, 64), strconv.FormatFloat(bing.ResourceSets[0].Resources[0].Point.Coordinates[1], 'f', -1, 64)
}

func substring(str string, start int, stopChar byte) (string,int) {
	i := start
	for i < len(str) && str[i] != stopChar {
		i++
	}
	return str[start:i],i
}

func notEmpty(s string) bool {
	for _,val := range s {
		if !unicode.IsSpace(val) {
			return true
		}
	}
	return false
}