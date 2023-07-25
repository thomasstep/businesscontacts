package main

import (
	"context"
	"encoding/csv"
	"log"
	"os"

	"googlemaps.github.io/maps"
)

func main() {
	mapsApiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if mapsApiKey == "" {
		log.Fatal("make sure to set GOOGLE_MAPS_API_KEY (export GOOGLE_MAPS_API_KEY=<your key>)")
	}

	c, err := maps.NewClient(maps.WithAPIKey(mapsApiKey))
	if err != nil {
		log.Fatalf("error creating google maps client: %s", err)
	}

	r := &maps.NearbySearchRequest{
		Location: &maps.LatLng{
			Lat: float64(1),
			Lng: float64(1),
		},
		Radius:   uint(8000),
	}
	places, err := c.NearbySearch(context.Background(), r)
	if err != nil {
		log.Fatalf("error searching nearby places: %s", err)
	}

	csvRes := [][]string{
		{"name", "address", "phone", "url"},
	}

	for _, place := range places.Results {
		placeDetails, err := c.PlaceDetails(context.Background(), &maps.PlaceDetailsRequest{
			PlaceID: place.PlaceID,
			Fields: []maps.PlaceDetailsFieldMask{
				maps.PlaceDetailsFieldMaskName,
				maps.PlaceDetailsFieldMaskFormattedAddress,
				maps.PlaceDetailsFieldMaskFormattedPhoneNumber,
				maps.PlaceDetailsFieldMaskURL,
			},
		})
		if err != nil {
			log.Printf("error reading place details: %s", err)
		}


		csvPlace := []string{
			placeDetails.Name,
			placeDetails.FormattedAddress,
			placeDetails.FormattedPhoneNumber,
			placeDetails.URL,
		}
		csvRes = append(csvRes, csvPlace)
	}

	// TODO Read or create, then append results if already existed
	f, err := os.Create("results.csv")
	if err != nil {
		log.Fatalf("error creating file: %s", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)

	w.WriteAll(csvRes)
	err = w.Error()
	if err != nil {
		log.Fatalf("error writing csv: %s", err)
	}
}
