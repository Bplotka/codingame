package main

import (
	"fmt"
	"math"
	"os"
	"sort"
)

var bricksNum int

func main() {
	fmt.Scan(&bricksNum)
	fmt.Printf("%d %d\n",
		logMin(buildNormalizedCuboid()).a,
		logMax(buildTallCuboid()).a,
	)
}

// Maximum (experienced that by doing manual experiments with small number of bricks).
func buildTallCuboid() match {
	return newMatch(bricksNum, 1)
}

// Minimum (experienced that by doing manual experiments with small number of bricks).
func buildNormalizedCuboid() match {
	// Try cubic root first.
	cubicRoot := math.Pow(float64(bricksNum), 1.0/3.0)
	if rounded, ok := tolerantIsInt(cubicRoot); ok {
		// It can be a cube!
		return newMatch(int(rounded), int(rounded))
	}

	// Let's grab all components for the brickNum (our cubic space V)
	var components []int
	smallestDivisible := 2
	number := bricksNum
	for number > 1 {
		smallestDivisible = findSmallestDivisible(number, smallestDivisible-1)
		components = append(components, smallestDivisible)

		number = number / smallestDivisible
	}

	// Sort from the biggest first!
	sort.Sort(sort.Reverse(sort.IntSlice(components)))

	bestMatch := newMatch(bricksNum, 1)
	if len(components) == 1 {
		return bestMatch
	}

	if len(components) == 2 || len(components) == 3 {
		return newMatch(components[0], components[1])
	}

	/*
		If we have more components than 3, then we need to:
		1) Naively: Do every permutation of these to find optimum. (That is what works! =D so no need for better solution.)
		2) We could optimize the solution since we know that x and y needs to be as close as possible to cubic root of V.

			We can construct function from surface and cubic space functions:
			func a(x, y, v int) int {
				return 2 * ((x * y) + (((x + y) * v) / (x * y)))
			}
			This function indicates (same as intuition) that x == y is the minimum of a. Having that in mind,
			we can just start from x ~ y and go down with x and up with y slowly and check if we can construct them having
			known components.

			This is not implemented here, since it was not needed to pass test cases (which proves that test cases should be better!)
	*/

	bestMatch = checkDeep([3]int{components[0], components[1], components[2]}, components[3:])
	return bestMatch
}

type match struct {
	x, y, z int
	v       int
	a       int // Surface.
}

func newMatch(x, y int) match {
	m := match{
		x: x,
		y: y,
		z: bricksNum / (x * y),
		v: bricksNum,
	}
	m.a = a(m.x, m.y, m.v)
	return m
}

func (m match) isSmaller(x, y int) bool {
	return m.a <= a(x, y, m.v)
}

func (m match) dimensions() (int, int, int) {
	return m.x, m.y, m.z
}

func a(x, y, v int) int {
	return 2 * ((x * y) + (((x + y) * v) / (x * y)))
}

func checkDeep(cuboid [3]int, components []int) match {
	if len(components) == 0 {
		return newMatch(cuboid[0], cuboid[1])
	}

	bestMatch := checkDeep(
		[3]int{
			cuboid[0] * components[0],
			cuboid[1],
			cuboid[2],
		},
		components[1:],
	)

	match2 := checkDeep(
		[3]int{
			cuboid[0],
			cuboid[1] * components[0],
			cuboid[2],
		},
		components[1:],
	)

	if match2.isSmaller(bestMatch.x, bestMatch.y) {
		bestMatch = match2
	}

	match3 := checkDeep(
		[3]int{
			cuboid[0],
			cuboid[1],
			cuboid[2] * components[0],
		},
		components[1:],
	)

	if match3.isSmaller(bestMatch.x, bestMatch.y) {
		bestMatch = match3
	}

	return bestMatch
}

func findSmallestDivisible(number int, greaterThan int) int {
	divisible := greaterThan + 1
	for divisible < number {
		if math.Mod(float64(number), float64(divisible)) != 0 {
			divisible++
			continue
		}
		return divisible
	}

	return number
}

func tolerantIsInt(number float64) (int, bool) {
	rounded := math.Floor(number + .5)
	if number-rounded > -0.0000001 && number-rounded < 0.0000001 {
		return int(rounded), true
	}

	return 0, false
}

func logMin(m match) match {
	fmt.Fprintln(os.Stderr, fmt.Sprintf("Minimum: %d x %d x %d", m.x, m.y, m.z))
	return m
}

func logMax(m match) match {
	fmt.Fprintln(os.Stderr, fmt.Sprintf("Maximum: %d x %d x %d", m.x, m.y, m.z))
	return m
}
