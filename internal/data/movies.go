package data

import "time"

type Movie struct {
	ID        int64     `json:"id"`         // Unique integer ID for the movie
	CreatedAt time.Time `json:"created_at"` // Timestamp for when the movie is added to our database
	Title     string    `json:"title"`      // Movie Title
	Year      int32     `json:"year"`       // Movie release year
	Runtime   Runtime   `json:"runtime"`    // Movie Runtime (in minutes)
	Genres    []string  `json:"genres"`     // Slice of genres for the movie.
	Version   int32     `json:"version"`    // The version number starts at 1 and will be incremented each time the movie information is updated.
}
