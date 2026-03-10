package ipapi

// ip-api.com response subset we need.
// Docs: https://ip-api.com/docs/api:json
type RawResponse struct {
	Status        string  `json:"status"`        // "success" | "fail"
	Message       string  `json:"message"`       // present when status=="fail"
	Query         string  `json:"query"`         // the IP you asked about
	Continent     string  `json:"continent"`     // e.g., "Africa"         (may be empty on free)
	ContinentCode string  `json:"continentCode"` // e.g., "AF"             (may be empty on free)
	Country       string  `json:"country"`
	CountryCode   string  `json:"countryCode"`
	Region        string  `json:"region"`     // region code, e.g., "LA"
	RegionName    string  `json:"regionName"` // e.g., "Lagos"
	City          string  `json:"city"`
	Zip           string  `json:"zip"`
	Lat           float64 `json:"lat"`
	Lon           float64 `json:"lon"`
}
