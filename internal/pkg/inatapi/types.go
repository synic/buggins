package inatapi

// Observations
type observationPhotos struct {
	LargeUrl string `json:"large_url"`
}

type observationTaxonName struct {
	Name *string `json:"name,omitempty"`
}

type observationTaxon struct {
	Name        *string               `json:"name,omitempty"`
	CommonName  *observationTaxonName `json:"common_name,omitempty"`
	DefaultName *observationTaxonName `json:"default_name,omitempty"`
}

type Observation struct {
	Species  *string             `json:"species_guess,omitempty"`
	Taxon    *observationTaxon   `json:"taxon,omitempty"`
	Username string              `json:"user_login"`
	Photos   []observationPhotos `json:"photos"`
	ID       int64               `json:"id"`
	UserID   int64               `json:"user_id"`
}

// Search Results
type searchResultItemRecordPhoto struct {
	MediumUrl string `json:"medium_url"`
}

type searchResultItemRecord struct {
	DefaultPhoto        *searchResultItemRecordPhoto `json:"default_photo,omitempty"`
	ID                  int64                        `json:"id"`
	Name                string                       `json:"name"`
	Rank                string                       `json:"rank"`
	ObservationCount    int64                        `json:"observations_count"`
	PreferredCommonName string                       `json:"preferred_common_name"`
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
