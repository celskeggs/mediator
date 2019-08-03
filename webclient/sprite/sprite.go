package sprite

type GameSprite struct {
	Icon         string `json:"icon"`
	SourceX      uint   `json:"sx"`
	SourceY      uint   `json:"sy"`
	SourceWidth  uint   `json:"sw"`
	SourceHeight uint   `json:"sh"`
	X            uint   `json:"x"`
	Y            uint   `json:"y"`
	Width        uint   `json:"w"`
	Height       uint   `json:"h"`
}

type SpriteView struct {
	Sprites []GameSprite `json:"sprites"`
}

func (a SpriteView) Equal(b SpriteView) bool {
	if len(a.Sprites) != len(b.Sprites) {
		return false
	}
	spritemap := map[GameSprite]uint{}
	for _, sprite := range a.Sprites {
		spritemap[sprite] += 1
	}
	for _, sprite := range b.Sprites {
		if spritemap[sprite] == 0 {
			return false
		}
		spritemap[sprite] -= 1
	}
	// at this point, we know that:
	//  - A and B have the same number of elements
	//  - every element in B is also in A
	// so we can conclude that A and B have the same elements
	return true
}
