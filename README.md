# scraper
This mini app consumes an intial webpage and some linked-to webpages,
processes data obtained and writes some of that data to standard output in a prescribed format.

Please note this app requires at least Go version 1.7.


###To run app locally:

```
# Download dependency
go get github.com/PuerkitoBio/goquery

# Download source
go get github.com/billbell73/scraper

# Then navigate to scraper directory and execute:
go run scraper.go
```

###To run tests:
```
go test
```
