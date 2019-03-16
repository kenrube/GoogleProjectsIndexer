package main

import (
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

type AllClasses struct {
	ClassCount int              `json:"class_count"`
	Classes    []LibraryClasses `json:"classes"`
}

type LibraryClasses struct {
	Name       string     `json:"library_name"`
	ClassCount int        `json:"class_count"`
	Classes    []ApiClass `json:"classes"`
}

type ApiClass struct {
	Name                string `json:"name"`
	NameExtended        string `json:"name_extended,omitempty"`
	Link                string `json:"link"`
	Description         string `json:"description,omitempty"`
	AddedInVersion      string `json:"added_in_version,omitempty"`
	DeprecatedInVersion string `json:"deprecated_in_version,omitempty"`
}

const baseDocLink string = "https://developer.android.com/reference"

func main() {
	c := colly.NewCollector()
	extensions.RandomUserAgent(c)

	var allClasses AllClasses
	var libraryClasses []LibraryClasses

	c.OnHTML("tr", func(e *colly.HTMLElement) {
		libraryClasses[len(libraryClasses)-1].ClassCount += 1
		libraryClasses[len(libraryClasses)-1].Classes =
			append(libraryClasses[len(libraryClasses)-1].Classes, parseApiClass(e))
	})

	librariesFile, err := os.Open("libraries.csv")
	check(err, "Can't open libraries.csv")
	r := csv.NewReader(librariesFile)
	libraries, err := r.ReadAll()
	check(err, "Can't read libraries.csv")

	for index := range libraries {
		libraryClasses = append(libraryClasses, LibraryClasses{libraries[index][0], 0, []ApiClass{}})
		link := baseDocLink + libraries[index][1]
		err := c.Visit(link)
		check(err, "Can't visit link", link)
	}
	for index := range libraryClasses {
		allClasses.ClassCount += libraryClasses[index].ClassCount
	}
	allClasses.Classes = libraryClasses

	jsonData, err := json.MarshalIndent(allClasses, "", "  ")
	check(err, "Can't marshal and indent allClasses")

	err = ioutil.WriteFile("classes_index.json", jsonData, os.ModePerm)
	check(err, "Can't write json to file")
	log.Println("Found", allClasses.ClassCount, "classes in", len(libraries), "libraries")
}

func parseApiClass(e *colly.HTMLElement) ApiClass {
	name := e.DOM.Find("td[class=jd-linkcol]>a[href]").First().Text()
	nameExtended := e.ChildText("td[class=jd-linkcol]")
	if nameExtended == name {
		nameExtended = "" // to omit field in json
	}
	link := e.ChildAttr("td[class=jd-linkcol]>a[href]", "href")
	description := e.ChildText("td[class=jd-descrcol]")
	description = regexp.MustCompile("\\s{2,}").ReplaceAllString(description, " ")
	addedInVersion := e.Attr("data-version-added")
	deprecatedInVersion := e.Attr("data-version-deprecated")

	return ApiClass{
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
