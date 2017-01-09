package main

import (
	"fmt"
	"math"
	"os"
)

const (
	MarsGravity = -3.711
	MaxHSpeed = 20
	MaxVSpeed = 40
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			d("ERROR: %v", r)
		}
	}()
	(&lander{}).Land()
}

func d(format string, a ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, a...))
}

type lander struct {
	// Read only - gathered from env.
	pos                                   point
	hSpeed, vSpeed, fuel, rotation, power int

	surface              []point
	landingCenterPoint   *point
	landingSiteTolerance int
	landingSurfaceEndID  int
}

// Land using PID controller approach. In every iteration check the estimated landing and adjust.
func (l *lander) Land() {
	l.discoverSurfaceAndLandingSite()
	d("Landing center: %s, tolerance: %d", l.landingCenterPoint.print(), l.landingSiteTolerance)

	for {
		// --- MAIN LOOP ---
		l.gatherInput()

		where, isLandingArea, when, eVSpeed, eHSpeed := l.estimateSurfaceReachable()
		d("Estimated landing: %s | ok? %v | epochs: %d, eV: %f, eH %f",
			where.print(), isLandingArea, when, eVSpeed, eHSpeed)

		// TODO: Calculate obstacles, use bezier.

		angleToAdjust, distance := l.angleAndDistanceToTarget(where)
		d("Angle to adjust: %f | distance %f", angleToAdjust, distance)

		// Adjusting phase.
		throttle := 0
		if angleToAdjust > 5 {
			throttle = 4
		}

		// We are free-falling to Landing Area, cool - but we need to brake ):
		if isLandingArea &&
			(math.Abs(float64(l.hSpeed)) >= MaxHSpeed || math.Abs(float64(l.vSpeed)) >= MaxVSpeed) {

			brakingVec := newPoint(-l.hSpeed, -l.vSpeed)
			angleToAdjust = where.Sub(l.pos).Angle(brakingVec)

			throttle = 4
			if angleToAdjust > 15 {
				throttle = 0
			}
		}

		l.engineSettings(int(angleToAdjust), throttle)
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
			l.landingSurfaceEndID = i
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

func Max(a, b float64) float64 {
	if a>=b {
		return a
	}
	return b
}

// Assuming no throttle.
func (l *lander) estimateSurfaceReachable() (where point, isLandingArea bool, when int, eVSpeed float64, eHSpeed float64) {
	epoch := 0
	tmpVSpeed := float64(l.vSpeed)
	tmpHSpeed := float64(l.hSpeed)

	pos := l.pos
	for {
		// Movement.
		pos.x += tmpHSpeed
		pos.y += tmpVSpeed

		safeZone := newCollCircle(pos, tmpVSpeed, tmpHSpeed)
		if collisionPt, surfaceEndID, is := safeZone.isCollidingWithSurface(l.surface); is {
			return collisionPt, (surfaceEndID == l.landingSurfaceEndID), epoch, tmpVSpeed, tmpHSpeed
		}

		// Next speeds.
		epoch++
		tmpVSpeed += MarsGravity

		// It is stupid, but codingame calculates on ints, so we need to adjust based on that.
		tmpVSpeed = float64(int(tmpVSpeed))
	}
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

// Norm2 returns the square of the norm.
func (p point) Norm2() float64 { return p.Dot(p) }

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
func (p point) Mul(m float64) point { return point{x: m * p.x, y: m * p.y} }

// Sub returns the standard vector difference of v and ov.
func (p point) Sub(ov point) point { return point{x: p.x - ov.x, y: p.y - ov.y} }

// Angle returns the angle between v and ov. (Degrees)
func (p point) Angle(ov point) float64 {
	s := ov.Sub(p)
	return math.Atan2(s.y, s.x) * (180 / math.Pi)
}

// Distance returns the Euclidean distance between v and ov.
func (p point) Distance(ov point) float64 { return p.Sub(ov).Norm() }

func (p point) print() string {
	return fmt.Sprintf("[%f, %f]", p.x, p.y)
}

type collisionCircle struct {
	center point
	r      float64
}

// r based on current speed.
func newCollCircle(pos point, vSpeed, hSpeed float64) collisionCircle {
	return collisionCircle{
		center: pos,
		r: Max(math.Abs(vSpeed), math.Abs(hSpeed)),
	}
}

func (co collisionCircle) isCollidingWithSurface(surface []point) (point, int, bool) {
	for i := range surface[1:] {
		start := surface[i]
		end := surface[i + 1]

		if end.x < (co.center.x - co.r) && start.x < (co.center.x - co.r) {
			// To far to be collision.
			continue
		}

		if end.x > (co.center.x + co.r) && start.x > (co.center.x + co.r) {
			// To far to be collision.
			break
		}

		// Fallback to proper intersection by finding the closest point and checking if it collide.
		startToEndVec := end.Sub(start)
		startToCenterVec := co.center.Sub(start)
		distanceFromStart := startToCenterVec.Dot(startToEndVec) / startToEndVec.Norm2()
		closestPt := point{
			x: start.x + startToEndVec.x * distanceFromStart,
			y: start.y + startToEndVec.y * distanceFromStart,
		}

		dist := co.center.Distance(closestPt)
		if dist < 100 {
			d("c: %v, closest: %v dist: %f", co.center.print(), closestPt.print(), dist)
		}
		if dist > co.r {
			continue
		}

		return closestPt, i+1, true

	}
	return point{}, 0, false
}

// Bezier evaluation.
func evalBezierXUsingHornerMethod(t float64, controlPoints []point) float64 {
	n := len(controlPoints) - 1
	u := float64(1 - t)
	bc := float64(1)
	tn := float64(1)
	tmp := controlPoints[0].x * u
	for i := 1; i < n; i++ {
		tn *= t
		bc *= float64(n-i+1) / float64(i)
		tmp = (tmp + tn*bc*controlPoints[i].x) * u
	}
	return (tmp + tn*t*controlPoints[n].x)
}

func evalBezierYUsingHornerMethod(t float64, controlPoints []point) float64 {
	n := len(controlPoints) - 1
	u := float64(1 - t)
	bc := float64(1)
	tn := float64(1)
	tmp := controlPoints[0].y * u
	for i := 1; i < n; i++ {
		tn *= t
		bc *= float64(n-i+1) / float64(i)
		tmp = (tmp + tn*bc*controlPoints[i].y) * u
	}
	return (tmp + tn*t*controlPoints[n].y)
}
