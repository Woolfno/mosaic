package response

type RandomPageTitle struct {
	Items []map[string]interface{} `json:"items"`
}

type MediaList struct {
	Items []Item `json:"items"`
}

type Item struct {
	Type   string   `json:"type"`
	SrcSet []SrcSet `json:"srcset"`
}

type SrcSet struct {
	Src   string `json:"src"`
	Scale string `json:"scale"`
}
