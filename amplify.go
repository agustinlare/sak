package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Variable struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Key         string      `json:"key"`
			Value       interface{} `json:"value"`
			Sensitive   bool        `json:"sensitive"`
			Category    string      `json:"category"`
			Hcl         bool        `json:"hcl"`
			CreatedAt   time.Time   `json:"created-at"`
			Description interface{} `json:"description"`
		} `json:"attributes"`
		Relationships struct {
			Varset struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
				Links struct {
					Related string `json:"related"`
				} `json:"links"`
			} `json:"varset"`
		} `json:"relationships"`
		Links struct {
			Self string `json:"self"`
		} `json:"links"`
	} `json:"data"`
}

type Response struct {
	Data []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Key         string      `json:"key"`
			Value       interface{} `json:"value"`
			Sensitive   bool        `json:"sensitive"`
			Category    string      `json:"category"`
			Hcl         bool        `json:"hcl"`
			CreatedAt   time.Time   `json:"created-at"`
			Description interface{} `json:"description"`
		} `json:"attributes"`
		Relationships struct {
			Varset struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
				Links struct {
					Related string `json:"related"`
				} `json:"links"`
			} `json:"varset"`
		} `json:"relationships"`
		Links struct {
			Self string `json:"self"`
		} `json:"links"`
	} `json:"data"`
	Links struct {
		Self  string      `json:"self"`
		First string      `json:"first"`
		Prev  interface{} `json:"prev"`
		Next  interface{} `json:"next"`
		Last  string      `json:"last"`
	} `json:"links"`
	Meta struct {
		Pagination struct {
			CurrentPage int         `json:"current-page"`
			PageSize    int         `json:"page-size"`
			PrevPage    interface{} `json:"prev-page"`
			NextPage    interface{} `json:"next-page"`
			TotalPages  int         `json:"total-pages"`
			TotalCount  int         `json:"total-count"`
		} `json:"pagination"`
	} `json:"meta"`
}

func amplify(c *gin.Context) {
	if c.Request.Method == "GET" {
		reply(c, true, "Amplify")
		c.HTML(
			http.StatusOK,
			"views/amplify.html",
			gin.H{
				"env": getEnv(),
			},
		)
	}

	if c.Request.Method == "POST" {
		var htmlVars gin.H

		if len(c.PostForm("action")) > 0 {
			htmlVars = gin.H{
				"act_selection": c.PostForm("action"),
				"env":           getEnv(),
			}
		}

		var varSingle Variable
		envSelection := c.PostForm("env_selection")
		varSelection := c.PostForm("var_selection")

		if len(envSelection) > 0 {
			listVars, err := getVars(envSelection)

			if err != nil {
				log.Println(err.Error())
			}

			if len(c.PostForm("var_selection")) > 0 {
				for i, v := range listVars.Data {
					if v.Attributes.Key == varSelection {
						varSingle.Data = listVars.Data[i]
						break
					}
				}

				resp, err := updateVars(envSelection, varSingle, c.PostForm("var_update"))
				if err != nil {
					htmlVars = gin.H{
						"error": err,
						"env":   getEnv(),
					}
				}

				var v Variable
				v.Data.Attributes.Key = varSelection

				htmlVars = gin.H{
					"env_selected": envSelection,
					"var_selected": varSelection,
					"messages":     "Updated terraform var " + resp,
				}

			} else {
				htmlVars = gin.H{
					"env_selected": envSelection,
					"vars":         listVars.Data,
				}
			}
		}
		reply(c, true, "Amplify")
		c.HTML(
			http.StatusOK,
			"views/amplify.html",
			htmlVars,
		)
	}
}

func getEnv() []string {
	var resp []string
	for _, v := range getConfig().Env {
		resp = append(resp, v.Name)
	}

	return resp
}

func getVarset(s string, c Config) string {
	var resp string
	for _, v := range c.Env {
		if v.Name == s {
			resp = v.Varset
		}
	}
	return resp
}

func getVars(s string) (*Response, error) {
	terraConfig := getConfig()
	url := fmt.Sprintf("https://app.terraform.io/api/v2/varsets/%s/relationships/vars", getVarset(s, terraConfig))
	client := &http.Client{}
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("Authorization", "Bearer "+terraConfig.Token)
	request.Header.Add("Content-Type", "application/vnd.api+json")
	resp, err := client.Do(request)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var result Response

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func updateVars(e string, v Variable, u string) (string, error) {
	v.Data.Attributes.Value = u
	data, _ := json.Marshal(v)
	r := strings.NewReader(string(data))
	url := fmt.Sprintf("https://app.terraform.io/api/v2/vars/%s", v.Data.ID)
	client := &http.Client{}
	request, _ := http.NewRequest("PATCH", url, r)
	request.Header.Add("Authorization", "Bearer "+getConfig().Token)
	request.Header.Add("Content-Type", "application/vnd.api+json")
	resp, _ := client.Do(request)

	if resp.StatusCode != http.StatusOK {
		resBody, _ := ioutil.ReadAll(resp.Body)
		return v.Data.Attributes.Key, fmt.Errorf("client: response body: %s", resBody)
	}

	return v.Data.Attributes.Key, nil
}
