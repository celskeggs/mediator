package sprite

import "github.com/celskeggs/mediator/platform/icon"

type GameSprite struct {
	Icon         string          `json:"icon"`
	Frames       []icon.SourceXY `json:"frames"`
	SourceWidth  uint            `json:"sw"`
	SourceHeight uint            `json:"sh"`
	X            uint            `json:"x"`
	Y            uint            `json:"y"`
	Width        uint            `json:"w"`
	Height       uint            `json:"h"`
	Name         string          `json:"name"`
	Verbs        []string        `json:"verbs"`
	UID          uint64          `json:"uid"`
}

func (s GameSprite) Equal(o GameSprite) bool {
	if !(s.Icon == o.Icon &&
		s.SourceWidth == o.SourceWidth && s.SourceHeight == o.SourceHeight &&
		s.X == o.X && s.Y == o.Y && s.Width == o.Width && s.Height == o.Height &&
		s.Name == o.Name && s.UID == o.UID) {
		return false
	}
	if len(s.Frames) != len(o.Frames) {
		return false
	}
	for i, v := range s.Frames {
		if o.Frames[i] != v {
			return false
		}
	}
	if len(s.Verbs) != len(o.Verbs) {
		return false
	}
	for i, v := range s.Verbs {
		if o.Verbs[i] != v {
			return false
		}
	}
	return true
}

type Flick struct {
	Icon         string          `json:"icon"`
	Frames       []icon.SourceXY `json:"frames"`
	SourceWidth  uint            `json:"sw"`
	SourceHeight uint            `json:"sh"`
	UID          uint64          `json:"uid"`
}

type StatEntry struct {
	Label        string          `json:"label"`
	Icon         string          `json:"icon"`
	Frames       []icon.SourceXY `json:"frames"`
	SourceWidth  uint            `json:"sw"`
	SourceHeight uint            `json:"sh"`
	Name         string          `json:"name"`
	Suffix       string          `json:"suffix"`
	Verbs        []string        `json:"verbs"`
	UID          uint64          `json:"uid"`
}

func (s StatEntry) Equal(o StatEntry) bool {
	if !(s.Label == o.Label && s.Icon == o.Icon &&
		s.SourceWidth == o.SourceWidth && s.SourceHeight == o.SourceHeight &&
		s.Name == o.Name && s.Suffix == o.Suffix && s.UID == o.UID) {
		return false
	}
	if len(s.Frames) != len(o.Frames) {
		return false
	}
	for i, v := range s.Frames {
		if o.Frames[i] != v {
			return false
		}
	}
	if len(s.Verbs) != len(o.Verbs) {
		return false
	}
	for i, v := range s.Verbs {
		if o.Verbs[i] != v {
			return false
		}
	}
	return true
}

type StatPanel struct {
	Entries []StatEntry `json:"entries"`
}

func (p *StatPanel) indexOfLabel(label string) int {
	for i, entry := range p.Entries {
		if entry.Label == label {
			return i
		}
	}
	return -1
}

func (p *StatPanel) Add(entry StatEntry) {
	if entry.Label != "" {
		if index := p.indexOfLabel(entry.Label); index != -1 {
			p.Entries[index] = entry
			return
		}
	}
	p.Entries = append(p.Entries, entry)
}

func (p StatPanel) Equal(o StatPanel) bool {
	if len(p.Entries) != len(o.Entries) {
		return false
	}
	for i, ent := range p.Entries {
		if !o.Entries[i].Equal(ent) {
			return false
		}
	}
	return true
}

type StatDisplay struct {
	Panels map[string]StatPanel `json:"panels"`
}

func (d StatDisplay) Equal(o StatDisplay) bool {
	if d.Panels == nil {
		return o.Panels == nil
	}
	if len(d.Panels) != len(o.Panels) {
		return false
	}
	for name, panel := range d.Panels {
		if ent, ok := o.Panels[name]; !ok {
			// nonexistent, so the keys are different
			return false
		} else if !panel.Equal(ent) {
			return false
		}
	}
	return true
}

type SpriteView struct {
	WindowTitle    string       `json:"windowtitle"`
	ViewPortWidth  uint         `json:"viewportwidth"`
	ViewPortHeight uint         `json:"viewportheight"`
	Sprites        []GameSprite `json:"sprites"`
	Stats          StatDisplay  `json:"stats"`
	Verbs          []string     `json:"verbs"`
}

func (a SpriteView) Equal(b SpriteView) bool {
	if a.ViewPortWidth != b.ViewPortWidth || a.ViewPortHeight != b.ViewPortHeight {
		return false
	}
	if !a.Stats.Equal(b.Stats) {
		return false
	}
	if len(a.Verbs) != len(b.Verbs) {
		return false
	}
	for i, verb := range a.Verbs {
		if b.Verbs[i] != verb {
			return false
		}
	}
	if len(a.Sprites) != len(b.Sprites) {
		return false
	}
	used := make([]bool, len(a.Sprites))
	for _, sprite := range b.Sprites {
		found := false
		for i, cmp := range a.Sprites {
			if !used[i] && sprite.Equal(cmp) {
				used[i] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	// at this point, we know that:
	//  - A and B have the same number of elements
	//  - every element in B is also in A
	// so we can conclude that A and B have the same elements
	return true
}
