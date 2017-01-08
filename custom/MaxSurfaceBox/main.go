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
		cuboidSurface(lmin(buildNormalizedCuboid(bricks))),
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
	rounded := 1
	// Find the biggest cubic root that can be fitted in searched cubic space.
	for {
		var ok bool
		rounded, ok = tolerantIsInt(math.Pow(float64(lowerBricksNum), 1.0/3.0))
		if ok {
			break
		}

		lowerBricksNum--
		continue
	}

	cuboid := []int{rounded, rounded, rounded}
	cuboid[0] = findSmallestDivisible(bricksNum, cuboid[0])
	excludedDims := map[int]struct{}{
		0: struct{}{},
	}
	for {
		if cuboidSpace(cuboid) == bricksNum {
			break
		}

		if cuboidSpace(cuboid) < bricksNum {
			dimID := minDim(cuboid, excludedDims)
			cuboid[dimID] = findSmallestDivisible(bricksNum, cuboid[dimID])

			if len(excludedDims) >= 2 {
				return buildTallCuboid(bricksNum)
			}

			excludedDims[dimID] = struct{}{}
			continue
		}

		dimID := maxDim(cuboid, excludedDims)
		cuboid[dimID] = findBiggestDivisible(bricksNum, cuboid[dimID])
	}

	return cuboid[0], cuboid[1], cuboid[2]
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

func findBiggestDivisible(number int, smallerThan int) int {
	divisible := smallerThan - 1
	for divisible > 1 {
		if math.Mod(float64(number), float64(divisible)) != 0 {
			divisible--
			continue
		}
		return divisible
	}

	return 1
}

func minDim(cuboid []int, excludingDims map[int]struct{}) int {
	min := math.MaxInt32
	minIndex := -1
	for i, value := range cuboid {
		if _, ok := excludingDims[i]; ok {
			continue
		}

		if value < min {
			minIndex = i
			min = value
		}
	}

	return minIndex
}

func maxDim(cuboid []int, excludingDims map[int]struct{}) int {
	max := 0
	maxIndex := -1
	for i, value := range cuboid {
		if _, ok := excludingDims[i]; ok {
			continue
		}

		if value > max {
			maxIndex = i
			max = value
		}
	}

	return maxIndex
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

func cuboidSpace(cuboid []int) int {
	return cuboid[0] * cuboid[1] * cuboid[2]
}
