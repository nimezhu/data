package data

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

type gAgent struct {
	root string
}

func (agent gAgent) getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := agent.tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

func (agent gAgent) tokenCacheFile() (string, error) {
	tokenCacheDir := filepath.Join(agent.root, "credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("gsheets.json")), nil
}
