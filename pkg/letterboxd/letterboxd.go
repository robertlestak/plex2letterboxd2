package letterboxd

import (
	"encoding/csv"
	"os"
)

type Entry struct {
	Title       string
	Year        string
	IMDbID      string
	Rating10    string
	WatchedDate string
}

func WriteCSV(filename string, entries []Entry) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"Title", "Year", "imdbID", "Rating10", "WatchedDate"}); err != nil {
		return err
	}

	for _, e := range entries {
		if err := w.Write([]string{e.Title, e.Year, e.IMDbID, e.Rating10, e.WatchedDate}); err != nil {
			return err
		}
	}

	return nil
}
