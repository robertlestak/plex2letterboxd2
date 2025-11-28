package main

import (
	"flag"
	"os"

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

func main() {
	l := log.WithFields(log.Fields{
		"component": "lup",
	})
	l.Info("lup called")
	username := flag.String("username", os.Getenv("LETTERBOXD_USERNAME"), "Letterboxd username")
	password := flag.String("password", os.Getenv("LETTERBOXD_PASSWORD"), "Letterboxd password")
	importFile := flag.String("import-file", "letterboxd.csv", "Path to Letterboxd import CSV file")
	initPlaywright := flag.Bool("init", false, "Initialize Playwright")
	flag.Parse()

	if *initPlaywright {
		if err := letterboxd.InitPlaywright(); err != nil {
			l.WithError(err).Fatal("could not initialize Playwright")
		}
		l.Info("Playwright initialized successfully")
		return
	}

	if *username == "" || *password == "" {
		l.Fatal("LETTERBOXD_USERNAME and LETTERBOXD_PASSWORD are required")
	}

	if _, err := letterboxd.ImportWatchedFilms(*username, *password, *importFile); err != nil {
		l.WithError(err).Fatal("could not import watched films")
	}
}
