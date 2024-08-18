package inat

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func fetchRecentProjectObservations(
	projectId string,
	pages int,
	pageSize int,
) ([]observation, error) {
	var (
		observations []observation
		currentPage  = 0
	)

	for currentPage < pages {
		log.Printf("Connecting to inat for page %d", currentPage)
		resp, err := http.Get(
			fmt.Sprintf(
				"https://inaturalist.org/observations/project/%s.json?order_by=id&order=desc&per_page=%d",
				projectId,
				pageSize,
			),
		)

		if err != nil {
			return observations, fmt.Errorf("http error: %w", err)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return observations, fmt.Errorf("error parsing body: %w", err)
		}

		var items []observation
		err = json.Unmarshal(body, &items)

		if err != nil {
			return observations, fmt.Errorf("error parsing json: %w", err)
		}

		for _, item := range items {
			if item.Photos != nil && len(item.Photos) > 0 {
				if item.Photos[0].LargeUrl != "" {
					observations = append(observations, item)
				}
			}
		}

		if len(items) < pageSize {
			break
		}

		currentPage += 1
	}

	log.Printf("%d recent observations to examine", len(observations))

	return observations, nil
}
