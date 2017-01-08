package main

import (
	"fmt"
	"os"
)

func main() {
	(&lander{}).Land()
}

func d(format string, a ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, a...))
}

type point struct {
	x, y int
}

func (p point) print() string {
	return fmt.Sprintf("[%d, %d]", p.x, p.y)
}

type lander struct{
	// Read only - gathered from env.
	X, Y, hSpeed, vSpeed, fuel, rotation, power int

	surface []point
	landingCenterPoint *point
	landingSiteTolerance int
}

func (l* lander) Land() {
	l.discoverSurfaceAndLandingSite()
	l.gatherInput()
	d("Landing center: %s, tolerance: %d", l.landingCenterPoint.print(), l.landingSiteTolerance)

	for {

		l.engineSettings(-90, 4)
		l.gatherInput()
	}
}

func (l *lander) discoverSurfaceAndLandingSite() {
	// surfaceN: the number of points used to draw the surface of Mars.
	var surfaceN int
	fmt.Scan(&surfaceN)

	var surface []point
	lastY := -1
	for i := 0; i < surfaceN; i++ {
		// landX: X coordinate of a surface point. (0 to 6999)
		// landY: Y coordinate of a surface point. By linking all the points together in a sequential fashion,
		// you form the surface of Mars.
		var landX, landY int
		fmt.Scan(&landX, &landY)
		surface = append(surface, point{x: landX, y: landY})

		if l.landingCenterPoint == nil && lastY == landY {
			l.landingSiteTolerance = landX - surface[i-1].x
			l.landingCenterPoint = &point{x: (l.landingSiteTolerance / 2) + surface[i-1].x, y: lastY}
		}
		lastY = landY
	}
	l.surface = surface
}

func (l *lander) gatherInput() {
	// hSpeed: the horizontal speed (in m/s), can be negative.
	// vSpeed: the vertical speed (in m/s), can be negative.
	// fuel: the quantity of remaining fuel in liters.
	// rotation: the rotation angle in degrees (-90 to 90).
	// power: the thrust power (0 to 4).
	fmt.Scan(&l.X, &l.Y, &l.hSpeed, &l.vSpeed, &l.fuel, &l.rotation, &l.power)
}

func (l *lander) engineSettings(rotationSetting, throttleSetting int) {
	// rotate power. rotate is the desired rotation angle. [ MINUS = RIGHT ]
	// power is the desired thrust power.
	fmt.Printf("%d %d\n", rotationSetting, throttleSetting)
}