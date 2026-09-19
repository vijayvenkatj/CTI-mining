package resources

type Pulse struct {
	ID                string      `json:"id"`
	Name              string      `json:"name"`
	Description       string      `json:"description"`
	AuthorName        string      `json:"author_name"`
	Modified          string      `json:"modified"`
	Created           string      `json:"created"`
	Revision          int         `json:"revision"`
	TLP               string      `json:"tlp"`
	Public            int         `json:"public"`
	Adversary         string      `json:"adversary"`
	Indicators        []Indicator `json:"indicators"`
	Tags              []string    `json:"tags"`
	TargetedCountries []string    `json:"targeted_countries"`
	MalwareFamilies   []string    `json:"malware_families"`
	AttackIDs         []string    `json:"attack_ids"`
	References        []string    `json:"references"`
	Industries        []string    `json:"industries"`
	ExtractSource     []string    `json:"extract_source"`
	MoreIndicators    bool        `json:"more_indicators"`
}
