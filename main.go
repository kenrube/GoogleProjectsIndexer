package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/gocolly/colly/queue"
)

type ApiClass struct {
	Index               int    `json:"index"`
	Name                string `json:"name"`
	NameExtended        string `json:"name_extended"`
	Link                string `json:"link"`
	Description         string `json:"description"`
	AddedInVersion      int    `json:"added_in_version"`
	DeprecatedInVersion int    `json:"deprecated_in_version"`
}

func main() {
	c := colly.NewCollector(
		colly.URLFilters(regexp.MustCompile("https://developer\\.android\\.com/reference/.+")),
		colly.MaxDepth(1),
	)

	extensions.RandomUserAgent(c)

	q, err := queue.New(8, nil)
	check(err)

	var index int
	var apiClasses []ApiClass

	c.OnHTML("tr[data-version-added]", func(e *colly.HTMLElement) {
		index++
		name := e.DOM.Find("td[class=jd-linkcol]>a[href]").First().Text()
		nameExtended := e.ChildText("td[class=jd-linkcol]")
		link := e.ChildAttr("td[class=jd-linkcol]>a[href]", "href")
		description := e.ChildText("td[class=jd-descrcol]")
		description = regexp.MustCompile("\\s{2,}").ReplaceAllString(description, " ")
		addedInVersion, _ := strconv.Atoi(e.Attr("data-version-added"))
		deprecatedInVersion, _ := strconv.Atoi(e.Attr("data-version-deprecated"))

		apiClass := ApiClass{
			Index:               index,
			Name:                name,
			NameExtended:        nameExtended,
			Link:                link,
			Description:         description,
			AddedInVersion:      addedInVersion,
			DeprecatedInVersion: deprecatedInVersion,
		}
		apiClasses = append(apiClasses, apiClass)

		/*err := q.AddURL(link)
		check(err)*/
	})

	// a[class=__asdk_search_extension_link__]

	err = q.AddURL("https://developer.android.com/reference/classes")
	check(err)

	err = q.Run(c)
	check(err)

	jsonData, err := json.MarshalIndent(apiClasses, "", "    ")
	check(err)

	err = ioutil.WriteFile("classes_index.json", jsonData, os.ModePerm)
	check(err)
	log.Println(string(jsonData))
}

func check(err error) {
	if err != nil {
		log.Fatal()
	}
}
