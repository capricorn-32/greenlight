package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	// Initialize a new httprouter instance.
	router := httprouter.New()

	// custome handler for 404 Method not allowed responses.
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	// custome handler for 405 Method not allowed responses.
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Remove the trailing slash '/' from the URL and redirect it.
	// example route /foo/ will redirect to /foo.
	router.RedirectTrailingSlash = true

	// registering routes
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

	// Return the httprouter instance.
	return router
}
