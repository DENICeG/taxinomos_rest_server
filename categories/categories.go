package categories

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
)

type CategoryList struct {
	Meta struct {
		Page struct {
			CurrentPage int `json:"current-page"`
			PerPage     int `json:"per-page"`
			From        int `json:"from"`
			To          int `json:"to"`
			Total       int `json:"total"`
			LastPage    int `json:"last-page"`
		} `json:"page"`
	} `json:"meta"`
	Links struct {
		First string `json:"first"`
		Next  string `json:"next"`
		Last  string `json:"last"`
	} `json:"links"`
	Data []*Category `json:"data"`
}

type Category struct {
	//Data struct {
	Type               string `json:"type"`
	ID                 string `json:"id"`
	CategoryAttributes struct {
		WebsiteCategoryId int    `json:"website-category-id"`
		MainCategory      string `json:"maincategory"`
		SubCategory       string `json:"subcategory"`
		Language          string `json:"lang"`
		Description       string `json:"description"`
	} `json:"attributes"`
	CategoryLinks struct {
		Self string `json:"self"`
	} `json:"links"`
	CategoryRelationships struct {
		Measurements struct {
			Meta struct {
				Total int `json:"total"`
			} `json:"meta"`
			Links struct {
				Self    string `json:"self"`
				Related string `json:"related"`
			} `json:"links"`
		} `json:"measurements"`
	} `json:"relationships"`
	//}
}

type Categories []Categories

func LoadCategoriesFromFile(filename string, categories interface{}) error {
	categoriesbytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(categoriesbytes, categories)
	if err != nil {
		return err
	}
	return nil
}

func LoadCategoriesFromCsvFile(filename string) ([]*Category, error) {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(fileContent)
	csvreader := csv.NewReader(reader)
	csvreader.Comma = ';'

	records, err := csvreader.ReadAll()
	if err != nil {
		return nil, err
	}
	var categories []*Category
	for _, v := range records {
		id, _ := strconv.Atoi(v[0])
		cat := new(Category)
		cat.Type = "categories"
		cat.ID = v[0]
		cat.CategoryAttributes.WebsiteCategoryId = id
		cat.CategoryAttributes.MainCategory = v[1]
		cat.CategoryAttributes.SubCategory = v[2]
		//cat.CategoryAttributes.Language = v[]
		cat.CategoryAttributes.Description = v[3]
		cat.CategoryLinks.Self = fmt.Sprintf("https://localhost:8080/api/v1/categories/%s", v[0])
		categories = append(categories, cat)

	}

	return categories, nil
}

func CreateCategoryList(categories []*Category) (*CategoryList, error) {
	catlist := new(CategoryList)
	for _, cat := range categories {
		catlist.Data = append(catlist.Data, cat)
	}
	catlist.Meta.Page.CurrentPage = 1
	catlist.Meta.Page.LastPage = 1
	catlist.Meta.Page.From = 1
	catlist.Meta.Page.To = len(catlist.Data) + 1
	catlist.Meta.Page.Total = len(catlist.Data) + 1
	catlist.Meta.Page.PerPage = len(catlist.Data) + 1

	catlist.Links.First = fmt.Sprintf("https://classify-rest.labs.nic.at/api/v1/categories?page%%5Bnumber%%5D=1&page%%5Bsize%%5D=%d", catlist.Meta.Page.Total)
	catlist.Links.Next = fmt.Sprintf("https://classify-rest.labs.nic.at/api/v1/categories?page%%5Bnumber%%5D=1&page%%5Bsize%%5D=%d", catlist.Meta.Page.Total)
	catlist.Links.Last = fmt.Sprintf("https://classify-rest.labs.nic.at/api/v1/categories?page%%5Bnumber%%5D=1&page%%5Bsize%%5D=%d", catlist.Meta.Page.Total)

	return catlist, nil
}
