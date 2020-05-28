package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"github.tamu.edu/rmyoung/mashup/keys"
	_"github.com/lib/pq"
	"strconv"
	"strings"
	"unicode"
)

func getItineraryLocationHandler(w http.ResponseWriter, r *http.Request) {
	itineraryCityBytes, err := json.Marshal(itineraryCity)

	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}	

	w.Write(itineraryCityBytes)
}

func createItineraryLocationHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	itineraryCity = r.Form.Get("itinerarydestination")
	if !isAlphabetic(itineraryCity) {
		itineraryCity = "incorrect format"
	} else {
		if _, ok := citiesHashMap[itineraryCity]; !ok {
			itineraryCity = "no itineraries"
		} else {
			TagsAutoComplete = nil
			tagsHashMap := make(map[string] string)

			sqlStatement := `SELECT tags FROM itinerary_table WHERE city = '` + itineraryCity + `';`
			rows, err := db.Query(sqlStatement)
			defer rows.Close()

			for rows.Next() {
				var autocompleteTags string
				err = rows.Scan(&autocompleteTags)
				if err != nil {
					fmt.Println(fmt.Errorf("Error: %v", err))
					panic(err)
				}

				tagArguments := strings.Fields(autocompleteTags)

				for i:=0; i < len(tagArguments); i++ {
					newTag := tagArguments[i]
					if _, ok := tagsHashMap[newTag]; !ok {
						TagsAutoComplete = append(TagsAutoComplete, newTag)
						tagsHashMap[newTag] = newTag
					}
				}
			}
		}
	}

	http.Redirect(w, r, "/pages/getItinerary.html", http.StatusFound)
}

func getCitiesHandler(w http.ResponseWriter, r *http.Request) {
	cityAutoCompleteBytes, err := json.Marshal(CityAutoComplete)

	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(cityAutoCompleteBytes)
}

func getTagsHandler(w http.ResponseWriter, r *http.Request) {
	tagsAutoCompleteBytes, err := json.Marshal(TagsAutoComplete)

	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(tagsAutoCompleteBytes)
}

func getUserItineraryHandler(w http.ResponseWriter, r *http.Request) {
	userItineraryBytes, err := json.Marshal(UserItinerary)

	// check for errors
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(userItineraryBytes)
}

func createUserItineraryHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	UserItinerary = nil

	destination := itineraryCity
	date := r.Form.Get("date")
	tags := r.Form.Get("tags")

	// just printing tags for now
	fmt.Println(tags)

	latitude, longitude := getLatitudeLongitude(destination)

	var forecast string
	if date != "" {
		forecast = getForecast(latitude, longitude, date)
	}

	if forecast == "" {
		forecast = "No Date Entered for Weather Forecast"
	}

	UserItinerary = append(UserItinerary, forecast)

	var sqlStatement string

	if notEmpty(tags) {
		sqlStatement = `SELECT itinerary FROM itinerary_table WHERE city = '` + destination + `' AND `
		tagArguments := strings.Fields(tags)
		for i:=0; i < len(tagArguments); i++ {
			sqlStatement += `tags ILIKE '% ` + tagArguments[i] + ` %'`
			if i < len(tagArguments)-1 {
				sqlStatement += ` AND `
			} else {
				sqlStatement += `;`
			}
		}
	} else {
		sqlStatement = `SELECT itinerary FROM itinerary_table WHERE city = '` + destination + `'`
	}

	fmt.Println(sqlStatement)
	
	rows, err := db.Query(sqlStatement)
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	numRows := 0
	for rows.Next() {
		var itinerary string
		err = rows.Scan(&itinerary)
		if err != nil {
			fmt.Println(fmt.Errorf("Error: %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if itinerary != "" && itinerary != " " {
			UserItinerary = append(UserItinerary, itinerary)
		}
		numRows++
	}
	if numRows == 0 {
		response := `No Itineraries matching "` + tags + `"`
		UserItinerary = append(UserItinerary, response)
	}

	http.Redirect(w, r, "/pages/getItinerary.html", http.StatusFound)
}

func createNewItineraryHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	itineraryCity = ""
	UserItinerary = nil

	http.Redirect(w, r, "/pages/getItinerary.html", http.StatusFound)
}

func createChangeTagsItinerary(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	UserItinerary = nil

	http.Redirect(w, r, "/pages/getItinerary.html", http.StatusFound)
}

func getLocationNullHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	itineraryCity = ""

	http.Redirect(w, r, "/pages/getItinerary.html", http.StatusFound)
}

func getForecast(latitude string, longitude string, time string) string {
	getDarkSkyURL := darkSkyURL + keys.DarkskyKey + "/" + latitude + "," + longitude

	if time != "" {
		getDarkSkyURL += "," + time + "T00:00:00"
	}

	resp, err := http.Get(getDarkSkyURL)

	if err != nil {
		fmt.Println(err)
	}

	respRaw, _ := ioutil.ReadAll(resp.Body)
	darkSkyForecast := DarkSkyForecast{}
	json.Unmarshal(respRaw, &darkSkyForecast)

	if len(darkSkyForecast.Daily.Data) == 0 {
		return "To receive weather forecast, enter a date"
	}

	months := [12]string{"January","February","March","April","May","June","July","August","September","October", "November", "December"}
	i, err := strconv.Atoi(time[5:7])

	forecast := "<h5><b> " + months[i-1] + " " + time[8:10] + " </b></h5> " + darkSkyForecast.Daily.Data[0].Summary + " <br> "
	forecast += "Temperature High: " + strconv.FormatFloat(darkSkyForecast.Daily.Data[0].TemperatureHigh, 'f', 0, 64) + " °F <br> "
	forecast += "Temperature Low: " + strconv.FormatFloat(darkSkyForecast.Daily.Data[0].TemperatureLow, 'f', 0, 64) + " °F"

	return forecast
}

func fill() {
	sqlStatement := `SELECT DISTINCT city FROM itinerary_table;`
	rows, err := db.Query(sqlStatement)
	defer rows.Close()

	for rows.Next() {
		var autocompleteCity string
		err = rows.Scan(&autocompleteCity)
		if err != nil {
			fmt.Println(fmt.Errorf("Error: %v", err))
			panic(err)
		}

		CityAutoComplete = append(CityAutoComplete, autocompleteCity)
		citiesHashMap[autocompleteCity] = autocompleteCity
	}
}

func isAlphabetic(s string) bool {
	if s == "" {
		return false
	}

	singleLetter := false
	
	for _, r := range s {
		if !unicode.IsLetter(r) {
			if r != ',' && r != ' ' && r != '-' {
				return false
			}
		} else if unicode.IsLetter(r) {
			singleLetter = true
		}
	}

	return singleLetter
}