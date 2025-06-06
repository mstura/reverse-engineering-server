package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/client"
)

func main() {
	if len(os.Args) != 2 {
		panic("no target provided")
	}

	realTarget := os.Args[1]

	app := fiber.New()

	app.All("*", func(c fiber.Ctx) error {
		targetUrl, _ := url.Parse(realTarget)
		cc := client.New()
		cc.SetTimeout(10 * time.Second)
		th := map[string]string{}

		c.Request().Header.Set("Host", targetUrl.Host)
		targetUrl = targetUrl.JoinPath(c.OriginalURL())
		m := c.Method()

		println(m, c.OriginalURL(), "->", m, targetUrl.String())
		fmt.Printf("%s", c.RequestCtx().Request.Header.RawHeaders())

		for k, v := range c.GetReqHeaders() {
			th[k] = strings.Join(v, ", ")
		}

		var cres *client.Response
		var err error

		cc.SetHeaders(th)
		cc.R().SetRawBody(c.Request().Body())

		switch c.Method() {
		case "GET":
			cres, err = cc.Get(targetUrl.String())
		case "DELETE":
			cres, err = cc.Delete(targetUrl.String())
		case "POST":
			cres, err = cc.Post(targetUrl.String())
		case "PUT":
			cres, err = cc.Put(targetUrl.String())
		default:
			fmt.Printf("Method %s proxying not supported, TODO", c.Method())
			return c.SendStatus(500)
		}

		if err != nil {
			return err
		}

		c.Status(cres.StatusCode())
		c.Response().CloseBodyStream()

		cres.Save(io.MultiWriter(c.Response().BodyWriter(), os.Stdout))

		return c.Response().CloseBodyStream()
	})

	app.Listen(":8080", fiber.ListenConfig{
		BeforeServeFunc: func(app *fiber.App) error {
			colors := app.Config().ColorScheme

			fmt.Printf("%sINFO%s Spying on target (%s%s%s)\n\n", colors.Green, colors.Reset, colors.Red, realTarget, colors.Reset)

			return nil
		},
	})
}
