package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
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

type Variables struct {
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

type VariableSet struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Name           string    `json:"name"`
			Description    string    `json:"description"`
			Global         bool      `json:"global"`
			UpdatedAt      time.Time `json:"updated-at"`
			VarCount       int       `json:"var-count"`
			WorkspaceCount int       `json:"workspace-count"`
		} `json:"attributes"`
		Relationships struct {
			Organization struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
			} `json:"organization"`
			Vars struct {
				Data []struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
			} `json:"vars"`
		} `json:"relationships"`
	} `json:"data"`
}

type NewVariable struct {
	Data struct {
		Type       string `json:"type"`
		Attributes struct {
			Category    string `json:"category"`
			Key         string `json:"key"`
			Value       string `json:"value"`
			Description string `json:"description"`
			Sensitive   string `json:"sensitive"`
			Hcl         string `json:"hcl"`
		} `json:"attributes"`
	} `json:"data"`
}

func getTerraformEnvs() []string {
	var resp []string
	for _, v := range getConfig().Terraform {
		resp = append(resp, v.Name)
	}

	return resp
}

func getVarsetsId(s string, c Config) string {
	var resp string
	for _, v := range c.Terraform {
		if v.Name == s {
			resp = v.Varset
		}
	}
	return resp
}

func getVars(s string) (*Variables, error) {
	terraConfig := getConfig()
	url := fmt.Sprintf("https://app.terraform.io/api/v2/varsets/%s/relationships/vars?page%ssize%s=100", getVarsetsId(s, terraConfig), "%5B", "%5D")
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
	var result Variables

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func updateVars(e string, v Variable, u string) (string, error) {
	v.Data.Attributes.Value = u
	data, _ := json.Marshal(v)
	url := fmt.Sprintf("https://app.terraform.io/api/v2/vars/%s", v.Data.ID)
	client := &http.Client{}
	request, _ := http.NewRequest("PATCH", url, strings.NewReader(string(data)))
	request.Header.Add("Authorization", "Bearer "+getConfig().Token)
	request.Header.Add("Content-Type", "application/vnd.api+json")
	resp, _ := client.Do(request)

	if resp.StatusCode != http.StatusOK {
		resBody, _ := ioutil.ReadAll(resp.Body)
		return v.Data.Attributes.Key, fmt.Errorf("client: response body: %s", resBody)
	}

	return v.Data.Attributes.Key, nil
}

// getVars("Ambientes Bajos", "ENV_APP_FOO", "FOO_VALUE")
func addVars(s string, k string, v string) error {
	terraConfig := getConfig()
	varSetId := getVarsetsId(s, terraConfig)
	varSet, err := getVarset(varSetId, terraConfig)

	if err != nil {
		return err
	}

	var newVar NewVariable
	newVar.Data.Type = varSet.Data.Type
	newVar.Data.Attributes.Category = "terraform"
	newVar.Data.Attributes.Key = k
	newVar.Data.Attributes.Value = v
	newVar.Data.Attributes.Sensitive = "true"
	newVar.Data.Attributes.Hcl = "false"
	data, _ := json.Marshal(newVar)

	url := fmt.Sprintf("https://app.terraform.io/api/v2/varsets/%s/relationships/vars", varSetId)
	client := &http.Client{}
	request, _ := http.NewRequest("POST", url, strings.NewReader(string(data)))
	request.Header.Add("Authorization", "Bearer "+terraConfig.Token)
	request.Header.Add("Content-Type", "application/vnd.api+json")
	resp, err := client.Do(request)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

// getVarset("varsetid", config_json)
func getVarset(s string, c Config) (*VariableSet, error) {
	url := fmt.Sprintf("https://app.terraform.io/api/v2/varsets/%s", s)
	client := &http.Client{}
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("Authorization", "Bearer "+c.Token)
	request.Header.Add("Content-Type", "application/vnd.api+json")
	resp, err := client.Do(request)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var result VariableSet

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
