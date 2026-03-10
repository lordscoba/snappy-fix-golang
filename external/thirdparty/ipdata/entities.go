package ipdata

// Docs: https://ipdata.co/docs.html
type RawResponse struct {
	IP            string  `json:"ip"`
	Type          string  `json:"type"`
	ContinentCode string  `json:"continent_code"`
	ContinentName string  `json:"continent_name"`
	CountryCode   string  `json:"country_code"`
	CountryName   string  `json:"country_name"`
	Region        string  `json:"region"`      // maps to RegionName
	RegionCode    string  `json:"region_code"` // maps to RegionCode
	City          string  `json:"city"`
	Postal        string  `json:"postal"` // maps to Zip
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
}
