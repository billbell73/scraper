// scraper.go consumes an intial webpage and some linked-to webpages,
// processes data obtained and and writes some of that data in a prescribed
// format to standard output.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const webAddress string = "http://hiring-tests.s3-website-eu-west-1.amazonaws.com/2015_Developer_Scrape/5_products.html"

type price float32
type pageSize int

type product struct {
	Title       string   `json:"title"`
	PageSize    pageSize `json:"size"`
	UnitPrice   price    `json:"unit_price"`
	Description string   `json:"description"`
}

type scraperDisplay struct {
	Results []product `json:"results"`
	Total   price     `json:"total"`
}

func main() {
	var products []product
	var numberProducts int

	doc := fetchDoc(webAddress)
	c := make(chan product)

	doc.Find(".product ").Each(func(i int, s *goquery.Selection) {
		numberProducts++
		go scrapeInfo(s, c, readProductPage, fetchDoc)
	})

	for j := 0; j < numberProducts; j++ {
		products = append(products, <-c)
	}

	toJSON(products, os.Stdout)
}

func fetchDoc(dest string) *goquery.Document {
	doc, err := goquery.NewDocument(dest)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

// readProductPage uses passed docFetcher function to get webpage at passed url
// and returns the size in bytes of the fetched html and the string 'content'
// value of a meta tag with 'name' attribute of "description".
func readProductPage(dest string, fn docFetcher) (pageSize, string) {
	linkedDoc := fn(dest)

	var description string

	linkedDoc.Find("meta").Each(func(i int, s *goquery.Selection) {
		nameAttr, _ := s.Attr("name")
		if nameAttr == "description" {
			description, _ = s.Attr("content")
		}
	})

	size := sizeOf(linkedDoc)

	return pageSize(size), description
}

func sizeOf(doc *goquery.Document) int {
	html, err := doc.Html()
	if err != nil {
		log.Fatal(err)
	}

	return len(html)
}

type docFetcher func(string) *goquery.Document
type productPageReader func(string, docFetcher) (pageSize, string)

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
		Title:       strings.TrimSpace(rawTitle),
		PageSize:    size,
		UnitPrice:   floatifyPrice(rawPrice),
		Description: description}

	c <- p
}

func floatifyPrice(s string) price {
	re := regexp.MustCompile("[0-9]+.[0-9]+")
	stringPrice := re.FindString(s)

	if stringPrice == "" {
		var empty price
		return empty
	}

	f64, err := strconv.ParseFloat(stringPrice, 32)
	if err != nil {
		log.Fatal(err)
	}

	return price(f64)
}

// toJSON takes a slice of products and an io.Writer, and gets the writer to
// output JSON-formatted info in the prescribed form about the products.
func toJSON(products []product, w io.Writer) {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")

	total := totalPrice(products)

	err := enc.Encode(scraperDisplay{
		Results: products,
		Total:   total,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func totalPrice(products []product) price {
	var total price

	for _, product := range products {
		total += product.UnitPrice
	}
	return total
}

func (p price) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%.2f", p)), nil
}

func (p pageSize) MarshalText() ([]byte, error) {
	sizeInKB := float64(p) / 1000
	rounded := roundToOneDecPlace(sizeInKB)
	return []byte(fmt.Sprintf("%vkb", rounded)), nil
}

func roundToOneDecPlace(f float64) float64 {
	timesTen := f * 10
	roundedtimesTen := math.Floor(timesTen + 0.5)
	return roundedtimesTen / 10
}
