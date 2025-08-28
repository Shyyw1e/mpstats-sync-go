package mpstats

type VersionsResp []struct {
	Version string `json:"version"`
}

type FullPage struct {
	FullName    string        `json:"full_name"`
	Color       string        `json:"color"`
	Description string        `json:"description"`
	FullText    string        `json:"full_text"`
	ParamNames  []string      `json:"param_names"`
	ParamValues []interface{} `json:"param_values"`
}
