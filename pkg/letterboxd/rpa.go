package letterboxd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
	log "github.com/sirupsen/logrus"
)

func InitPlaywright() error {
	l := log.WithFields(log.Fields{
		"component": "letterboxd",
		"fn":        "InitPlaywright",
	})
	l.Info("InitPlaywright called")
	if err := playwright.Install(); err != nil {
		l.WithError(err).Error("could not install playwright")
		return err
	}
	return nil
}

func ImportWatchedFilms(username, password, importFile string) (int64, error) {
	l := log.WithFields(log.Fields{
		"component":  "letterboxd",
		"fn":         "ImportWatchedFilms",
		"username":   username,
		"importFile": importFile,
	})
	l.Info("ImportWatchedFilms called")
	pw, err := playwright.Run()
	if err != nil {
		l.WithError(err).Error("could not start playwright")
		return 0, err
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(os.Getenv("HEADLESS") != "false"),
	})
	if err != nil {
		l.WithError(err).Error("could not launch browser")
		return 0, err
	}
	page, err := browser.NewPage()
	if err != nil {
		l.WithError(err).Error("could not create page")
		return 0, err
	}
	if _, err = page.Goto("https://letterboxd.com/sign-in/"); err != nil {
		l.WithError(err).Error("could not go to letterboxd sign-in page")
		return 0, err
	}
	// fill in #field-username and #field-password and then click .formactions button.standalone-flow-button
	if err := page.Locator("#field-username").First().Fill(username); err != nil {
		l.WithError(err).Error("could not fill in username")
		return 0, err
	}
	if err := page.Locator("#field-password").First().Fill(password); err != nil {
		l.WithError(err).Error("could not fill in password")
		return 0, err
	}
	if err := page.Locator(".formactions button.standalone-flow-button").First().Click(); err != nil {
		l.WithError(err).Error("could not click sign-in button")
		return 0, err
	}
	time.Sleep(5 * time.Second) // wait for login to complete
	// now go to https://letterboxd.com/import/
	if _, err = page.Goto("https://letterboxd.com/import/"); err != nil {
		l.WithError(err).Error("could not go to letterboxd import page")
		return 0, err
	}
	// set the file input to importFile
	if err := page.Locator("input#upload-imdb-import").First().SetInputFiles(importFile); err != nil {
		l.WithError(err).Error("could not set input file")
		return 0, err
	}
	time.Sleep(20 * time.Second) // wait for file to upload and process
	// get the content of ul#import-films li.import-film p.import-original
	contentList, err := page.Locator("ul#import-films").Locator("li.import-film").Locator("p.import-original").AllTextContents()
	if err != nil {
		l.WithError(err).Error("could not get import film original titles")
		return 0, err
	}
	for _, c := range contentList {
		l.Infof("Import film: %s", strings.TrimSpace(c))
	}
	// click the a.submit-matched-films link
	if err := page.Locator("a.submit-matched-films").First().Click(); err != nil {
		l.WithError(err).Error("could not click submit matched films")
		return 0, err
	}
	time.Sleep(20 * time.Second) // wait for import to complete
	// get the content of .import-progress
	content, err := page.Locator(".import-progress").First().TextContent()
	if err != nil {
		l.WithError(err).Error("could not get import progress content")
		return 0, err
	}
	l.Infof("Import progress: %s", content)
	// content in the form of: "Saved %d films" - extract the number
	var importedCount int64
	_, err = fmt.Sscanf(content, "Saved %d films", &importedCount)
	if err != nil {
		l.WithError(err).Error("could not parse imported film count")
		return 0, err
	}
	l.Infof("Imported %d films to Letterboxd", importedCount)
	if err := browser.Close(); err != nil {
		l.WithError(err).Error("could not close browser")
		return 0, err
	}
	if err := pw.Stop(); err != nil {
		l.WithError(err).Error("could not stop playwright")
		return 0, err
	}
	l.Info("ImportWatchedFilms completed successfully")
	return importedCount, nil
}
