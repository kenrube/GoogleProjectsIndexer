package main

import (
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

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
	SourceLink          string `json:"source_link,omitempty"`
	Description         string `json:"description,omitempty"`
	AddedInVersion      string `json:"added_in_version,omitempty"`
	DeprecatedInVersion string `json:"deprecated_in_version,omitempty"`
}

const baseDocLink string = "https://developer.android.com/reference"
const baseSourceCodeLink string = "https://android.googlesource.com"

var whitespaceRegexp = regexp.MustCompile("\\s{2,}")
var sourceMappings [][]string

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

	librariesFile, _ := os.Open("libraries.csv")
	r := csv.NewReader(librariesFile)
	libraries, _ := r.ReadAll()

	sourceMappingsFile, _ := os.Open("source_mapping.csv")
	r = csv.NewReader(sourceMappingsFile)
	sourceMappings, _ = r.ReadAll()

	for _, library := range libraries {
		libraryClasses = append(libraryClasses, LibraryClasses{library[0], 0, []ApiClass{}})
		link := baseDocLink + library[1]
		_ = c.Visit(link)
	}
	for _, libraryClass := range libraryClasses {
		allClasses.ClassCount += libraryClass.ClassCount
	}
	allClasses.Classes = libraryClasses

	jsonData, _ := json.MarshalIndent(allClasses, "", "  ")

	_ = ioutil.WriteFile("classes_index.json", jsonData, os.ModePerm)
	log.Println("Found", allClasses.ClassCount, "classes in", len(libraries), "libraries")
}

func parseApiClass(e *colly.HTMLElement) ApiClass {
	name := e.DOM.Find("td[class=jd-linkcol]>a[href]").First().Text()
	nameExtended := e.ChildText("td[class=jd-linkcol]")
	if nameExtended == name {
		nameExtended = "" // to omit field in json
	}
	link := e.ChildAttr("td[class=jd-linkcol]>a[href]", "href")
	sourceLink := ""
	for _, mapping := range sourceMappings {
		if strings.HasPrefix(link, baseDocLink+mapping[0]) {
			classDocTitle := link[len(baseDocLink+mapping[0]):]
			title := classDocTitle[:strings.Index(classDocTitle, ".")]
			if strings.HasSuffix(title, "R") { // exclude auto-generated R classes
				break
			}
			sourceLink = baseSourceCodeLink + mapping[1] + title + ".java"
		}
	}
	description := e.ChildText("td[class=jd-descrcol]")
	description = whitespaceRegexp.ReplaceAllString(description, " ")
	addedInVersion := e.Attr("data-version-added")
	deprecatedInVersion := e.Attr("data-version-deprecated")

	return ApiClass{
		Name:                name,
		NameExtended:        nameExtended,
		Link:                link,
		SourceLink:          sourceLink,
		Description:         description,
		AddedInVersion:      addedInVersion,
		DeprecatedInVersion: deprecatedInVersion,
	}
}
