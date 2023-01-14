package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func index(c *gin.Context) {
	c.HTML(
		http.StatusOK,
		"views/index.html",
		gin.H{
			"healthstatus": getHealth(),
			"links":        getLinks(),
		},
	)
}

func indexPost(c *gin.Context) {
	if c.PostForm("healthswitch-btn") != "" {
		switchHealth()
	}

	c.HTML(
		http.StatusOK,
		"views/index.html",
		gin.H{
			"healthstatus": getHealth(),
			"links":        getLinks(),
		},
	)
}

func health(c *gin.Context) {
	// Info.Println("OK")
	c.JSON(http.StatusOK, "OK")
}

func unhealth(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, "Service Unavailable")
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, "Pong")
}

func rickroll(c *gin.Context) {
	c.Redirect(http.StatusFound, "https://www.youtube.com/watch?v=dQw4w9WgXcQ")
}

func conTest(c *gin.Context) {
	method := c.Request.Method

	if method == "GET" {
		c.HTML(
			http.StatusOK,
			"views/con-test.html",
			gin.H{
				"links": getLinks(),
			},
		)
	}

	if method == "POST" {
		var resp string
		var err error
		var headers gin.H

		if c.PostForm("ipcheck") != "" {
			resp, err = isReachable(c.PostForm("ipcheck"))
		} else if c.PostForm("dns") != "" {
			resp, err = dnsResolver(c.PostForm("dns"))
		} else if c.PostForm("mongodb") != "" {
			resp, err = mongodb(c.PostForm("mongodb"))
		}

		if err != nil {
			Error.Println(err.Error())
			headers = gin.H{
				"error": err,
				"links": getLinks(),
			}

		} else {
			Info.Println(resp)
			headers = gin.H{
				"messages": resp,
				"links":    getLinks(),
			}
		}

		c.HTML(
			http.StatusOK,
			"views/con-test.html",
			headers,
		)
	}
}

func vtDropfile(c *gin.Context) {
	if c.Request.Method == "GET" {
		c.HTML(
			http.StatusOK,
			"views/vt-deploy.html",
			gin.H{
				"links": getLinks(),
			},
		)
	}

	if c.Request.Method == "POST" {
		resp, err := vtUpload(c)

		if err != nil {
			Error.Println(err)
			c.HTML(
				http.StatusInternalServerError,
				"",
				gin.H{
					"links": getLinks(),
				},
			)
			return
		}

		os.Setenv("UPLOADED_VT_FILE", resp)
		Warning.Println(resp)
		c.HTML(
			http.StatusOK,
			"views/vt-deploy.html",
			gin.H{
				"links": getLinks(),
			},
		)
	}
}

func invokeLambda(c *gin.Context) {
	user := c.PostForm("username")
	pass := c.PostForm("password")

	encText, err := encrypt(pass, getConfig().Secretstring)
	if err != nil {
		Error.Printf("Unable to verify password: %s", err)
		c.JSON(http.StatusInternalServerError, "Server Internal Error")
		return
	}

	retrived_password, err := getPassword(user)
	if err != nil {
		Error.Printf("Unable to retrive password for user %s: %s", user, err.Error())
		c.JSON(http.StatusInternalServerError, "Server Internal Error")
		return
	}

	if encText != retrived_password {
		Error.Printf("%s access was denied by admin", user)
		c.JSON(http.StatusForbidden, "Forbidden")
		return
	}

	action := c.PostForm("action")
	env := c.PostForm("env")
	err = actionHandler(action, env, user)

	if err != nil {
		Error.Println(err)
		c.JSON(http.StatusInternalServerError, "Server Internal Error")
		return
	}

	c.JSON(http.StatusOK, fmt.Sprintf("Action %s plan should end in a few minutes. Executed by user %s", action, user))

}

func newUser(c *gin.Context) {
	data := c.Request.URL.Query()

	encText, err := encrypt(data["password"][0], getConfig().Secretstring)
	if err != nil {
		Error.Printf("Unable to encrypt data user %s: %s", data["password"][0], err)
		c.JSON(http.StatusInternalServerError, "Server Internal Error")
		return
	}

	err = createUser(data["username"][0], encText)
	if err != nil {
		Error.Printf("Unable to create user: %s", err)
		c.JSON(http.StatusInternalServerError, "Server Internal Error")
		return
	}

	c.JSON(http.StatusCreated, "Created")
}

func retrive(c *gin.Context) {
	decText, err := decrypt(c.Request.URL.Query()["secret"][0], getConfig().Secretstring)
	if err != nil {
		Error.Printf("Unable to decrypt string: %s", err)
		c.JSON(http.StatusInternalServerError, "Server Internal Error")
	}

	c.JSON(http.StatusOK, decText)
}

func amplify(c *gin.Context) {
	if c.Request.Method == "GET" {
		c.HTML(
			http.StatusOK,
			"views/amplify.html",
			gin.H{
				"env": getTerraformEnvs(),
			},
		)
	}

	if c.Request.Method == "POST" {
		var htmlVars gin.H
		var actSelection string = c.PostForm("action")

		if len(c.PostForm("action")) > 0 {
			htmlVars = gin.H{
				"act_selected": actSelection,
				"env":          getTerraformEnvs(),
			}
		}

		var envSelection string = c.PostForm("env_selection")
		var secretString string = c.PostForm("secret")

		if len(envSelection) > 0 && c.PostForm("action") == "edit" {
			listVars, err := getVars(envSelection)

			if err != nil {
				Error.Println(err.Error())
				htmlVars = gin.H{
					"env_selected": envSelection,
					"act_selected": actSelection,
					"error":        err.Error(),
				}

			} else {
				htmlVars = gin.H{
					"env_selected": envSelection,
					"act_selected": actSelection,
					"vars":         listVars.Data,
				}
			}
		}

		if len(secretString) > 0 {
			encText, err := encrypt(secretString, getConfig().Secretstring)
			if err != nil {
				Error.Printf("Unable to verify password: %s", err)
				c.JSON(http.StatusInternalServerError, "Server Internal Error")
				return
			}

			retrived_password, err := getPassword("amplify")
			if err != nil {
				Error.Printf("Unable to retrive password for user: %s", err.Error())
				c.JSON(http.StatusInternalServerError, "Server Internal Error")
				return
			}

			if encText != retrived_password {
				Error.Printf("access was denied by admin")
				c.JSON(http.StatusForbidden, "forbidden")
				return
			}

			switch actSelection {
			case "new":
				key := c.PostForm("var_name")
				value := c.PostForm("var_value")

				err := addVars(envSelection, key, value)
				if err != nil {
					htmlVars = gin.H{"error": err, "env": getTerraformEnvs(), "act_selected": actSelection}
				} else {
					htmlVars = gin.H{
						"messages": fmt.Sprintf("New variable %s created in %s", key, envSelection),
						"env":      getTerraformEnvs(), "act_selected": actSelection,
					}
				}

			case "edit":
				var varSingle Variable
				varSelection := c.PostForm("var_selection")
				listVars, _ := getVars(envSelection)

				for i, v := range listVars.Data {
					if v.Attributes.Key == varSelection {
						varSingle.Data = listVars.Data[i]
						break
					}
				}

				resp, err := updateVars(envSelection, varSingle, c.PostForm("var_update"))
				if err != nil {
					Error.Println(err)
					htmlVars = gin.H{
						"error":        err.Error(),
						"env":          getTerraformEnvs(),
						"act_selected": actSelection,
					}
				} else {
					htmlVars = gin.H{
						"env_selected": envSelection,
						"var_selected": varSelection,
						"act_selected": actSelection,
						"messages":     "Updated terraform var " + resp,
					}
				}
			}
		}

		c.HTML(
			http.StatusOK,
			"views/amplify.html",
			htmlVars,
		)
	}
}
