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
	Name         string
	RelativeLink string
}

var libraries = [...]Library{
	{"Android Platform", "/classes"},
	{"Android Support Library", "/android/support/classes"},
	{"AndroidX", "/androidx/classes"},
	{"AndroidX Test", "/androidx/test/classes"},
	{"Architecture Components", "/android/arch/classes"},
	{"Android Automotive Library", "/android/car/classes"},
	{"Databinding Library", "/android/databinding/classes"},
	{"Constraint Layout Library", "/android/support/constraint/classes"},
	{"Material Components", "/com/google/android/material/classes"},
	{"Test Support Library", "/android/support/test/classes"},
	{"Wearable Library", "/android/support/wearable/classes"},
	{"Play Billing Library", "/com/android/billingclient/classes"},
	{"Play Core Library", "/com/google/android/play/core/classes"},
	{"Play Install Referrer Library", "/com/android/installreferrer/classes"},
	{"Android Things", "/com/google/android/things/classes"},
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
	AddedInVersion      string `json:"added_in_version,omitempty"`
	DeprecatedInVersion string `json:"deprecated_in_version,omitempty"`
}

func main() {
	c := colly.NewCollector()
	extensions.RandomUserAgent(c)

	var id int
	var libraryClasses []LibraryClasses

	c.OnHTML("tr", func(e *colly.HTMLElement) {
		id++
		libraryClasses[len(libraryClasses)-1].Classes =
			append(libraryClasses[len(libraryClasses)-1].Classes, parseApiClass(id, e))
	})

	for index := range libraries {
		libraryClasses = append(libraryClasses, LibraryClasses{libraries[index].Name, []ApiClass{}})
		// TODO visit links concurrently
		link := "https://developer.android.com/reference" + libraries[index].RelativeLink
		err := c.Visit(link)
		check(err, "Can't visit link ", link)
	}

	jsonData, err := json.MarshalIndent(libraryClasses, "", "    ")
	check(err, "Can't marshal and indent libraryClasses")

	err = ioutil.WriteFile("classes_index.json", jsonData, os.ModePerm)
	check(err, "Can't write json to file")
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

func check(err error, message ...interface{}) {
	if err != nil {
		log.Println(message)
	}
}
