package sprite

import (
	"fmt"
	"github.com/celskeggs/mediator/platform/types"
)

type Sound struct {
	File    string `json:"file"`
	Repeat  bool   `json:"repeat"`
	Wait    bool   `json:"wait"`
	Channel uint   `json:"channel"`
	Volume  uint   `json:"volume"`
}

var _ types.Value = &Sound{}

func (s Sound) Reference() *types.Ref {
	return &types.Ref{s}
}

func (s Sound) Var(name string) types.Value {
	switch name {
	case "file":
		return types.String(s.File)
	case "repeat":
		return types.Bool(s.Repeat)
	case "wait":
		return types.Bool(s.Wait)
	case "channel":
		return types.Int(s.Channel)
	case "volume":
		return types.Int(s.Volume)
	default:
		panic("no such field " + name + " on /sound")
	}
}

func (s Sound) SetVar(name string, value types.Value) {
	switch name {
	case "file":
		s.File = types.Unstring(value)
	case "repeat":
		s.Repeat = types.Unbool(value)
	case "wait":
		s.Wait = types.Unbool(value)
	case "channel":
		s.Channel = uint(types.Unint(value))
	case "volume":
		s.Volume = uint(types.Unint(value))
	default:
		panic("no such field " + name + " on /sound")
	}
}

func (s Sound) Invoke(name string, parameters ...types.Value) types.Value {
	panic("no such proc " + name + " on /sound")
}

func (s Sound) String() string {
	return fmt.Sprintf("[sound: %q]", s.File)
}
