package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"math"

	"github.com/spf13/cobra"
	"googlemaps.github.io/maps"
)

var Lat float64
var Lng float64
var Keyword string
var Type string
var Limit int64
var NextPageToken string

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

	if NextPageToken != "" {
		r = &maps.NearbySearchRequest{
			PageToken: NextPageToken,
		}
	}

	csvRes := [][]string{
		{"name", "address", "phone", "url", "nextPageToken"},
	}
	pageToken := ""
	i := int64(0)
	totalIter := int64(math.Ceil(float64(Limit) / float64(20)))

	for i <= totalIter {
		if i > 0 {
			r = &maps.NearbySearchRequest{
				PageToken: pageToken,
			}
		}

		places, err := c.NearbySearch(context.Background(), r)
		if err != nil {
			log.Printf("error searching nearby places: %s", err)
			break
		}

		if places.NextPageToken == "" {
			// This means there are no more results, just write what we have
			break
		}
		pageToken = places.NextPageToken

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
				continue
			}

			csvPlace := []string{
				placeDetails.Name,
				placeDetails.FormattedAddress,
				placeDetails.FormattedPhoneNumber,
				placeDetails.URL,
				pageToken,
			}
			csvRes = append(csvRes, csvPlace)
		}

		// increment
		i = i + 1
	}

	f, err := os.OpenFile("results.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file: %s", err)
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

			if NextPageToken != "" {
				return
			}

			if Lat == -100000 {
				log.Fatal("--lat is a required flag")
			}
			if Lng == -100000 {
				log.Fatal("--lng is a required flag")
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

	rootCmd.Flags().Float64Var(&Lat, "lat", float64(-100000), "Latitude from which the search is based on")
	rootCmd.Flags().Float64Var(&Lng, "lng", float64(-100000), "Longitude from which the search is based on")
	rootCmd.MarkFlagRequired("lat")
	rootCmd.MarkFlagRequired("lng")
	rootCmd.Flags().StringVar(&Keyword, "keyword", "", "Keyword to search on")
	rootCmd.Flags().StringVar(&Type, "type", "", "Type of business to search on; one of https://developers.google.com/maps/documentation/places/web-service/supported_types")
	rootCmd.Flags().Int64Var(&Limit, "limit", int64(100), "Amount of contacts to lookup")
	rootCmd.Flags().StringVar(&NextPageToken, "nextPageToken", "", "Token to lookup where you left off")


	if err := rootCmd.Execute(); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}
