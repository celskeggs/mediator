package atoms

import (
	"fmt"
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/datum"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
	"github.com/celskeggs/mediator/webclient/sprite"
)

type StatContext struct {
	// currentPanel == "" means the default panel, which is titled "Stats", but is not the same as the "Stats" panel
	currentPanel string
	client       types.Value
	display      sprite.StatDisplay
}

func (s *StatContext) renderDatumToStat(datum types.Value) sprite.StatEntry {
	if str, ok := datum.(types.String); ok {
		return sprite.StatEntry{
			Name: string(str),
		}
	} else if integer, ok := datum.(types.Int); ok {
		return sprite.StatEntry{
			Name: fmt.Sprintf("%d", int(integer)),
		}
	} else if types.IsType(datum, "/atom") {
		ok, _, gameSprite := datum.Var("appearance").(Appearance).ToSprite(0, 0, datum.Var("dir").(common.Direction))
		if !ok {
			// make sure that no icon is used
			gameSprite = sprite.GameSprite{}
		}
		verbs := WorldOf(datum.(*types.Datum)).ListVerbsOnAtom(s.client, datum.(*types.Datum))
		return sprite.StatEntry{
			Icon:         gameSprite.Icon,
			SourceX:      gameSprite.SourceX,
			SourceY:      gameSprite.SourceY,
			SourceWidth:  gameSprite.SourceWidth,
			SourceHeight: gameSprite.SourceHeight,
			Name:         types.Unstring(datum.Var("name")),
			Suffix:       types.Unstring(datum.Var("suffix")),
			Verbs:        verbs,
		}
	} else {
		util.FIXME("figure out what we should actually do when given an unknown datum")
		return sprite.StatEntry{
			Name: datum.String(),
		}
	}
}

func (s *StatContext) Stat(name string, value types.Value) {
	if s.display.Panels == nil {
		s.display.Panels = map[string]sprite.StatPanel{}
	}
	panel := s.display.Panels[s.currentPanel]
	if list, ok := value.(datum.List); ok && name == "" {
		for _, element := range datum.Elements(list) {
			if element == nil {
				continue
			}
			panel.Add(s.renderDatumToStat(element))
		}
	} else if value == nil {
		if name != "" {
			panel.Add(sprite.StatEntry{
				Label: name,
			})
		}
	} else {
		stat := s.renderDatumToStat(value)
		stat.Label = name
		panel.Add(stat)
	}
	s.display.Panels[s.currentPanel] = panel
}

func (s *StatContext) StatPanel(panel string) bool {
	if s.display.Panels == nil {
		s.display.Panels = map[string]sprite.StatPanel{}
	}
	s.currentPanel = panel
	// create the entry if it doesn't exist
	s.display.Panels[s.currentPanel] = s.display.Panels[s.currentPanel]
	util.FIXME("actually determine whether a panel is presently displayed")
	return true
}

func (s *StatContext) Display() sprite.StatDisplay {
	return s.display
}
