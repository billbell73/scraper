// scraper.go consumes an intial webpage and some linked-to webpages,
// processes some data obtained and presents that data in a specified
// format.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const webAddress string = "http://hiring-tests.s3-website-eu-west-1.amazonaws.com/2015_Developer_Scrape/5_products.html"

type product struct {
	title       string
	unitPrice   float32
	pageSize    int
	description string
}

type productDisplay struct {
	Title       string      `json:"title"`
	PageSize    string      `json:"size"`
	UnitPrice   json.Number `json:"unit_price"`
	Description string      `json:"description"`
}

type scraperDisplay struct {
	Results *[]productDisplay `json:"results"`
	Total   json.Number      `json:"total"`
}

func main() {
	var products []product
	var numberProducts int

	doc := fetchDoc(webAddress)
	c := make(chan product)

	doc.Find(".product ").Each(func(i int, s *goquery.Selection) {
		numberProducts = i
		go scrapeInfo(s, c, readProductPage, fetchDoc)
	})

	for j := 0; j <= numberProducts; j++ {
		products = append(products, <-c)
	}

	fmt.Printf("%v", toJSON(products))
}

func fetchDoc(dest string) *goquery.Document {
	doc, err := goquery.NewDocument(dest)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

func floatifyPrice(s string) float32 {
	re := regexp.MustCompile("[0-9]+.[0-9]+")
	price := re.FindString(s)

	if price == "" {
		var empty float32
		return empty
	}

	f64, err := strconv.ParseFloat(price, 32)
	if err != nil {
		log.Fatal(err)
	}

	return float32(f64)
}

// readProductPage uses passed docFetcher function to get webpage at passed url
// and returns the size of the fetched html in bytes and the string 'content'
// value of a meta tag with 'name' attribute of "description".
func readProductPage(dest string, fn docFetcher) (int, string) {
	linkedDoc := fn(dest)

	var description string

	linkedDoc.Find("meta").Each(func(i int, s *goquery.Selection) {
		nameAttr, _ := s.Attr("name")
		if nameAttr == "description" {
			description, _ = s.Attr("content")
		}
	})

	return sizeOf(linkedDoc), description
}

func sizeOf(doc *goquery.Document) int {
	html, err := doc.Html()
	if err != nil {
		log.Fatal(err)
	}

	return len(html)
}

type docFetcher func(string) *goquery.Document
type productPageReader func(string, docFetcher) (int, string)

// scrapeInfo extracts info on a product from a passed goquery selection,
// calls a productPageReader method to get further info and then populates
// a product struct, which it sends into the channel it is passed.
func scrapeInfo(s *goquery.Selection, c chan product, ppr productPageReader, df docFetcher) {
	productLink := s.Find("a")
	rawTitle := productLink.Text()
	destination, _ := productLink.Attr("href")

	rawPrice := s.Find(".pricePerUnit").Text()

	size, description := ppr(destination, df)

	p := product{
		title:       strings.TrimSpace(rawTitle),
		unitPrice:   floatifyPrice(rawPrice),
		pageSize:    size,
		description: description}

	c <- p
}

// iterateProducts takes a slice of products and outputs cumulative unit price
// and a pointer to a slice of productDisplay structs.
func iterateProducts(products []product) (float32, *[]productDisplay) {
	var totalPrice float32
	var d []productDisplay

	for _, product := range products {
		totalPrice += product.unitPrice

		pd := productDisplay{
			Title:       product.title,
			PageSize:    displaySize(product.pageSize),
			UnitPrice:   displayPrice(product.unitPrice),
			Description: product.description,
		}
		d = append(d, pd)
	}
	return totalPrice, &d
}

func displaySize(size int) string {
	sizeInKB := float64(size) / 1000
	rounded := roundToOneDecPlace(sizeInKB)
	return fmt.Sprintf("%v", rounded) + "kb"
}

func roundToOneDecPlace(f float64) float64 {
	timesTen := f * 10
	roundedtimesTen := math.Floor(timesTen + 0.5)
	return roundedtimesTen / 10
}

func displayPrice(price float32) json.Number {
	twoDecPlacesString := fmt.Sprintf("%.2f", price)

	dec := json.NewDecoder(strings.NewReader(twoDecPlacesString))
	var x json.Number
	dec.Decode(&x)
	return x
}

// toJSON takes a slice of products and outputs a json string
// containing info about the products in the specified format.
func toJSON(products []product) string {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")

	totalPrice, results := iterateProducts(products)

	err := enc.Encode(scraperDisplay{
		Results: results,
		Total:   displayPrice(totalPrice),
	})
	if err != nil {
		log.Fatal(err)
	}

	return buf.String()
}
