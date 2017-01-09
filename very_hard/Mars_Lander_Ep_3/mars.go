package main

import (
	"fmt"
	"math"
	"os"
)

const MarsGravity = -3.711

func main() {
	(&lander{}).Land()
}

func d(format string, a ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, a...))
}

// 2d vector.
type point struct {
	x, y float64
}

func newPoint(x, y int) point {
	return point{x: float64(x), y: float64(y)}
}

// Dot returns the standard dot product of v and ov.
func (p point) Dot(ov point) float64 { return p.x*ov.x + p.y*ov.y }

// Norm returns the vector's norm.
func (p point) Norm() float64 { return math.Sqrt(p.Dot(p)) }

// Normalize returns a unit vector in the same direction as v.
func (p point) Normalize() point {
	if p == (point{0, 0}) {
		return p
	}
	return p.Mul(1 / p.Norm())
}

// Mul returns the standard scalar product of v and m.
func (p point) Mul(m float64) point { return point{m * p.x, m * p.y} }

// Sub returns the standard vector difference of v and ov.
func (p point) Sub(ov point) point { return point{p.x - ov.x, p.y - ov.y} }

// Angle returns the angle between v and ov. (Degrees)
func (p point) Angle(ov point) float64 {
	s := ov.Sub(p)
	return math.Atan2(s.y, s.x) * (180 / math.Pi)
}

func (p point) didCrossedSurface(surface []point) bool {
	for i, surfacePt := range surface {
		if p.x > surfacePt.x {
			continue
		}

		a := surfacePt.y - surface[i-1].y
		b := surfacePt.x - surface[i-1].x
		a2 := p.x - surface[i-1].x
		x := (b * a2) / a
		if (p.y - surface[i-1].y) <= x {
			return true
		}
	}

	return false
}

func (p point) print() string {
	return fmt.Sprintf("[%f, %f]", p.x, p.y)
}

type lander struct {
	// Read only - gathered from env.
	pos                                   point
	hSpeed, vSpeed, fuel, rotation, power int

	surface              []point
	landingCenterPoint   *point
	landingSiteTolerance int
}

func (l *lander) Land() {
	l.discoverSurfaceAndLandingSite()
	l.gatherInput()
	d("Landing center: %s, tolerance: %d", l.landingCenterPoint.print(), l.landingSiteTolerance)

	for {
		where, when, eVSpeed, eHSpeed := l.estimateSurfaceCross()
		d("Estimated landing: %s | epochs: %d, eV: %f, eH %f",
			where.print(), when, eVSpeed, eHSpeed)

		angleToAdjust, distance := l.angleAndDistanceToTarget(where)
		d("Angle to adjust: %f | distance %f", angleToAdjust, distance)

		// Calculate what throttle and rotation to add.
		throttle := 0
		if angleToAdjust > 5 {
			throttle = 4
		}

		l.engineSettings(int(angleToAdjust), throttle)
		l.gatherInput()
	}
}

func (l *lander) rotationAndPowerToAdjustAngle(angleToAdjust float64) {
	//a := float64(l.power) * math.Cos(float64(l.rotation))
	//l.rotation
}

func (l *lander) angleAndDistanceToTarget(currentTarget point) (angle float64, distance float64) {
	desiredDir := l.landingCenterPoint.Sub(l.pos)
	headingDir := currentTarget.Sub(l.pos)
	angle = desiredDir.Angle(headingDir)
	return angle, l.landingCenterPoint.Sub(l.pos).Norm()
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
		surface = append(surface, newPoint(landX, landY))

		if l.landingCenterPoint == nil && lastY == landY {
			l.landingSiteTolerance = landX - int(surface[i-1].x)
			p := newPoint((l.landingSiteTolerance/2)+int(surface[i-1].x), lastY)
			l.landingCenterPoint = &p
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
	var X, Y int
	fmt.Scan(&X, &Y, &l.hSpeed, &l.vSpeed, &l.fuel, &l.rotation, &l.power)
	l.pos = newPoint(X, Y)
}

func (l *lander) engineSettings(rotationSetting, throttleSetting int) {
	// rotate power. rotate is the desired rotation angle. [ MINUS = RIGHT ]
	// power is the desired thrust power.
	fmt.Printf("%d %d\n", rotationSetting, throttleSetting)
}

// Assuming no throttle.
func (l *lander) estimateSurfaceCross() (where point, when int, eVSpeed float64, eHSpeed float64) {
	epoch := 0
	tmpVSpeed := float64(l.vSpeed)
	tmpHSpeed := float64(l.hSpeed)
	pos := l.pos
	for {
		// Movement.
		pos.x += tmpHSpeed
		pos.y += tmpVSpeed

		if pos.didCrossedSurface(l.surface) {
			return pos, epoch, tmpVSpeed, tmpHSpeed
		}

		// Next speeds.
		epoch++
		tmpVSpeed += MarsGravity

		// It is stupid, but codingame calculates on ints, so we need to adjust based on that.
		tmpVSpeed = float64(int(tmpVSpeed))
	}
}
