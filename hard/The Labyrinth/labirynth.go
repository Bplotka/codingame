package main

import (
	"fmt"
	"math"
	"os"
	"strings"
)

type Dir int

func (d Dir) Go() {
	if d == RIGHT {
		fmt.Println("RIGHT")
	} else if d == DOWN {
		fmt.Println("DOWN")
	} else if d == LEFT {
		fmt.Println("LEFT")
	} else if d == UP {
		fmt.Println("UP")
	} else {
		fmt.Fprintln(os.Stderr, "Error")
	}
}

func (d Dir) Next() Dir {
	if d == RIGHT {
		return DOWN
	} else if d == DOWN {
		return LEFT
	} else if d == LEFT {
		return UP
	} else if d == UP {
		return RIGHT
	} else {
		fmt.Fprintln(os.Stderr, "Error")
		return NONE
	}
}
func (d Dir) Opposite() Dir {
	if d == RIGHT {
		return LEFT
	} else if d == DOWN {
		return UP
	} else if d == LEFT {
		return RIGHT
	} else if d == UP {
		return DOWN
	} else {
		return NONE
	}

}

const (
	JETPACK_ROUNDS = 1200

	RIGHT = Dir(0)
	DOWN  = Dir(1)
	LEFT  = Dir(2)
	UP    = Dir(3)
	NONE  = Dir(4)
)

type pos struct {
	x int // row addr
	y int // col addr
}

type field struct {
	availableDirs      map[Dir]*edge
	availableDirsOrder []Dir

	isProccessed  bool
	isJunction    bool
	isStartPoint  bool
	isControlRoom bool

	distance int
	controlDistance int
}

func (f *field) getFewestMarkDir(previousDir Dir) Dir {
	if len(f.availableDirsOrder) == 1 {
		return f.availableDirsOrder[0]
	}

	minMark := math.MaxInt32
	minDir := previousDir.Opposite()
	for _, dir := range f.availableDirsOrder {
		if previousDir != NONE && dir == previousDir.Opposite() {
			continue
		}
		path := f.availableDirs[dir]
		pathMinMark := 0
		if path != nil {
			pathMinMark = path.marks
		}

		if pathMinMark < minMark {
			minMark = pathMinMark
			minDir = dir
		}
	}

	return minDir
}

func (f *field) getFewestMarkDirNotExceedingAlarmRound(previousDir Dir, alarmRounds int) Dir {
	if len(f.availableDirsOrder) == 1 {
		return f.availableDirsOrder[0]
	}

	minMark := math.MaxInt32
	minDir := previousDir.Opposite()
	for _, dir := range f.availableDirsOrder {
		if previousDir != NONE && dir == previousDir.Opposite() {
			continue
		}
		path := f.availableDirs[dir]
		if path == nil || path.distance >= alarmRounds {
			continue
		}

		pathMinMark := path.marks
		if pathMinMark < minMark {
			minMark = pathMinMark
			minDir = dir
		}
	}

	return minDir
}

func (f *field) getLowestDistanceDir(previousDir Dir) Dir {
	lowestDist := math.MaxInt32
	minDir := previousDir.Opposite()
	for _, dir := range f.availableDirsOrder {
		if dir == previousDir.Opposite() {
			continue
		}

		path := f.availableDirs[dir]
		pathMinDist := math.MaxInt32
		if path != nil {
			pathMinDist = path.distance
		}

		if pathMinDist < lowestDist {
			lowestDist = pathMinDist
			minDir = dir
		}
	}

	return minDir
}

func (f *field) getLowestDistanceControlDir(previousDir Dir) Dir {
	lowestDist := math.MaxInt32
	minDir := previousDir.Opposite()
	for _, dir := range f.availableDirsOrder {
		if dir == previousDir.Opposite() {
			continue
		}

		path := f.availableDirs[dir]
		pathMinDist := math.MaxInt32
		if path != nil {
			pathMinDist = path.controlRoomDistance
		}

		if pathMinDist < lowestDist {
			lowestDist = pathMinDist
			minDir = dir
		}
	}

	return minDir
}

type edge struct {
	marks       int
	distance    int // In rounds from home.
	localDistance int
	controlRoomDistance int
	previousDir Dir
}

func newEdge() *edge {
	return &edge{
		controlRoomDistance: math.MaxInt32,
	}
}

type runner struct {
	maze [][]*field

	jetPackRounds int
	rows          int
	cols          int
	alarmRounds   int

	controlRoomPos pos
	kirkPos        pos
}

func (r *runner) run() {
	// First iteration grabs the starting point.
	fmt.Scan(&r.kirkPos.x, &r.kirkPos.y)

	r.maze = make([][]*field, r.rows)
	for i := 0; i < r.rows; i++ {
		r.maze[i] = make([]*field, r.cols)
		var row string
		fmt.Scan(&row)
		for j, char := range strings.Split(row, "") {
			if char == "C" {
				r.controlRoomPos = pos{x: i, y: j}
				r.maze[i][j] = &field{isControlRoom: true,controlDistance: math.MaxInt32}
				continue
			}

			if char == "." {
				r.maze[i][j] = &field{controlDistance: math.MaxInt32}
				continue
			}

			if char == "T" {
				r.maze[i][j] = &field{isStartPoint: true, controlDistance: math.MaxInt32}
				continue
			}
		}
	}
	r.touchAlarm()
}

func (r *runner) touchAlarm() {
	var currentPath *edge
	for {
		//nextMove := RIGHT
		// if r.controlRoomPos != nil {
		//    xDiff := r.controlRoomPos.x - r.kirkPos.x
		//    yDiff := r.controlRoomPos.y - r.kirkPos.y
		//
		//    if xDiff > yDiff {
		//       if xDiff > 0 {
		//          nextMove = DOWN
		//       } else {
		//          nextMove = UP
		//       }
		//    } else {
		//       if yDiff > 0 {
		//          nextMove = RIGHT
		//       } else {
		//          nextMove = LEFT
		//       }
		//    }
		//}
		currentField := r.maze[r.kirkPos.x][r.kirkPos.y]

		isAlreadyMarked := true
		if !currentField.isProccessed {
			isAlreadyMarked = false
			controlRoomDir := r.processField(currentField)

			if controlRoomDir != NONE {
				// We could go there and set alarm, but let's check if we have enough way home.
				if currentPath == nil || currentPath.distance+1 <= r.alarmRounds {
					// We are ok! Let's set alarm and let's go back.
					if currentPath == nil {
						currentPath = newEdge()
					}
					currentPath.distance++
					currentPath.localDistance++
					currentField.availableDirs[currentPath.previousDir.Opposite()] = currentPath
					r.setAlarmAndGoBack(controlRoomDir)
				} else {
					fmt.Fprintln(os.Stderr,
						fmt.Sprintf("Control room is nearby, but we have too long distance %d to go. Alarm: %d",
							currentPath.distance, r.alarmRounds))
					currentPath.controlRoomDistance =  1 - (currentPath.localDistance + 1)
				}
			}
		}

		// Fill previousDir and inc distance when currentPath is present.
		previousDir := NONE
		if currentPath != nil {
			currentPath.distance++
			currentPath.localDistance++
			previousDir = currentPath.previousDir
		}

		dir := NONE
		if currentPath != nil && currentPath.distance >= r.alarmRounds {
			// Stop searching - not worth it.
			fmt.Fprintln(os.Stderr, "Putting artifical wall! Distance is too long.")
			dir = currentField.getFewestMarkDirNotExceedingAlarmRound(previousDir, r.alarmRounds)
			currentPath.marks = 2
			currentPath.distance-=2
		} else {
			// Find a direction with the fewest marks. Excluding the previousDir if not NONE.
			dir = currentField.getFewestMarkDir(previousDir)
		}

		fmt.Fprintln(os.Stderr,
			fmt.Sprintf("Dirs found: %v. Field was already processed? %v. \nDir chosen: %v. Prev dir: %v. Curr path: %+v",
				currentField.availableDirs, isAlreadyMarked, dir, previousDir, currentPath))

		if currentField.isJunction {
			// It's junction we need to increase the mark and create a new path if it's nil.
			if currentPath != nil {
				if isAlreadyMarked && currentPath.marks == 1 &&
					currentField.availableDirs[currentPath.previousDir.Opposite()] == nil {
					// Walk back!
					dir = currentPath.previousDir.Opposite()
				}

				currentPath.marks++
				if currentField.availableDirs[currentPath.previousDir.Opposite()] != nil {
					// If it is already set, it means we had been there (e.g it was dead end)
					currentPath.marks = 2
				}

				currentField.availableDirs[currentPath.previousDir.Opposite()] = currentPath
			}

			if currentField.availableDirs[dir] == nil {
				currentField.availableDirs[dir] = newEdge()
				currentField.availableDirs[dir].distance = math.MaxInt32 // !
			}

			if currentPath != nil && currentPath.controlRoomDistance != math.MaxInt32 {
				// Set control room distance if trackable already.
				lowestCtrlDistDir := currentField.getLowestDistanceControlDir(NONE)
				lowestCtrlDist := 0
				if lowestCtrlDistDir != NONE {
					lowestCtrlDist = currentField.availableDirs[lowestCtrlDistDir].controlRoomDistance
				}
				if lowestCtrlDistDir == currentPath.previousDir.Opposite() {
					lowestCtrlDist += currentPath.localDistance
				}

				if currentPath.distance + lowestCtrlDist <= r.alarmRounds {
					// We are good to follow the best ctrl path and return!
					dir = r.returnToControlRoom()
					currentPath = currentField.availableDirs[dir]
					currentPath.previousDir = dir
					continue
				}

				currentField.availableDirs[dir].controlRoomDistance = lowestCtrlDist
				fmt.Fprintln(os.Stderr, fmt.Sprintf("Control room update %v", lowestCtrlDist))
				// Check if we are close enough to home.
			}

			lowestDistDir := currentField.getLowestDistanceDir(NONE)
			lowestDist := 0
			if lowestDistDir != NONE && !currentField.isStartPoint {
				lowestDist = currentField.availableDirs[lowestDistDir].distance
			}

			currentField.availableDirs[dir].distance = lowestDist
			currentField.availableDirs[dir].localDistance = 0
		} else {
			if currentPath == nil {
				// Just the first iteration.
				currentPath = newEdge()
			}

			currentField.availableDirs[currentPath.previousDir.Opposite()] = currentPath

			// Just path, so if path nil it is the same as currentPath.
			if currentField.availableDirs[dir] == nil {
				if currentField.isStartPoint {
					currentField.availableDirs[dir] = newEdge()
				} else {
					currentField.availableDirs[dir] = currentPath
				}
			}
		}

		// path is now our currentPath!
		currentPath = currentField.availableDirs[dir]
		currentPath.previousDir = dir

		dir.Go()
		r.jetPackRounds--
		r.updateMazeFromInput()
	}
}

func (r *runner) returnToControlRoom() Dir{
	previousDir := NONE
	for {
		currentField := r.maze[r.kirkPos.x][r.kirkPos.y]

		dir := currentField.getLowestDistanceControlDir(previousDir)

		fmt.Fprintln(os.Stderr,
			fmt.Sprintf("CTRL! Dirs found: %v\nDir chosen: %v. Prev dir: %v ctrl dist: %d",
				currentField.availableDirs, dir, previousDir, currentField.availableDirs[dir].controlRoomDistance))

		if currentField.availableDirs[dir].controlRoomDistance <= 2 {
			// We are close, in one move we will be close. Return to normal flow.
			return dir
		}

		dir.Go()
		previousDir = dir
		r.jetPackRounds--
		r.updateMazeFromInput()
	}
}

func (r *runner) setAlarmAndGoBack(controllerRoomDir Dir) {
	controllerRoomDir.Go()
	r.jetPackRounds--
	r.updateMazeFromInput()

	currentField := r.maze[r.kirkPos.x][r.kirkPos.y]

	if !currentField.isProccessed {
		_ = r.processField(currentField)
	}

	previousDir := controllerRoomDir
	for {
		currentField = r.maze[r.kirkPos.x][r.kirkPos.y]

		dir := currentField.getLowestDistanceDir(previousDir)

		fmt.Fprintln(os.Stderr,
			fmt.Sprintf("Dirs found: %v\nDir chosen: %v. Prev dir: %v",
				currentField.availableDirs, dir, previousDir))

		dir.Go()
		previousDir = dir
		r.jetPackRounds--
		r.updateMazeFromInput()
	}
}

func (r *runner) processField(currentField *field) (controlRoomDir Dir) {
	// Let's check what is around!
	dir := UP
	controlRoomDir = NONE
	currentField.availableDirs = map[Dir]*edge{}
	for i := 0; i < 4; i++ {
		dir = dir.Next()
		adjacentField := r.whatIsIn(r.kirkPos, dir, 1)
		if adjacentField == nil {
			continue
		}

		if adjacentField.isControlRoom {
			controlRoomDir = dir
			fmt.Fprintln(os.Stderr, "Found!")
			// DO not set the control room as available path (:
			continue
		}

		currentField.availableDirs[dir] = nil
		currentField.availableDirsOrder = append(currentField.availableDirsOrder, dir)
	}

	if len(currentField.availableDirs) > 2 {
		currentField.isJunction = true
	}
	currentField.isProccessed = true

	return controlRoomDir
}

func (r *runner) updateMazeFromInput() {
	// Kirk location.
	fmt.Scan(&r.kirkPos.x, &r.kirkPos.y)

	for i := 0; i < r.rows; i++ {
		var row string
		fmt.Scan(&row)
		for j, char := range strings.Split(row, "") {
			if r.maze[i][j] != nil {
				continue
			}

			if char == "C" {
				r.controlRoomPos = pos{x: i, y: j}
				r.maze[i][j] = &field{isControlRoom: true, controlDistance: math.MaxInt32}
				continue
			}

			if char == "." {
				r.maze[i][j] = &field{controlDistance: math.MaxInt32}
				continue
			}
		}
	}
}

func (r *runner) whatIsIn(p pos, dir Dir, dist int) *field {
	if dir == RIGHT {
		return r.maze[p.x][p.y+dist]
	}
	if dir == DOWN {
		return r.maze[p.x+dist][p.y]
	}
	if dir == LEFT {
		return r.maze[p.x][p.y-dist]
	}
	if dir == UP {
		return r.maze[p.x-dist][p.y]
	}

	return nil
}

// Author: witcher92
// Inspired by Trémaux's algorithm.
func main() {
	// Rows: number of rows.
	// Cols: number of columns.
	// AlarmRounds: number of rounds between the time the alarm countdown is activated and the time the alarm goes off.

	r := runner{}
	fmt.Scan(&r.rows, &r.cols, &r.alarmRounds)
	r.run()
}