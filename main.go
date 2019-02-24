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

	q, _ := queue.New(8, nil)

	var index int
	var apiClasses []ApiClass

	// todo correctly parse complex cases like ActivityInstrumentationTestCase<T extends Activity>
	c.OnHTML("tr[data-version-added]", func(e *colly.HTMLElement) {
		index++
		name := e.ChildText("td[class=jd-linkcol]>a[href]")
		link := e.ChildAttr("td[class=jd-linkcol]>a[href]", "href")
		description := e.ChildText("td[class=jd-descrcol]")
		description = regexp.MustCompile("\\s{2,}").ReplaceAllString(description, " ")
		addedInVersion, _ := strconv.Atoi(e.Attr("data-version-added"))
		deprecatedInVersion, _ := strconv.Atoi(e.Attr("data-version-deprecated"))

		apiClass := ApiClass{
			Index:               index,
			Name:                name,
			Link:                link,
			Description:         description,
			AddedInVersion:      addedInVersion,
			DeprecatedInVersion: deprecatedInVersion,
		}
		apiClasses = append(apiClasses, apiClass)

		//q.AddURL(link)
	})

	/*c.OnHTML("a[class=__asdk_search_extension_link__]", func(e *colly.HTMLElement) {
		log.Println("Source file link:", e.Attr("href"))
	})*/

	q.AddURL("https://developer.android.com/reference/classes")
	q.Run(c)

	jsonData, _ := json.MarshalIndent(apiClasses, "", "    ")
	ioutil.WriteFile("classes_index.json", jsonData, os.ModePerm)
	log.Println(string(jsonData))
}
