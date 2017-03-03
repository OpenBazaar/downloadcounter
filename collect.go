package downloadcounter

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"strings"

	"log"

	"github.com/OpenBazaar/downloadcounter/connect"
)

const endpointTemplate = "https://api.github.com/repos/%s/releases"

var (
	insertStmt *sql.Stmt

	httpClient = &http.Client{
		Timeout: time.Second * 10,
	}
)

// version -> count
type downloadCounts map[string]uint

type githubReleases struct {
	Repo   string
	Counts []downloadCounts
}

// http response structs
type githubReleaseAsset struct {
	Count uint `json:"download_count"`
}

type githubRelease struct {
	Name   string `json:"name"`
	Assets []githubReleaseAsset
}

// Collect gets the latest download counts and stores them
func Collect(conns *connect.Connections) error {
	err := initDB(conns.DB, conns.Config.Table)
	if err != nil {
		return err
	}

	counts, err := getRepoCounts(conns.Config.Repo)
	if err != nil {
		return err
	}

	log.Println("Counts:", counts)

	err = saveCounts(conns, counts)
	if err != nil {
		return err
	}

	fmt.Println("collecting")

	return nil
}

func initDB(db *sql.DB, table string) error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
     id INT NOT NULL AUTO_INCREMENT,
     created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
     count INT NOT NULL,
		 repo VARCHAR(255) NOT NULL,
		 version VARCHAR(255) NOT NULL,
		 PRIMARY KEY (id)
	);`, table)

	_, err := db.Exec(query)
	return err
}

func getRepoCounts(repo string) (downloadCounts, error) {
	// Get all release data
	resp, err := httpClient.Get(fmt.Sprintf(endpointTemplate, repo))
	if err != nil {
		return nil, err
	}

	// Read response into a byte slce
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	// Unmarshal
	releases := []githubRelease{}
	err = json.Unmarshal(body, &releases)
	if err != nil {
		return nil, err
	}

	// Create a map of version -> count where count is the sum of downloads
	// for all assets in the release
	counts := downloadCounts{}
	for _, release := range releases {
		// Skip test releases
		if strings.Contains(strings.ToLower(release.Name), "test") {
			continue
		}

		for _, asset := range release.Assets {
			counts[release.Name] += asset.Count
		}
	}

	return counts, err
}

// saveCounts inserts each count into the table
func saveCounts(conns *connect.Connections, counts downloadCounts) error {
	if insertStmt == nil {
		var err error
		insertStmt, err = conns.DB.Prepare(fmt.Sprintf("INSERT INTO %s (repo, version, count) VALUES (?,?,?)", conns.Config.Table))
		if err != nil {
			return err
		}
	}

	for version, count := range counts {
		_, err := insertStmt.Exec(conns.Config.Repo, version, count)
		if err != nil {
			return err
		}
	}

	return nil
}
