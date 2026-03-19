package main

type Position struct {
	X int
	Y int
}

func (position *Position) Equals(x, y int) bool {
	return position.X == x && position.Y == y
}

func ComparePositions(pos1, pos2 Position) int {
	if pos1.Y < pos2.Y {
		return -1
	} else if pos1.Y > pos2.Y {
		return 1
	} else if pos1.X < pos2.X {
		return -1
	} else if pos1.X > pos2.X {
		return 1
	} else {
		return 0
	}
}
