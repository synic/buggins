package inatapi

// Observations
type Photo struct {
	MediumURL string `json:"medium_url"`
}

type observationTaxonName struct {
	Name string `json:"name"`
}

type observationTaxon struct {
	Name        string               `json:"name"`
	CommonName  observationTaxonName `json:"common_name"`
	DefaultName observationTaxonName `json:"default_name"`
}

type observationUser struct {
	Username    string `json:"login"`
	Name        string `json:"string"`
	UserIconURL string `json:"user_icon_url"`
	ID          int64  `json:"id"`
}

type Observation struct {
	Taxon      observationTaxon `json:"taxon"`
	Species    string           `json:"species_guess"`
	ObservedOn string           `json:"observed_on"`
	Username   string           `json:"user_login"`
	User       observationUser  `json:"user"`
	Photos     []Photo          `json:"photos"`
	ID         int64            `json:"id"`
	UserID     int64            `json:"user_id"`
}

func (o Observation) GetTaxonNames() (string, string) {
	taxonName := "unknown"
	commonName := "unknown"
	taxon := o.Taxon

	if taxon.Name != "" {
		taxonName = taxon.Name

		if taxon.CommonName.Name != "" {
			commonName = taxon.CommonName.Name
		} else if taxon.DefaultName.Name != "" {
			commonName = taxon.DefaultName.Name
		} else if o.Species != "" {
			commonName = o.Species
		}
	}

	return taxonName, commonName
}

// taxa
type Taxa struct {
	Rank                string `json:"rank"`
	DefaultPhoto        Photo  `json:"default_photo"`
	PreferredCommonName string `json:"preferred_common_name"`
	ObservationCount    int64  `json:"observations_count"`
}

// Search Results
type SearchRecord struct {
	Name         string `json:"name"`
	DefaultPhoto Photo  `json:"default_photo"`
	Taxa
	ID int64 `json:"id"`
}

type SearchResultItem struct {
	Type    string       `json:"type"`
	Matches []string     `json:"matches"`
	Record  SearchRecord `json:"record"`
}

type SearchResult struct {
	Results      []SearchResultItem `json:"results"`
	TotalResults int                `json:"total_results"`
	Page         int                `json:"page"`
	PerPage      int                `json:"per_page"`
}
