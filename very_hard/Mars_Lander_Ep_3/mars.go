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

func de(format string, a ...interface{}) {
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
	l.gatherInput()
	d("Landing center: %s, tolerance: %d", l.landingCenterPoint.print(), l.landingSiteTolerance)

	for {
		// MAIN LOOP.
		where, isLandingArea, when, eVSpeed, eHSpeed := l.estimateSurfaceReachable()
		d("Estimated landing: %s | ok? %v | epochs: %d, eV: %f, eH %f",
			where.print(), isLandingArea, when, eVSpeed, eHSpeed)

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

	safeZone := collisionCircle{center: l.pos}
	for {
		// Movement.
		safeZone.center.x += tmpHSpeed
		safeZone.center.y += tmpVSpeed

		safeZone.r = Max(tmpHSpeed, tmpVSpeed) + 2
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

//func (p point) didCrossedSurface(surface []point) bool {
//	for i, surfacePt := range surface {
//		if p.x > surfacePt.x {
//			continue
//		}
//
//		a := surfacePt.y - surface[i-1].y
//		b := surfacePt.x - surface[i-1].x
//		a2 := p.x - surface[i-1].x
//		x := (b * a2) / a
//		if (p.y - surface[i-1].y) <= x {
//			return true
//		}
//	}
//
//	return false
//}

func (p point) print() string {
	return fmt.Sprintf("[%f, %f]", p.x, p.y)
}

type collisionCircle struct {
	center point
	r      float64
}

func (co collisionCircle) isCollidingWithSurface(surface []point) (point, int, bool) {
	for i := range surface[1:] {
		start := surface[i]
		end := surface[i+1]

		if end.x < (co.center.x-co.r) && start.x < (co.center.x-co.r) {
			// To far to be collision.
			continue
		}

		// Fallback to proper intersection.
		d := end.Sub(start)
		f := co.center.Sub(start)

		a := d.Dot(d)
		b := 2 * f.Dot(d)
		c := f.Dot(f) - co.r*co.r

		discriminant := b*b - 4*a*c

		if discriminant < 0 {
			// No intersection.
			continue
		}

		// There is a solution to the equation.
		discriminant = math.Sqrt(discriminant)
		t1 := (-b - discriminant) / (2 * a)
		t2 := (-b + discriminant) / (2 * a)

		// 3x HIT cases:
		//          -o->             --|-->  |            |  --|->
		// Impale(t1 hit,t2 hit), Poke(t1 hit,t2>1), ExitWound(t1<0, t2 hit),

		// 3x MISS cases:
		//       ->  o                     o ->              | -> |
		// FallShort (t1>1,t2>1), Past (t1<0,t2<0), CompletelyInside(t1<0, t2>1)
		de("%f x %f", t1, t2)
		if t1 >= 0 && t1 <= 1 {
			// t1 is the intersection, and it's closer than t2 (since t1 uses -b - discriminant)
			// Impale, Poke
			return calcIntersectPt(t1, start, d), i+1, true
		}

		// here t1 didn't intersect so we are either started inside the sphere or completely past it
		if t2 >= 0 && t2 <= 1 {
			// ExitWound
			return calcIntersectPt(t2, start, d), i+1, true
		}
		// no intn: FallShort, Past, CompletelyInside
	}
	return point{}, 0, false
}

func calcIntersectPt(t float64, start, d point) point {
	return point{
		x: start.x + t*d.x,
		y: start.y + t*d.y,
	}
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
