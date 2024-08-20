package inatapi

// Observations
type observationPhotos struct {
	LargeUrl string `json:"large_url"`
}

type observationTaxonName struct {
	Name string `json:"name"`
}

type observationTaxon struct {
	Name        string               `json:"name"`
	CommonName  observationTaxonName `json:"common_name"`
	DefaultName observationTaxonName `json:"default_name"`
}

type Observation struct {
	Species  string              `json:"species_guess"`
	Taxon    observationTaxon    `json:"taxon"`
	Username string              `json:"user_login"`
	Photos   []observationPhotos `json:"photos"`
	ID       int64               `json:"id"`
	UserID   int64               `json:"user_id"`
}

// Search Results
type searchResultItemRecordPhoto struct {
	Url       string `json:"url"`
	MediumUrl string `json:"medium_url"`
	SquareUrl string `json:"square_url"`
}

type searchResultItemRecordTaxaData struct {
	Rank                string `json:"rank"`
	ObservationCount    int64  `json:"observations_count"`
	PreferredCommonName string `json:"preferred_common_name"`
}

type searchResultItemRecord struct {
	DefaultPhoto searchResultItemRecordPhoto `json:"default_photo"`
	ID           int64                       `json:"id"`
	Name         string                      `json:"name"`

	// only present when `item.Type` equals "Taxa"
	searchResultItemRecordTaxaData
}

type SearchResultItem struct {
	Type    string                 `json:"type"`
	Record  searchResultItemRecord `json:"record"`
	Matches []string               `json:"matches"`
}

type SearchResult struct {
	Results      []SearchResultItem `json:"results"`
	TotalResults int                `json:"total_results"`
	Page         int                `json:"page"`
	PerPage      int                `json:"per_page"`
}
