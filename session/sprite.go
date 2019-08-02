package session

type GameSprite struct {
	Icon string `json:"icon"`
	SourceX uint `json:"sx"`
	SourceY uint `json:"sy"`
	SourceWidth uint `json:"sw"`
	SourceHeight uint `json:"sh"`
	X uint `json:"x"`
	Y uint `json:"y"`
	Width uint `json:"w"`
	Height uint `json:"h"`
}

type SpriteView struct {
	Sprites map[string]GameSprite `json:"sprites"`
}

func (a SpriteView) Equal(b SpriteView) bool {
	if len(a.Sprites) != len(b.Sprites) {
		return false
	}
	for k, v := range a.Sprites {
		v2, found := b.Sprites[k]
		if !found {
			return false
		}
		if v != v2 {
			return false
		}
	}
	return true
}
