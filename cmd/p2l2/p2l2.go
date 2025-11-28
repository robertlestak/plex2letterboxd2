package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/robertlestak/plex2letterboxd2/pkg/letterboxd"
	log "github.com/sirupsen/logrus"
)

func init() {
	ll, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		ll = log.InfoLevel
	}
	log.SetLevel(ll)
}

type MediaContainer struct {
	Directories []Directory `xml:"Directory"`
	Videos      []Video     `xml:"Video"`
}

type Directory struct {
	Key   string `xml:"key,attr"`
	Type  string `xml:"type,attr"`
	Title string `xml:"title,attr"`
}

type Video struct {
	Title        string  `xml:"title,attr"`
	Year         int     `xml:"year,attr"`
	ViewCount    int     `xml:"viewCount,attr"`
	LastViewedAt int64   `xml:"lastViewedAt,attr"`
	UserRating   float64 `xml:"userRating,attr"`
	GUID         string  `xml:"guid,attr"`
	Guids        []Guid  `xml:"Guid"`
}

type Guid struct {
	ID string `xml:"id,attr"`
}

func main() {
	l := log.WithFields(log.Fields{
		"fn":  "main",
		"app": "plex2letterboxd2",
	})
	l.Info("plex2letterboxd2 started")
	plexURL := flag.String("plex-url", os.Getenv("PLEX_URL"), "Plex server URL")
	plexToken := flag.String("plex-token", os.Getenv("PLEX_TOKEN"), "Plex authentication token")
	output := flag.String("output", "letterboxd.csv", "Output CSV file")
	importToLetterboxd := flag.Bool("import", true, "Import the generated CSV to Letterboxd")
	initPlaywright := flag.Bool("init-playwright", false, "Initialize Playwright for Letterboxd import")
	initPlaywrightAndExit := flag.Bool("init-playwright-only", false, "Initialize Playwright and exit")
	letterboxdUsername := flag.String("letterboxd-username", os.Getenv("LETTERBOXD_USERNAME"), "Letterboxd username")
	letterboxdPassword := flag.String("letterboxd-password", os.Getenv("LETTERBOXD_PASSWORD"), "Letterboxd password")
	flag.Parse()

	if *initPlaywright || *initPlaywrightAndExit {
		if err := letterboxd.InitPlaywright(); err != nil {
			l.WithError(err).Fatal("could not initialize Playwright")
		}
		l.Info("Playwright initialized successfully")
		if *initPlaywrightAndExit {
			return
		}
	}

	if *plexURL == "" || *plexToken == "" {
		l.Fatal("PLEX_URL and PLEX_TOKEN are required")
	}

	if *importToLetterboxd {
		if *letterboxdUsername == "" || *letterboxdPassword == "" {
			l.Fatal("LETTERBOXD_USERNAME and LETTERBOXD_PASSWORD are required for import")
		}
	}

	l.Info("Fetching library sections...")
	sectionsURL := fmt.Sprintf("%s/library/sections?X-Plex-Token=%s", *plexURL, url.QueryEscape(*plexToken))
	resp, err := http.Get(sectionsURL)
	if err != nil {
		l.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Fatal(err)
	}

	var sections MediaContainer
	if err := xml.Unmarshal(body, &sections); err != nil {
		l.Fatal(err)
	}

	var entries []letterboxd.Entry

	for _, section := range sections.Directories {
		if section.Type != "movie" {
			continue
		}

		l.Infof("Fetching movies from section: %s", section.Title)
		moviesURL := fmt.Sprintf("%s/library/sections/%s/all?includeGuids=1&X-Plex-Token=%s", *plexURL, section.Key, url.QueryEscape(*plexToken))
		resp, err := http.Get(moviesURL)
		if err != nil {
			l.Warnf("Error fetching section %s: %v", section.Title, err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			l.Warnf("Error reading section %s: %v", section.Title, err)
			continue
		}

		var movies MediaContainer
		if err := xml.Unmarshal(body, &movies); err != nil {
			l.Warnf("Error parsing section %s: %v", section.Title, err)
			continue
		}

		for _, movie := range movies.Videos {
			if movie.ViewCount == 0 {
				continue
			}

			entry := letterboxd.Entry{
				Title: movie.Title,
				Year:  strconv.Itoa(movie.Year),
			}

			if movie.LastViewedAt > 0 {
				entry.WatchedDate = time.Unix(movie.LastViewedAt, 0).Format("2006-01-02")
			}

			if movie.UserRating > 0 {
				entry.Rating10 = strconv.FormatFloat(movie.UserRating, 'f', 0, 64)
			}

			for _, guid := range movie.Guids {
				if strings.HasPrefix(guid.ID, "imdb://") {
					entry.IMDbID = strings.TrimPrefix(guid.ID, "imdb://")
					break
				}
			}

			entries = append(entries, entry)
		}
	}

	l.Infof("Found %d watched movies", len(entries))
	l.Infof("Writing to %s...", *output)

	if err := letterboxd.WriteCSV(*output, entries); err != nil {
		l.Fatal(err)
	}

	fmt.Printf("Successfully exported %d movies to %s\n", len(entries), *output)

	if *importToLetterboxd {
		l.Info("Importing to Letterboxd...")
		importedCount, err := letterboxd.ImportWatchedFilms(*letterboxdUsername, *letterboxdPassword, *output)
		if err != nil {
			l.WithError(err).Fatal("could not import watched films to Letterboxd")
		}
		l.Infof("Import to Letterboxd completed successfully: %d films imported", importedCount)
		if importedCount != int64(len(entries)) {
			l.Warnf("Imported count (%d) does not match exported count (%d)", importedCount, len(entries))
		}
	}
	l.Info("plex2letterboxd2 completed successfully")
}
