package inatapi

// Observations
type Photo struct {
	MediumUrl string `json:"medium_url"`
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
	Species  string           `json:"species_guess"`
	Taxon    observationTaxon `json:"taxon"`
	Username string           `json:"user_login"`
	Photos   []Photo          `json:"photos"`
	ID       int64            `json:"id"`
	UserID   int64            `json:"user_id"`
}

// taxa
type Taxa struct {
	Rank                string `json:"rank"`
	ObservationCount    int64  `json:"observations_count"`
	DefaultPhoto        Photo  `json:"default_photo"`
	PreferredCommonName string `json:"preferred_common_name"`
}

// Search Results
type SearchRecord struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`

	DefaultPhoto Photo `json:"default_photo"`
	// only present when `item.Type` equals "Taxa"
	Taxa
}

type SearchResultItem struct {
	Type    string       `json:"type"`
	Record  SearchRecord `json:"record"`
	Matches []string     `json:"matches"`
}

type SearchResult struct {
	Results      []SearchResultItem `json:"results"`
	TotalResults int                `json:"total_results"`
	Page         int                `json:"page"`
	PerPage      int                `json:"per_page"`
}
