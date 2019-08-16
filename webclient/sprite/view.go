package sprite

type ViewUpdate struct {
	NewState  *SpriteView `json:"newstate"`
	TextLines []string    `json:"textlines"`
	Sounds    []Sound     `json:"sounds"`
}
