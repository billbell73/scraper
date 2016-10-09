package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestFloatifyPrice(t *testing.T) {
	var floatifyPriceTests = []struct {
		s        string
		expected price
	}{
		{"\n£1.50/unit\n", 1.5},
		{"\n£2.54/unit\n", 2.54},
		{"\n£22.53/unit\n", 22.53},
	}

	for _, fpt := range floatifyPriceTests {
		actual := floatifyPrice(fpt.s)
		if actual != fpt.expected {
			t.Errorf("floatifyPrice(%q): expected %g, actual %g", fpt.s, fpt.expected, actual)
		}
	}
}

var fakeHTMLString = "<html><head><title>Apricot Ripe</title>" +
	"<meta name=\"description\" content=\"Buy Sainsbury&#39;s!\"/>" +
	"<meta name=\"keyword\" content=\"blank\"/></head><body></body></html>"

func stubDocFetcher(s string) *goquery.Document {
	fakeHTMLReader := strings.NewReader(fakeHTMLString)
	doc, _ := goquery.NewDocumentFromReader(fakeHTMLReader)
	return doc
}

func TestReadProductPage(t *testing.T) {
	actualSize, actualDesc := readProductPage("", stubDocFetcher)
	expectedSize := pageSize(len(fakeHTMLString))
	if actualSize != expectedSize {
		t.Errorf("whoops: expected %d, actual %d", expectedSize, actualSize)
	}
	if actualDesc != "Buy Sainsbury's!" {
		t.Errorf("whoops: expected \"Buy Sainsbury's!\", actual %q", actualDesc)
	}
}

func stubProductPageReader(s string, fn docFetcher) (pageSize, string) {
	return 42, "life, etc."
}

func TestScrapeInfo(t *testing.T) {
	fakeHTMLString2 := "<a href=\"example.com\">hi</a>" +
		"<p class=\"pricePerUnit\">£3.50</p>"
	fakeHTMLReader2 := strings.NewReader(fakeHTMLString2)
	doc2, _ := goquery.NewDocumentFromReader(fakeHTMLReader2)

	ch := make(chan product)
	go scrapeInfo(doc2.Selection, ch, stubProductPageReader, stubDocFetcher)
	actualProduct := <-ch

	expectedProduct := product{
		Title:       "hi",
		UnitPrice:   3.5,
		PageSize:    42,
		Description: "life, etc.",
	}

	if actualProduct != expectedProduct {
		t.Errorf("whoops: expected %v, actual %v", expectedProduct, actualProduct)
	}
}

func TestRoundToOneDecPlace(t *testing.T) {
	var roundingTests = []struct {
		size     float64
		expected float64
	}{
		{40.0, 40.0},
		{61.55, 61.6},
		{41.04, 41.0},
	}

	for _, rt := range roundingTests {
		actual := roundToOneDecPlace(rt.size)
		if actual != rt.expected {
			t.Errorf("roundingTests(%q): expected %g, actual %g", rt.size, rt.expected, actual)
		}
	}
}

func TestPageSizeMarshallText(t *testing.T) {
	var sizeMarshallTests = []struct {
		size     pageSize
		expected string
	}{
		{40000, "40kb"},
		{61550, "61.6kb"},
		{541040, "541kb"},
	}

	for _, smt := range sizeMarshallTests {
		actual, _ := smt.size.MarshalText()
		if string(actual) != smt.expected {
			t.Errorf("roundingTests(%q): expected %g, actual %g", smt.size, smt.expected, actual)
		}
	}
}

func TestToJSON(t *testing.T) {
	product1 := product{
		Title:       "hi",
		PageSize:    42000,
		UnitPrice:   3.5,
		Description: "life, etc.",
	}

	product2 := product{
		Title:       "hiya & lowa",
		PageSize:    4200,
		UnitPrice:   3.6,
		Description: "life, the etc.",
	}

	products := []product{product1, product2}

	expectedJSON := `{
    "results": [
        {
            "title": "hi",
            "size": "42kb",
            "unit_price": 3.50,
            "description": "life, etc."
        },
        {
            "title": "hiya & lowa",
            "size": "4.2kb",
            "unit_price": 3.60,
            "description": "life, the etc."
        }
    ],
    "total": 7.10
}
`

	buf := new(bytes.Buffer)
	toJSON(products, buf)

	actualJSON := buf.String()

	if actualJSON != expectedJSON {
		t.Errorf("whoops: expected %s, actual %s", expectedJSON, actualJSON)
	}
}
