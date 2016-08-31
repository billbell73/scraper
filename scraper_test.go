package main

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestFloatifyPrice(t *testing.T) {
	var floatifyPriceTests = []struct {
		s        string
		expected float32
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

var fakeHtmlString = "<html><head><title>Apricot Ripe</title>" +
	"<meta name=\"description\" content=\"Buy Sainsbury&#39;s!\"/>" +
	"<meta name=\"keyword\" content=\"blank\"/></head><body></body></html>"

func stubDocFetcher(s string) *goquery.Document {
	fakeHtmlReader := strings.NewReader(fakeHtmlString)
	doc, _ := goquery.NewDocumentFromReader(fakeHtmlReader)
	return doc
}

func TestReadProductPage(t *testing.T) {
	actualSize, actualDesc := readProductPage("", stubDocFetcher)
	expectedSize := len(fakeHtmlString)
	if actualSize != expectedSize {
		t.Errorf("whoops: expected %d, actual %d", expectedSize, actualSize)
	}
	if actualDesc != "Buy Sainsbury's!" {
		t.Errorf("whoops: expected \"Buy Sainsbury's!\", actual %q", actualDesc)
	}
}

func stubProductPageReader(s string, fn docFetcher) (int, string) {
	return 42, "life, etc."
}

func TestScrapeInfo(t *testing.T) {
	fakeHtmlString2 := "<a href=\"example.com\">hi</a>" +
		"<p class=\"pricePerUnit\">£3.50</p>"
	fakeHtmlReader2 := strings.NewReader(fakeHtmlString2)
	doc2, _ := goquery.NewDocumentFromReader(fakeHtmlReader2)

	ch := make(chan product)
	go scrapeInfo(doc2.Selection, ch, stubProductPageReader, stubDocFetcher)
	actualProduct := <-ch

	expectedProduct := product{
		title:       "hi",
		unitPrice:   3.5,
		pageSize:    42,
		description: "life, etc.",
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

func TestDisplaySize(t *testing.T) {
	var displaySizeTests = []struct {
		size     int
		expected string
	}{
		{40000, "40kb"},
		{61550, "61.6kb"},
		{541040, "541kb"},
	}

	for _, dst := range displaySizeTests {
		actual := displaySize(dst.size)
		if actual != dst.expected {
			t.Errorf("roundingTests(%q): expected %g, actual %g", dst.size, dst.expected, actual)
		}
	}
}

func TestToJSON(t *testing.T) {
	product1 := product{
		title:       "hi",
		unitPrice:   3.5,
		pageSize:    42000,
		description: "life, etc.",
	}

	product2 := product{
		title:       "hiya & lowa",
		unitPrice:   3.6,
		pageSize:    4200,
		description: "life, the etc.",
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
	actualJSON := toJSON(products)

	if actualJSON != expectedJSON {
		t.Errorf("whoops: expected %s, actual %s", expectedJSON, actualJSON)
	}
}
