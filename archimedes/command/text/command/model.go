package command

type RhemaFile []WorkChunk

type WorkChunk struct {
	Author          string  `json:"author"`
	Book            string  `json:"book"`
	Type            string  `json:"type"`
	Reference       string  `json:"reference"`
	PerseusTextLink string  `json:"perseusTextLink,omitempty"`
	Rhemai          []Rhema `json:"rhemai"`
}

type Rhema struct {
	Greek        string   `json:"greek"`
	Translations []string `json:"translations"`
	Section      string   `json:"section"`
}
