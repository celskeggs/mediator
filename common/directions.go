package common

import "github.com/celskeggs/mediator/util"

type Direction uint8

const (
	North     Direction = 1
	South     Direction = 2
	East      Direction = 4
	West      Direction = 8
	Northeast           = North | East
	Northwest           = North | West
	Southeast           = South | East
	Southwest           = South | West
)

func (d Direction) IsValid() bool {
	return (d&(North|South)) != (North|South) &&
		(d&(East|West)) != (East|West) &&
		(d>>4) == 0 &&
		d != 0
}

func (d Direction) NearestCardinal() Direction {
	if !d.IsValid() {
		return d
	}
	util.FIXME("figure out how to err towards the latest direction with diagonals")
	if (d & North) != 0 {
		return North
	}
	if (d & South) != 0 {
		return South
	}
	return d
}

func (d Direction) XY() (x, y int) {
	switch d {
	case South:
		return 0, -1
	case North:
		return 0, 1
	case East:
		return 1, 0
	case West:
		return -1, 0
	case Southeast:
		return 1, -1
	case Southwest:
		return -1, -1
	case Northeast:
		return 1, 1
	case Northwest:
		return -1, 1
	default:
		panic("given invalid direction to convert to XY")
	}
}

func (d Direction) EightDirectionIndex() uint {
	// reflects the EightDirections array below
	switch d {
	case South:
		return 0
	case North:
		return 1
	case East:
		return 2
	case West:
		return 3
	case Southeast:
		return 4
	case Southwest:
		return 5
	case Northeast:
		return 6
	case Northwest:
		return 8
	default:
		panic("given invalid direction to convert to index")
	}
}

var EightDirections = []Direction{
	South, North, East, West, Southeast, Southwest, Northeast, Northwest,
}

func (d Direction) FourDirectionIndex() uint {
	// reflects the FourDirections array below
	switch d {
	case South:
		return 0
	case North:
		return 1
	case East:
		return 2
	case West:
		return 3
	default:
		panic("given invalid direction to convert to index")
	}
}

var FourDirections = []Direction{
	South, North, East, West,
}
