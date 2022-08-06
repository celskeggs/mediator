package types

import (
	"fmt"
	"github.com/celskeggs/mediator/util"
)

type SrcSettingType int

const (
	SrcSettingTypeUsr SrcSettingType = iota
	SrcSettingTypeUsrContents
	SrcSettingTypeUsrLoc
	SrcSettingTypeUsrGroup
	SrcSettingTypeWorld
	SrcSettingTypeWorldContents
	SrcSettingTypeView
	SrcSettingTypeOView
)

const SrcDistUnspecified int = -1

func (s SrcSettingType) String() string {
	// these need to match the constant name in this package, so that code generation works correctly
	switch s {
	case SrcSettingTypeUsr:
		return "SrcSettingTypeUsr"
	case SrcSettingTypeUsrContents:
		return "SrcSettingTypeUsrContents"
	case SrcSettingTypeUsrLoc:
		return "SrcSettingTypeUsrLoc"
	case SrcSettingTypeUsrGroup:
		return "SrcSettingTypeUsrGroup"
	case SrcSettingTypeWorld:
		return "SrcSettingTypeWorld"
	case SrcSettingTypeWorldContents:
		return "SrcSettingTypeWorldContents"
	case SrcSettingTypeView:
		return "SrcSettingTypeView"
	case SrcSettingTypeOView:
		return "SrcSettingTypeOView"
	default:
		panic("unknown SrcSetting value " + string(int(s)))
	}
}

type SrcSetting struct {
	Type SrcSettingType
	Dist int
	In   bool
}

func (s SrcSetting) IsZero() bool {
	return s == SrcSetting{}
}


type ProcArgumentAs int

const (
	ProcArgumentNone ProcArgumentAs = iota
	ProcArgumentText
	ProcArgumentMessage
	ProcArgumentNum
	ProcArgumentIcon
	ProcArgumentSound
	ProcArgumentFile
	ProcArgumentKey
	ProcArgumentNull
	ProcArgumentMob
	ProcArgumentObj
	ProcArgumentTurf
	ProcArgumentArea
	ProcArgumentAnything
)

func (t ProcArgumentAs) String() string {
	switch t {
	case ProcArgumentNone:
		return "none"
	case ProcArgumentText:
		return "text"
	case ProcArgumentMessage:
		return "message"
	case ProcArgumentNum:
		return "num"
	case ProcArgumentIcon:
		return "icon"
	case ProcArgumentSound:
		return "sound"
	case ProcArgumentFile:
		return "file"
	case ProcArgumentKey:
		return "key"
	case ProcArgumentNull:
		return "null"
	case ProcArgumentMob:
		return "mob"
	case ProcArgumentObj:
		return "obj"
	case ProcArgumentTurf:
		return "turf"
	case ProcArgumentArea:
		return "area"
	case ProcArgumentAnything:
		return "anything"
	default:
		panic(fmt.Sprintf("unexpected proc argument as-type %d", t))
	}
}

func (t ProcArgumentAs) Expression() string {
	return "ProcArgument" + util.ToTitle(t.String())
}

func ProcArgumentFromString(arg string) ProcArgumentAs {
	switch arg {
	case "text":
		return ProcArgumentText
	case "message":
		return ProcArgumentMessage
	case "num":
		return ProcArgumentNum
	case "icon":
		return ProcArgumentIcon
	case "sound":
		return ProcArgumentSound
	case "file":
		return ProcArgumentFile
	case "key":
		return ProcArgumentKey
	case "null":
		return ProcArgumentNull
	case "mob":
		return ProcArgumentMob
	case "obj":
		return ProcArgumentObj
	case "turf":
		return ProcArgumentTurf
	case "area":
		return ProcArgumentArea
	case "anything":
		return ProcArgumentAnything
	default:
		return ProcArgumentNone
	}
}

// really just used for verbs (except background which isn't supported yet)
type ProcSettings struct {
	Src      SrcSetting
	ArgTypes []ProcArgumentAs
}

func (p ProcSettings) IsZero() bool {
	return p.Src == SrcSetting{} && len(p.ArgTypes) == 0
}
