package main

const (
	DEFAULT_GO_CODE string = `package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Static("/", "dist")
	e.Use(middleware.Static("/dist"))

	e.GET("/api/hello", func(ctx echo.Context) error {
		return ctx.JSON(
			http.StatusOK, echo.Map{
				"message": "Hello World!",
			},
		)
	})

	e.Logger.Fatal(e.Start("%v"))
}
	`

	DEFAULT_VITE_CODE string = `import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vite.dev/config/
export default defineConfig({
	plugins: [svelte()],
	base: '/',
	server: {
		proxy: {
		'/api': 'http://localhost%v', // Proxy API calls to Echo
		}
	},
	build: {
		outDir: '../dist',
		emptyOutDir: true,
	}
})
	`
)
