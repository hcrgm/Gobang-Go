package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/kataras/go-sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/random"
	"gobang"
	"golang.org/x/net/html"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Template struct {
	templates *template.Template
}

type Config struct {
	debug    bool
	port     int
	useOAuth bool
	github   *OAuthConfig
}

type OAuthConfig struct {
	client_id     string
	client_secret string
}

var config *Config

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	configFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	json, err := simplejson.NewJson(configFile)
	if err != nil {
		panic(err)
	}
	config = &Config{
		debug:    json.Get("debug").MustBool(false),
		port:     json.Get("port").MustInt(8011),
		useOAuth: json.Get("useOAuth").MustBool(false),
		github: &OAuthConfig{
			client_id:     json.GetPath("github").Get("client_id").MustString(""),
			client_secret: json.GetPath("github").Get("client_secret").MustString(""),
		},
	}
	if config.useOAuth {
		if len(config.github.client_id) == 0 || len(config.github.client_secret) == 0 {
			panic("Wrong config:Github OAuth")
		}
	}
	e := echo.New()
	// debug
	e.Debug = config.debug
	e.Use(middleware.Gzip())
	e.Static("/", "public")
	if config.debug {
		e.Use(middleware.Logger())
	}
	t := &Template{
		templates: template.Must(template.New("").Funcs(template.FuncMap{
			"loop": func(n int) []int {
				slice := make([]int, n)
				for i := 0; i < n; i++ {
					slice[i] = i + 1
				}
				return slice
			},
		}).ParseGlob("template/*.html")),
	}
	e.Renderer = t
	e.GET("/status", gobang.HandleStatusSocket)
	e.GET("/socket", gobang.HandleGameSocket)
	e.GET("/game", gobang.Game)
	e.Any("/login", Login)
	e.Logger.Fatal(e.Start(":" + strconv.Itoa(config.port)))
}

func Login(c echo.Context) error {
	if !config.useOAuth {
		return c.HTML(http.StatusOK, `[<b style="color:#f44336">Login function not enabled</b>]`)
	}
	response := ""
	w := c.Response().Writer()
	r := c.Request()
	sess := sessions.Start(w, r)
	switch c.FormValue("action") {
	case "logout":
		sess.Delete("name")
		fallthrough
	case "info":
		username := "Anonymous"
		signBtn := "Sign in"
		if name := sess.GetString("name"); len(name) != 0 {
			username = name
			signBtn = "Sign out"
		}
		response = fmt.Sprintf(`<span id="username">%s</span>&nbsp;&nbsp;<a id="btn_login">[%s]</a><script>$("#btn_login").one("click", login);</script>`, username, signBtn)
	case "oauth":
		sha := sha1.New()
		sha.Write([]byte(random.String(16)))
		state := fmt.Sprintf("%x", sha.Sum(nil))
		sess.Set("state", state)
		return c.Redirect(http.StatusFound, "https://github.com/login/oauth/authorize?client_id="+config.github.client_id+"&state="+state)
	case "oauth-callback":
		if len(c.FormValue("code")) == 0 || len(c.FormValue("state")) == 0 || sess.GetString("state") != c.FormValue("state") {
			return c.HTML(http.StatusBadRequest, "Bad Request")
		}
		if code := c.FormValue("code"); len("code") != 0 {
			v := url.Values{}
			v.Set("client_id", config.github.client_id)
			v.Set("client_secret", config.github.client_secret)
			v.Set("code", code)
			req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(v.Encode()))
			if err != nil {
				return c.HTML(http.StatusInternalServerError, err.Error())
			}
			req.Header.Add("Accept", "application/json")
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response, err := http.DefaultClient.Do(req)
			if err != nil {
				return c.HTML(http.StatusInternalServerError, "Cannot send the request to GitHub")
			}
			json, err := simplejson.NewFromReader(response.Body)
			if err != nil {
				return c.HTML(http.StatusInternalServerError, "Cannot parse the result")
			}
			accessToken := json.Get("access_token").MustString("")
			if len(accessToken) == 0 {
				return c.HTML(http.StatusInternalServerError, "Cannot get the access token")
			}
			v = url.Values{}
			v.Set("access_token", accessToken)
			req, err = http.NewRequest("GET", "https://api.github.com/user", nil)
			if err != nil {
				return c.HTML(http.StatusInternalServerError, err.Error())
			}
			req.Header.Set("Authorization", "token "+accessToken)
			response, err = http.DefaultClient.Do(req)
			if err != nil {
				return c.HTML(http.StatusInternalServerError, "Cannot send the request to GitHub")
			}
			json, err = simplejson.NewFromReader(response.Body)
			if err != nil {
				return c.HTML(http.StatusInternalServerError, "Cannot parse the result")
			}
			name := json.Get("name").MustString("")
			if len(name) == 0 {
				name = json.Get("login").MustString("")
				if len(name) == 0 {
					return c.HTML(http.StatusInternalServerError, "Cannot get the name")
				}
			}
			name = html.EscapeString(name)
			sess.Set("name", name)
			return c.Redirect(http.StatusFound, "index.html")
		}
	}
	return c.HTML(http.StatusOK, response)
}
