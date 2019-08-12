package parsemap

import (
	"errors"
	"fmt"
	"strings"
)

type Location struct {
	X uint32
	Y uint32
	Z uint32
}

func NewLocationI(x int, y int, z int) Location {
	return Location{
		X: uint32(x),
		Y: uint32(y),
		Z: uint32(z),
	}
}

type RawMap struct {
	PathSets map[string][]string
	Content  [][][]string
}

type MapObserver interface {
	SetSize(size Location)
	AddAtom(xyz Location, path string) error
}

func readPathSets(header string) (map[string][]string, error) {
	lines := strings.Split(strings.TrimSpace(header), "\n")
	out := map[string][]string{}
	for _, line := range lines {
		if line[0] != '"' {
			return nil, errors.New("expected \" at start of pathset line")
		}
		if line[len(line)-1] != ')' {
			return nil, errors.New("expected ) at end of pathset line")
		}
		parts := strings.Split(line[1:len(line)-1], "\" = (")
		if len(parts) != 2 {
			return nil, errors.New("formatting broken: expected two fields")
		}
		abbreviation := parts[0]
		paths := strings.Split(parts[1], ",")
		_, prev := out[abbreviation]
		if prev {
			return nil, fmt.Errorf("unexpected duplicate abbreviation: %s", abbreviation)
		}
		out[abbreviation] = paths
	}
	return out, nil
}

// row-column to x-y-z
func convertRCto3D(rc [][]string) [][][]string {
	xyz := make([][][]string, len(rc[0]))
	for coli := 0; coli < len(rc[0]); coli++ {
		xyz[coli] = make([][]string, len(rc))
	}
	for rowi, row := range rc {
		for coli, col := range row {
			xyz[coli][len(rc)-rowi-1] = []string{col}
		}
	}
	return xyz
}

func readContent(body string) ([][][]string, error) {
	lines := strings.Split(strings.TrimSpace(body), "\n")
	if lines[0] != "(1,1,1) = {\"" {
		return nil, fmt.Errorf("could not understand first worldmap line")
	}
	if lines[len(lines)-1] != "\"}" {
		return nil, fmt.Errorf("could not understand last worldmap line")
	}
	rowColumn := make([][]string, len(lines)-2)
	for rowi, row := range lines[1: len(lines)-1] {
		rowColumn[rowi] = make([]string, len(row))
		for coli := 0; coli < len(row); coli++ {
			rowColumn[rowi][coli] = row[coli: coli+1]
		}
	}
	return convertRCto3D(rowColumn), nil
}

func readRawMap(text string) (*RawMap, error) {
	parts := strings.Split(strings.TrimSpace(text), "\n\n")
	if len(parts) != 2 {
		return nil, fmt.Errorf("worldmap format support incomplete: more than one double newline: %v", parts)
	}
	pathSets, err := readPathSets(parts[0])
	if err != nil {
		return nil, err
	}
	content, err := readContent(parts[1])
	if err != nil {
		return nil, err
	}
	return &RawMap{
		PathSets: pathSets,
		Content:  content,
	}, nil
}

func ProduceMap(text string, observer MapObserver) error {
	text = strings.Replace(text, "\r\n", "\n", -1)
	raw, err := readRawMap(text)
	if err != nil {
		return err
	}
	observer.SetSize(NewLocationI(len(raw.Content), len(raw.Content[0]), len(raw.Content[0][0])))
	for x := 0; x < len(raw.Content); x++ {
		for y := 0; y < len(raw.Content[x]); y++ {
			for z := 0; z < len(raw.Content[x][y]); z++ {
				entry := raw.Content[x][y][z]
				types, found := raw.PathSets[entry]
				if !found {
					return fmt.Errorf("no such abbreviation: %s", entry)
				}
				for _, path := range types {
					err := observer.AddAtom(NewLocationI(x, y, z), path)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
