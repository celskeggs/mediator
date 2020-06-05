package pathfinding

import (
	"container/heap"
	"math"
)

// A* search algorithm implementation on a grid, accepting diagonal traversal.

type Point struct {
	X, Y uint
}

func (p Point) Adjacent(exists func(Point) bool) (px []Point) {
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			np := Point{
				X: p.X + uint(x),
				Y: p.Y + uint(y),
			}
			if (x != 0 || y != 0) && exists(np) {
				px = append(px, np)
			}
		}
	}
	return px
}

// compute the largest of the horizontal and vertical distance, because diagonals.
func (p Point) Heuristic(goal Point) uint32 {
	var distX, distY uint
	if p.X >= goal.X {
		distX = p.X - goal.X
	} else {
		distX = goal.X - p.X
	}
	if p.Y >= goal.Y {
		distY = p.Y - goal.Y
	} else {
		distY = goal.Y - p.Y
	}
	if distX > distY {
		return uint32(distX)
	} else {
		return uint32(distY)
	}
}

type PointEntry struct {
	point Point
	index int

	hasCameFrom bool
	cameFrom    Point
	gScore      uint32
	fScore      uint32
}

// based on example from https://golang.org/pkg/container/heap/
type SearchState struct {
	points map[Point]*PointEntry
	pqueue []*PointEntry
}

func NewSearchState() *SearchState {
	ss := &SearchState{
		points: map[Point]*PointEntry{},
		pqueue: nil,
	}
	heap.Init(ss)
	return ss
}

func (ss *SearchState) GetEntry(p Point) *PointEntry {
	pe, ok := ss.points[p]
	if !ok {
		pe = &PointEntry{
			point:    p,
			index:    -1,
			cameFrom: Point{},
			gScore:   math.MaxUint32,
			fScore:   math.MaxUint32,
		}
		ss.points[p] = pe
	}
	return pe
}

func (ss *SearchState) Len() int {
	return len(ss.pqueue)
}

func (ss *SearchState) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return ss.pqueue[i].fScore < ss.pqueue[j].fScore
}

func (ss *SearchState) Swap(i, j int) {
	ss.pqueue[i], ss.pqueue[j] = ss.pqueue[j], ss.pqueue[i]
	ss.pqueue[i].index = i
	ss.pqueue[j].index = j
}

func (ss *SearchState) Push(x interface{}) {
	pe := x.(*PointEntry)
	pe.index = len(ss.pqueue)
	ss.pqueue = append(ss.pqueue, pe)
}

func (ss *SearchState) Pop() interface{} {
	old := ss.pqueue
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	ss.pqueue = ss.pqueue[0 : n-1]
	return item
}

func reversePoints(x []Point) {
	for i := 0; i < len(x)/2; i++ {
		x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
	}
}

// implemented from https://en.wikipedia.org/wiki/A*_search_algorithm
func Search(start Point, goal Point, canTraverse func(Point) bool) []Point {
	ss := NewSearchState()
	startEnt := ss.GetEntry(start)
	startEnt.gScore = 0
	startEnt.fScore = start.Heuristic(goal)
	heap.Push(ss, startEnt)

	for ss.Len() > 0 {
		current := heap.Pop(ss).(*PointEntry)
		if current.point == goal {
			// reconstruct path
			totalPath := []Point{current.point}
			for current.hasCameFrom {
				current = ss.GetEntry(current.cameFrom)
				totalPath = append(totalPath, current.point)
			}
			reversePoints(totalPath)
			return totalPath
		}

		tentativeGScore := current.gScore + 1
		for _, neighborPoint := range current.point.Adjacent(canTraverse) {
			neighbor := ss.GetEntry(neighborPoint)
			if tentativeGScore < neighbor.gScore {
				neighbor.hasCameFrom = true
				neighbor.cameFrom = current.point
				neighbor.gScore = tentativeGScore
				neighbor.fScore = neighbor.gScore + neighbor.point.Heuristic(goal)
				if neighbor.index == -1 {
					heap.Push(ss, neighbor)
				} else {
					heap.Fix(ss, neighbor.index)
				}
			}
		}
	}

	// goal not reached
	return nil
}
