package main

import (
	"encoding/json"
	"log"
	"regexp"
	"strconv"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
)

type ApiClass struct {
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

	q, _ := queue.New(8, nil)

	// API class count: 4387
	apiClasses := make([]ApiClass, 5000)

	// todo correctly parse complex cases like ActivityInstrumentationTestCase<T extends Activity>
	c.OnHTML("tr[data-version-added]", func(e *colly.HTMLElement) {
		name := e.ChildText("td[class=jd-linkcol]>a[href]")
		link := e.ChildAttr("td[class=jd-linkcol]>a[href]", "href")
		description := e.ChildText("td[class=jd-descrcol]")
		description = regexp.MustCompile("\\s{2,}").ReplaceAllString(description, " ")
		addedInVersion, _ := strconv.Atoi(e.Attr("data-version-added"))
		deprecatedInVersion, _ := strconv.Atoi(e.Attr("data-version-deprecated"))

		apiClass := ApiClass{
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

	jsonData, _ := json.MarshalIndent(apiClasses, "", " ")
	log.Println(string(jsonData))
}
