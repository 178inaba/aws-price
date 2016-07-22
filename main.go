package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

const (
	baseURL        = "https://pricing.us-east-1.amazonaws.com"
	offerIndexPath = "/offers/v1.0/aws/index.json"
	ec2            = "AmazonEC2"
)

type offerIndex struct {
	FormatVersion   string               `json:"formatVersion"`
	Disclaimer      string               `json:"disclaimer"`
	PublicationDate string               `json:"publicationDate"`
	Offers          map[string]offerPath `json:"offers"`
}

type offerPath struct {
	OfferCode         string `json:"offerCode"`
	VersionIndexURL   string `json:"versionIndexURL"`
	CurrentVersionURL string `json:"currentVersionURL"`
}

type offer struct {
	FormatVersion   string                                `json:"formatVersion"`
	Disclaimer      string                                `json:"disclaimer"`
	OfferCode       string                                `json:"offerCode"`
	Version         string                                `json:"version"`
	PublicationDate string                                `json:"publicationDate"`
	Products        map[string]product                    `json:"products"`
	Terms           map[string]map[string]map[string]term `json:"terms"`
}

type product struct {
	Sku           string            `json:"sku"`
	ProductFamily string            `json:"productFamily"`
	Attributes    map[string]string `json:"attributes"`
}

type term struct {
	OfferTermCode      string                    `json:"offerTermCode"`
	Sku                string                    `json:"sku"`
	EffectiveDate      string                    `json:"effectiveDate"`
	TermAttributesType string                    `json:"termAttributesType"`
	TermAttributes     map[string]string         `json:"termAttributes"`
	PriceDimensions    map[string]priceDimension `json:"priceDimensions"`
}

type priceDimension struct {
	RateCode      string            `json:"rateCode"`
	Description   string            `json:"description"`
	Unit          string            `json:"unit"`
	StartingRange string            `json:"startingRange"`
	EndingRange   string            `json:"endingRange"`
	PricePerUnit  map[string]string `json:"pricePerUnit"`
}

func main() {
	var err error

	offerIndexJSON, err := getJSON(baseURL + offerIndexPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
	}

	var offerIndex offerIndex
	err = json.Unmarshal(offerIndexJSON, &offerIndex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
	}

	offerPath := offerIndex.Offers[ec2].CurrentVersionURL
	offerJSON, err := getJSON(baseURL + offerPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
	}

	var offer offer
	err = json.Unmarshal(offerJSON, &offer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
	}

	for _, product := range offer.Products {
		var isOregon, isLinux bool
		for name, value := range product.Attributes {
			if name == "location" && value == "US West (Oregon)" {
				isOregon = true
			} else if name == "operatingSystem" && value == "Linux" {
				isLinux = true
			}
		}

		if isOregon && isLinux {
			if product.Attributes["instanceType"] == "m2.2xlarge" {
				b, _ := json.Marshal(product)
				var out bytes.Buffer
				json.Indent(&out, b, "", "\t")
				out.WriteTo(os.Stdout)
			}
			//fmt.Printf("%+v\n", product.Attributes["instanceType"])
		}
	}
	/*
		for _, skus := range offer.Terms {
			for _, terms := range skus {
				for _, term := range terms {
					fmt.Println(term)
				}
			}
		}
	*/
}

func getJSON(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "json get error")
	}
	defer resp.Body.Close()

	json, err := ioutil.ReadAll(resp.Body)
	return json, nil
}
