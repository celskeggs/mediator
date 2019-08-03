package dmi

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type DMIState struct {
	Directions int
	Frames     int
	Rewind     bool
	Delay      []int
}

type DMIInfo struct {
	Width  int
	Height int
	States map[string]DMIState
}

func getBodyLines(whole string) ([]string, error) {
	lines := strings.Split(strings.TrimSpace(whole), "\n")
	if lines[0] != "# BEGIN DMI" || lines[len(lines)-1] != "# END DMI" {
		return nil, errors.New("did not find expected # BEGIN DMI and # END DMI in Description section")
	}
	return lines[1 : len(lines)-1], nil
}

func getKVSections(whole string) (sections []map[string]string, err error) {
	var section map[string]string
	lines, err := getBodyLines(whole)
	if err != nil {
		return nil, err
	}
	for _, line := range lines {
		if line[0] != '\t' {
			// no tab: start a section
			section = make(map[string]string)
			sections = append(sections, section)
		} else {
			// tab: continue a section
			if section == nil {
				return nil, errors.New("expected first line to be unindented")
			}
			line = line[1:]
		}
		kv := strings.SplitN(line, " = ", 2)
		if len(kv) < 2 {
			return nil, errors.New("expected ' = ' in line")
		}
		key, value := kv[0], kv[1]
		_, alreadyExists := section[key]
		if alreadyExists {
			return nil, errors.New("duplicate key")
		}
		section[key] = value
	}
	return sections, nil
}

func parseHeader(section map[string]string) (width int, height int, err error) {
	if section["version"] != "4.0" {
		return 0, 0, errors.New("expected version = 4.0 in DMI")
	}
	if section["width"] == "" || section["height"] == "" || len(section) != 3 {
		return 0, 0, errors.New("expected exactly three keys in header: version, width, height")
	}
	widthU, err := strconv.ParseUint(section["width"], 10, 31)
	if err != nil {
		return 0, 0, err
	}
	heightU, err := strconv.ParseUint(section["height"], 10, 31)
	if err != nil {
		return 0, 0, err
	}
	return int(widthU), int(heightU), nil
}

func parseString(value string) (string, error) {
	if !strings.HasPrefix(value, "\"") || !strings.HasSuffix(value, "\"") {
		return "", errors.New("invalid string format")
	}
	// TODO: escaping, maybe?
	return value[1 : len(value)-1], nil
}

func parseState(section map[string]string) (string, DMIState, error) {
	for _, field := range []string{"state", "dirs", "frames"} {
		if section[field] == "" {
			return "", DMIState{}, fmt.Errorf("no '%s' field specified", field)
		}
	}
	for field := range section {
		if field != "state" && field != "dirs" && field != "frames" && field != "rewind" && field != "delay" {
			return "", DMIState{}, fmt.Errorf("unexpected field '%s' found", field)
		}
	}
	stateName, err := parseString(section["state"])
	if err != nil {
		return "", DMIState{}, err
	}

	dirs, err := strconv.ParseUint(section["dirs"], 10, 31)
	if err != nil {
		return "", DMIState{}, err
	}
	frames, err := strconv.ParseUint(section["frames"], 10, 31)
	if err != nil {
		return "", DMIState{}, err
	}
	// TODO: is this the correct default value?
	rewind := false
	if section["rewind"] != "" {
		if section["rewind"] != "0" && section["rewind"] != "1" {
			return "", DMIState{}, errors.New("expected 'rewind' to be 0 or 1")
		}
		rewind = section["rewind"] != "0"
	}

	var delay []int
	if delayValue, found := section["delay"]; found {
		for _, str := range strings.Split(delayValue, ",") {
			delayU, err := strconv.ParseUint(str, 10, 31)
			if err != nil {
				return "", DMIState{}, err
			}
			delay = append(delay, int(delayU))
		}
	}

	return stateName, DMIState{
		Directions: int(dirs),
		Frames:     int(frames),
		Rewind:     rewind,
		Delay:      delay,
	}, nil
}

func ParseDMI(png []byte) (*DMIInfo, error) {
	description, err := parseDescription(png)
	if err != nil {
		return nil, err
	}
	sections, err := getKVSections(description)
	if err != nil {
		return nil, err
	}
	if len(sections) < 1 {
		return nil, errors.New("no header section")
	}
	width, height, err := parseHeader(sections[0])
	if err != nil {
		return nil, err
	}
	states := map[string]DMIState{}
	for i := 1; i < len(sections); i++ {
		key, state, err := parseState(sections[i])
		if err != nil {
			return nil, err
		}
		if _, alreadyExists := states[key]; alreadyExists {
			return nil, errors.New("duplicate state")
		}
		states[key] = state
	}
	return &DMIInfo{
		Width:  width,
		Height: height,
		States: states,
	}, nil
}
