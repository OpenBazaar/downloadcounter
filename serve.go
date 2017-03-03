package downloadcounter

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"database/sql"

	"github.com/OpenBazaar/downloadcounter/connect"
	"github.com/gocraft/health"
)

var (
	selectStmt    *sql.Stmt
	currentResult resultCache
	csvHeaders    = []string{"Version", "Count", "As Of"}
)

type selectResultItem struct {
	Repo    string
	Version string
	Count   int
	AsOf    string
}

type resultCache struct {
	createdAt time.Time
	data      []selectResultItem
}

// Serve serves up the latest count data
func Serve(conns *connect.Connections) error {
	http.HandleFunc("/", newCountHandler(conns))
	log.Fatal(http.ListenAndServe(":8080", nil))
	return nil
}

func newCountHandler(conns *connect.Connections) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		job := conns.Stream.NewJob("request")

		// Check if cache is out of date and if it is update it
		if currentResult.createdAt.Add(1 * time.Hour).Before(time.Now()) {
			job.Event("updating_cache")

			results, err := getLatestCounts(conns)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("error\n"))
				conns.Stream.EventErr("get_lat_counts", err)
				job.Complete(health.Error)
			}

			currentResult.data = results
			currentResult.createdAt = time.Now()
		}

		// Output data
		writeCSV(w, currentResult.data)

		job.Complete(health.Success)
	}
}

func getLatestCounts(conns *connect.Connections) ([]selectResultItem, error) {
	// Create stmt if it hasn't been created yet
	if selectStmt == nil {
		var err error
		selectStmt, err = conns.DB.Prepare(`
				SELECT repo, version, sum(count) AS count, date(created_at) AS as_of
				FROM download_counts
				WHERE date(created_at) = (select MAX(date(created_at)))
				GROUP BY repo, as_of, version;
			`)
		if err != nil {
			return nil, err
		}
	}

	// Get updated data
	rows, err := selectStmt.Query()
	if err != nil {
		return nil, err
	}

	results := []selectResultItem{}
	for rows.Next() {
		result := selectResultItem{}

		err = rows.Scan(
			&result.Repo,
			&result.Version,
			&result.Count,
			&result.AsOf,
		)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

func writeCSV(w io.Writer, counts []selectResultItem) error {
	_, err := w.Write([]byte(newCSV(counts)))
	return err
}

func newCSV(counts []selectResultItem) string {
	csv := strings.Join(csvHeaders, ", ")

	for _, count := range counts {
		csv += fmt.Sprintf("\n%s,%d,%s", count.Version, count.Count, count.AsOf)
	}

	return csv
}
