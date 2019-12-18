package types

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

// really just used for verbs (except background which isn't supported yet)
type ProcSettings struct {
	Src SrcSetting
}

func (p ProcSettings) IsZero() bool {
	return p == ProcSettings{}
}
