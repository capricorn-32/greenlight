package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"greenlight.abhishek/internal/validator"
)

type Movie struct {
	ID        int64     `json:"id"`         // Unique integer ID for the movie
	CreatedAt time.Time `json:"created_at"` // Timestamp for when the movie is added to our database
	Title     string    `json:"title"`      // Movie Title
	Year      int32     `json:"year"`       // Movie release year
	Runtime   Runtime   `json:"runtime"`    // Movie Runtime (in minutes)
	Genres    []string  `json:"genres"`     // Slice of genres for the movie.
	Version   int32     `json:"version"`    // The version number starts at 1 and will be incremented each time the movie information is updated.
}

type MovieModel struct {
	DB *sql.DB
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	// Title validation
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	// Year validation
	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greator than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	// Runtime validation
	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integers")

	// Genres validation
	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

func (m MovieModel) Insert(movie *Movie) error {
	query := `
		INSERT INTO movies (title, year, runtime genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version
	`
	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
	}

	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// SQL query for retrieving the movie data
	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE id = $1
	`

	// Movie struct to hold the data returned by the query.
	var movie Movie
	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

func (m MovieModel) Update(movie *Movie) error {
	return nil
}

func (m MovieModel) Delete(id int64) error {
	return nil
}

// // Implement a MarshalJSON() method on the Movie struct, so that it satisfies the
// // json.Marshaler interface.
// func (m Movie) MarshalJSON() ([]byte, error) {
// 	// Declare a variable to hold the custom runtime string (this will be the empty
// 	// string "" by default).
// 	var runtime string
// 	// If the value of the Runtime field is not zero, set the runtime variable to be a
// 	// string in the format "<runtime> mins".
// 	if m.Runtime != 0 {
// 		runtime = fmt.Sprintf("%d mins", m.Runtime)
// 	}
// 	// Create an anonymous struct to hold the data for JSON encoding. This has exactly
// 	// the same fields, types and tags as our Movie struct, except that the Runtime
// 	// field here is a string, instead of an int32. Also notice that we don't include
// 	// a CreatedAt field at all (there's no point including one, because we don't want
// 	// it to appear in the JSON output).
// 	aux := struct {
// 		ID      int64    `json:"id"`
// 		Title   string   `json:"title"`
// 		Year    int32    `json:"year"`
// 		Runtime string   `json:"runtime"` // This is a string.
// 		Genres  []string `json:"genres"`
// 		Version int32    `json:"version"`
// 	}{
// 		// Set the values for the anonymous struct.
// 		ID:      m.ID,
// 		Title:   m.Title,
// 		Year:    m.Year,
// 		Runtime: runtime, // Note that we assign the value from the runtime variable here.
// 		Genres:  m.Genres,
// 		Version: m.Version,
// 	}
// 	// Encode the anonymous struct to JSON, and return it.
// 	return json.Marshal(aux)
// }
