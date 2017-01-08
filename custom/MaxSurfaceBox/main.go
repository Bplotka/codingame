package main

import (
	"fmt"
	"math"
	"os"
)

func main() {
	var bricks int
	fmt.Scan(&bricks)

	fmt.Printf("%d %d\n",
		cuboidSurface(lmin(buildNormalizedCuboid2(bricks))),
		cuboidSurface(lmax(buildTallCuboid(bricks))),
	)
}

// Maximum (experienced that by doing manual experiments with small number of bricks).
func buildTallCuboid(bricksNum int) (int, int, int) {
	return bricksNum, 1, 1
}

// Minimum (experienced that by doing manual experiments with small number of bricks).
func buildNormalizedCuboid(bricksNum int) (int, int, int) {
	// Try cubic root first.
	if rounded, ok := tolerantIsInt(math.Pow(float64(bricksNum), 1.0/3.0)); ok {
		// It can be a cube!
		return int(rounded), int(rounded), int(rounded)
	}
	lowerBricksNum := bricksNum - 1
	for lowerBricksNum > 0 {
		rounded, ok := tolerantIsInt(math.Pow(float64(lowerBricksNum), 1.0/3.0))
		if !ok {
			lowerBricksNum--
			continue
		}

		rest := bricksNum - lowerBricksNum
		// Try to build wall/ walls on top of found cubic. We can end up with something like
		// rounded x rounded x (rounded + some layer).
		if rest == (rounded * rounded) {
			return rounded, rounded, rounded + 1
		}

		restExcludingOneLayer := rest - (rounded * rounded)
		if restExcludingOneLayer < 0 {
			// Not enough bricks even for single layer. Damn it.
			lowerBricksNum--
			continue
		}

		// One layer and something! Try to now build up on different dimension and have
		// rounded x (rounded + some layer) x (rounded + some layer).
		if restExcludingOneLayer != (rounded * (rounded + 1)) {
			lowerBricksNum--
			continue
		}

		// Nice!
		return rounded, rounded + 1, rounded + 1
	}

	// fallback to tall cuboid.
	return buildTallCuboid(bricksNum)
}

// Minimum (experienced that by doing manual experiments with small number of bricks).
func buildNormalizedCuboid2(bricksNum int) (int, int, int) {
	// Try cubic root first.
	if rounded, ok := tolerantIsInt(math.Pow(float64(bricksNum), 1.0/3.0)); ok {
		// It can be a cube!
		return int(rounded), int(rounded), int(rounded)
	}

	// ok then let's start will tall one and try to pack it down.
	x, y, z := buildTallCuboid(bricksNum)
	for {
		// Pack x in y dir.
		rounded, ok := tolerantIsInt(float64(x) / 2.)
		if !ok {
			break
		}
		x = rounded
		y = 2 * y

		if x <= z || x <= y {
			break
		}

		// Pack x in z dir.
		rounded, ok = tolerantIsInt(float64(x) / 2.)
		if !ok {
			break
		}
		x = rounded
		z = 2 * z

		if x <= z || x <= y {
			break
		}
	}
	return x, y, z
}

func tolerantIsInt(number float64) (int, bool) {
	rounded := math.Floor(number + .5)
	if number-rounded > -0.0000001 && number-rounded < 0.0000001 {
		return int(rounded), true
	}

	return 0, false
}

func lmin(x, y, z int) (int, int, int) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf("Minimum: %d x %d x %d", x, y, z))
	return x, y, z
}

func lmax(x, y, z int) (int, int, int) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf("Maximum: %d x %d x %d", x, y, z))
	return x, y, z
}

func cuboidSurface(x, y, z int) int {
	return (2 * x * y) + (2 * y * z) + (2 * x * z)
}
