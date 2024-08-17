package inat

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

type observation struct {
	ID       int64               `json:"id"`
	UserID   int64               `json:"user_id"`
	Username string              `json:"user_login"`
	Photos   []observationPhotos `json:"photos"`
	Species  *string             `json:"species_guess,omitempty"`
	Taxon    *observationTaxon   `json:"taxon,omitempty"`
}
