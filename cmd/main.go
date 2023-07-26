package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"googlemaps.github.io/maps"
)

var Lat float64
var Lng float64
var Keyword string
var Type string

func contacts(cmd *cobra.Command, args []string) {
	mapsApiKey := os.Getenv("GOOGLE_MAPS_API_KEY")

	c, err := maps.NewClient(maps.WithAPIKey(mapsApiKey))
	if err != nil {
		log.Fatalf("error creating google maps client: %s", err)
	}

	r := &maps.NearbySearchRequest{
		Location: &maps.LatLng{
			Lat: Lat,
			Lng: Lng,
		},
		Radius:   uint(8000),
	}

	if Type != "" {
		ptype, err := maps.ParsePlaceType(Type)
		if err != nil {
			log.Printf("error parsing place type, but continuing to run: %s", err)
		}

		r.Type = ptype
	}

	if Keyword != "" {
		r.Keyword = Keyword
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

func main() {
	rootCmd := &cobra.Command{
		Use:   "contacts",
		Short: "Contacts searches for business' contact informatoin",
		Long: `Searches for business contact info using Google Places API
						and formats it into a CSV`,
		PreRun: func(cmd *cobra.Command, args []string) {
			mapsApiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
			if mapsApiKey == "" {
				log.Fatal("set GOOGLE_MAPS_API_KEY (export GOOGLE_MAPS_API_KEY=<your key>)")
			}

			if Keyword == "" && Type == "" {
				log.Fatal("at least one of --keyword or --type need to be used")
			}
		},
		Run: contacts,
		PostRun: func(cmd *cobra.Command, args []string) {
      fmt.Print("Results written to results.csv")
    },
	}

	rootCmd.Flags().Float64Var(&Lat, "lat", float64(0), "Latitude from which the search is based on")
	rootCmd.Flags().Float64Var(&Lng, "lng", float64(0), "Longitude from which the search is based on")
	rootCmd.MarkFlagRequired("lat")
	rootCmd.MarkFlagRequired("lng")
	rootCmd.Flags().StringVar(&Keyword, "keyword", "", "Keyword to search on")
	rootCmd.Flags().StringVar(&Type, "type", "", "Type of business to search on; one of https://developers.google.com/maps/documentation/places/web-service/supported_types")

	if err := rootCmd.Execute(); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}
