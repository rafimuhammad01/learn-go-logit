package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	logrustash "github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

func makeLogEntry(c echo.Context) *logrus.Entry {
	log := logrus.New()

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", os.Getenv("LOGSTASH_HOST"), os.Getenv("LOGSTASH_PORT")), &tls.Config{RootCAs: nil})
	if err != nil {
		log.Fatal(err)
	}

	var hook logrus.Hook
	if c == nil {
		hook = logrustash.New(conn, logrustash.DefaultFormatter(logrus.Fields{
			"type": "law",
		}))
	} else {
		hook = logrustash.New(conn, logrustash.DefaultFormatter(logrus.Fields{
			"type":   "law",
			"uri":    c.Request().RequestURI,
			"method": c.Request().Method,
			"param":  c.Request().URL.Query(),
			"body":   c.Request().Body,
		}))
	}

	log.Hooks.Add(hook)

	return log.WithTime(time.Now())
}

func middlewareLogging(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		makeLogEntry(c).Info("incoming request")
		return next(c)
	}
}

func main() {

	e := echo.New()
	e.Use(middlewareLogging)

	err := godotenv.Load()
	if err != nil {
		makeLogEntry(nil).Error("Error loading .env file")
	}

	e.GET("/echo", func(c echo.Context) error {
		param := c.QueryParam("param")
		if param == "" {
			param = "Silahkan masukan parameter di URL :)"
		}

		return c.HTML(http.StatusOK, fmt.Sprintf("<h1>%s</h1>", param))
	})

	makeLogEntry(nil).Fatal(e.Start(":" + os.Getenv("PORT")))
}
