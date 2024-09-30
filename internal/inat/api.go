package inat

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Api struct {
}

func New() Api {
	return Api{}
}

func (a Api) Search(sources []string, q string) (SearchResult, error) {
	var sr SearchResult

	res, err := http.Get(
		fmt.Sprintf("https://api.inaturalist.org/v1/search?q=%s&sources=%s",
			url.QueryEscape(q), strings.Join(sources, ",")),
	)

	if err != nil {
		return sr, fmt.Errorf("http error: %w", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return sr, fmt.Errorf("error parsing body: %w", err)
	}

	err = json.Unmarshal(body, &sr)

	if err != nil {
		return sr, fmt.Errorf("error parsing json: %w", err)
	}

	return sr, nil
}

func (a Api) FetchRecentProjectObservations(
	projectID int64,
	pages int,
	pageSize int,
) ([]Observation, error) {
	var (
		observations []Observation
		currentPage  = 0
	)

	for currentPage < pages {
		log.Printf("Connecting to inat for page %d", currentPage)
		res, err := http.Get(
			fmt.Sprintf(
				"https://inaturalist.org/observations/project/%d.json?order_by=id&order=desc&per_page=%d",
				projectID,
				pageSize,
			),
		)

		if err != nil {
			return observations, fmt.Errorf("http error: %w", err)
		}

		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)

		if err != nil {
			return observations, fmt.Errorf("error parsing body: %w", err)
		}

		var items []Observation
		err = json.Unmarshal(body, &items)

		if err != nil {
			return observations, fmt.Errorf("error parsing json: %w", err)
		}

		for _, item := range items {
			if len(item.Photos) > 0 {
				if item.Photos[0].MediumURL != "" {
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
