package main

import (
	"fmt"
	"os"
	"strings"
)

type BikeStatus struct {
	x      int
	lane   int
	speed  int
	isDead bool
}

type BridgeSolver struct {
	bridge [][]bool
	bikes  []BikeStatus

	bikesToSurvive int
	bridgeLength   int

	tmpSpeed int
}

func (b *BridgeSolver) Run() {
	var speed int
	fmt.Scan(&speed)
	for i := range b.bikes {
		var isDestroyed int
		fmt.Scan(&b.bikes[i].x, &b.bikes[i].lane, &isDestroyed)
		b.bikes[i].isDead = isDestroyed == 0
		b.bikes[i].speed = speed
	}

	b.tmpSpeed = speed
	sequence, ok := b.findPath([]string{}, b.bikes, "WAIT")
	if !ok {
		panic("Sorry, no way through this bridge ):")
	}

	fmt.Fprintln(os.Stderr, fmt.Sprintf("Lets run! %v", sequence))
	for _, op := range sequence {
		// A single line containing one of 6 keywords: SPEED, SLOW, JUMP, WAIT, UP, DOWN.
		fmt.Println(op)

		var speed int
		fmt.Scan(&speed)
		for _, bike := range b.bikes {
			var isDestroyed int
			fmt.Scan(&bike.x, &bike.lane, &isDestroyed)
			bike.isDead = isDestroyed == 0
			bike.speed = speed
		}
	}
}

func (b *BridgeSolver) simBikes(bikes []BikeStatus, op string) []BikeStatus {
	newBikes := append([]BikeStatus{}, bikes...)

	for i, bike := range newBikes {
		if bike.isDead {
			continue
		}
		newBikes[i] = b.simBike(bike, op)
	}
	return newBikes
}

func (b *BridgeSolver) simBike(bike BikeStatus, op string) BikeStatus {
	oldLane := -1
	switch op {
	case "SPEED":
		bike.speed++
	case "SLOW":
		bike.speed--
	case "UP":
		oldLane = bike.lane
		bike.lane--
	case "DOWN":
		oldLane = bike.lane
		bike.lane++
	}

	// Any hole in the meantime?
	if op != "JUMP" {
		for i := bike.x + 1; i < bike.x+bike.speed; i++ {
			if !b.bridge[bike.lane][i] {
				bike.isDead = true
			}

			if oldLane > -1 {
				if !b.bridge[oldLane][i] {
					bike.isDead = true
				}
			}
		}
	}

	if !b.bridge[bike.lane][bike.x+bike.speed] {
		bike.isDead = true
	}

	bike.x += bike.speed
	b.tmpSpeed = bike.speed
	return bike
}

func (b *BridgeSolver) isFinish(x int) bool {
	return x > b.bridgeLength
}

// Naive, recursive.
func (b *BridgeSolver) findPath(sequence []string, bikes []BikeStatus, previousOp string) ([]string, bool) {
	optionsToCheck := map[string]int{
		"SPEED": 0, "SLOW": 2, "WAIT": 1, "JUMP": 3, "UP": 4, "DOWN": 5,
	}

	if b.tmpSpeed <= 1 {
		delete(optionsToCheck, "SLOW")
	}

	if b.tmpSpeed <= 0 {
		delete(optionsToCheck, "WAIT")
	}

	if previousOp == "UP" {
		delete(optionsToCheck, "DOWN")
	} else if previousOp == "DOWN" {
		delete(optionsToCheck, "UP")
	}

	possibleOpsPerBike := []map[string]struct{}{}
	for _, bike := range bikes {
		if bike.isDead {
			continue
		}

		// FINISH.
		if b.isFinish(bike.x) {
			fmt.Fprintln(os.Stderr, "Found path!")
			return sequence, true
		}

		if bike.lane == 0 {
			delete(optionsToCheck, "UP")
		} else if bike.lane == 3 {
			delete(optionsToCheck, "DOWN")
		}

		possibleOps := map[string]struct{}{}
		for op := range optionsToCheck {
			newBike := b.simBike(bike, op)
			if !newBike.isDead {
				possibleOps[op] = struct{}{}
			}
		}

		possibleOpsPerBike = append(possibleOpsPerBike, possibleOps)
	}

	for op := range optionsToCheck {
		remainAlive := 0
		for _, bike := range possibleOpsPerBike {
			_, ok := bike[op]
			if ok {
				remainAlive++
			}
		}

		if len(possibleOpsPerBike)+remainAlive <= b.bikesToSurvive {
			delete(optionsToCheck, op)
		}
	}

	// ORDER!
	orderedOptionsToCheck := []string{}
	for op, weight := range optionsToCheck {
		if len(orderedOptionsToCheck) > 0 && optionsToCheck[orderedOptionsToCheck[0]] > weight {
			orderedOptionsToCheck = append([]string{op}, orderedOptionsToCheck...)
		}

		orderedOptionsToCheck = append(orderedOptionsToCheck, op)
	}

	for _, op := range orderedOptionsToCheck {
		// GO routines?
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Checking op %v on seq %v", op, sequence))
		newSequence, ok := b.findPath(append(sequence, op), b.simBikes(bikes, op), op)
		if ok {
			return newSequence, true
		}
	}

	return []string{}, false
}

func main() {
	var bikeNum int
	var bikesToSurvive int
	fmt.Scan(&bikeNum)
	fmt.Scan(&bikesToSurvive)

	b := BridgeSolver{bikesToSurvive: bikesToSurvive}
	for i := 0; i < 4; i++ {
		var line string
		fmt.Scan(&line)
		fmt.Fprintln(os.Stderr, line)
		b.bridge = append(b.bridge, []bool{})
		for _, char := range strings.Split(line, "") {
			b.bridge[i] = append(b.bridge[i], char == ".")
		}

		b.bridgeLength = len(b.bridge[i])

		// Fake bridge, just for sim to work.
		for x := b.bridgeLength - 1; x < b.bridgeLength+50; x++ {
			b.bridge[i] = append(b.bridge[i], true)
		}
	}

	for i := 0; i < bikeNum; i++ {
		b.bikes = append(b.bikes, BikeStatus{})
	}

	b.Run()
}
