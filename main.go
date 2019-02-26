package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

type Library struct {
	Name string
	Link string
}

type LibraryClasses struct {
	Name    string     `json:"library_name"`
	Classes []ApiClass `json:"classes"`
}

type ApiClass struct {
	Id                  int    `json:"id"`
	Name                string `json:"name"`
	NameExtended        string `json:"name_extended"`
	Link                string `json:"link"`
	Description         string `json:"description"`
	AddedInVersion      string `json:"added_in_version"`
	DeprecatedInVersion string `json:"deprecated_in_version"`
}

func main() {
	c := colly.NewCollector(
		colly.URLFilters(regexp.MustCompile("https://developer\\.android\\.com/reference/.+")),
		colly.MaxDepth(1),
	)

	extensions.RandomUserAgent(c)

	var id int
	var libraryClasses []LibraryClasses
	libraries := [...]Library{
		{Name: "Android Platform", Link: "/classes"},
		{Name: "Android Support Library", Link: "/android/support/classes"},
		{Name: "AndroidX", Link: "/androidx/classes"},
		{Name: "AndroidX Test", Link: "/androidx/test/classes"},
		{Name: "Architecture Components", Link: "/android/arch/classes"},
		{Name: "Android Automotive Library", Link: "/android/car/classes"},
		{Name: "Databinding Library", Link: "/android/databinding/classes"},
		{Name: "Constraint Layout Library", Link: "/android/support/constraint/classes"},
		{Name: "Material Components", Link: "/com/google/android/material/classes"},
		{Name: "Test Support Library", Link: "/android/support/test/classes"},
		{Name: "Wearable Library", Link: "/android/support/wearable/classes"},
		{Name: "Play Billing Library", Link: "/com/android/billingclient/classes"},
		{Name: "Play Core Library", Link: "/com/google/android/play/core/classes"},
		{Name: "Play Install Referrer Library", Link: "/com/android/installreferrer/classes"},
		{Name: "Android Things", Link: "/com/google/android/things/classes"},
	}

	c.OnHTML("tr", func(e *colly.HTMLElement) {
		id++
		libraryClasses[len(libraryClasses)-1].Classes =
			append(libraryClasses[len(libraryClasses)-1].Classes, parseApiClass(id, e))
	})

	for index := range libraries {
		libraryClasses = append(libraryClasses, LibraryClasses{libraries[index].Name, []ApiClass{}})
		// TODO visit links concurrently
		err := c.Visit("https://developer.android.com/reference" + libraries[index].Link)
		check(err)
	}

	jsonData, err := json.MarshalIndent(libraryClasses, "", "    ")
	check(err)

	err = ioutil.WriteFile("classes_index.json", jsonData, os.ModePerm)
	check(err)
	log.Println(string(jsonData))
}

func parseApiClass(id int, e *colly.HTMLElement) ApiClass {
	name := e.DOM.Find("td[class=jd-linkcol]>a[href]").First().Text()
	nameExtended := e.ChildText("td[class=jd-linkcol]")
	link := e.ChildAttr("td[class=jd-linkcol]>a[href]", "href")
	description := e.ChildText("td[class=jd-descrcol]")
	description = regexp.MustCompile("\\s{2,}").ReplaceAllString(description, " ")
	addedInVersion := e.Attr("data-version-added")
	deprecatedInVersion := e.Attr("data-version-deprecated")

	return ApiClass{
		Id:                  id,
		Name:                name,
		NameExtended:        nameExtended,
		Link:                link,
		Description:         description,
		AddedInVersion:      addedInVersion,
		DeprecatedInVersion: deprecatedInVersion,
	}
}

func check(err error) {
	if err != nil {
		log.Println(err)
	}
}
