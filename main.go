package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kardianos/osext"
	"github.com/pkg/term"
)

type keyCommand int

const (
	forwardsCMD    keyCommand = 1
	turnLeftCMD    keyCommand = 2
	turnRightCMD   keyCommand = 3
	turnArroundCMD keyCommand = 4
)

var (
	gpx          string
	currentDir   string
	longitudeStr string
	longitude    float64
	latitudeStr  string
	latitude     float64
	heading      float64
)

func main() {

	heading = 90
	exec, err := osext.Executable()
	if err != nil {
		fmt.Println(err)
	}
	currentDir = strings.TrimSuffix(exec, "walker")
	gpx = fmt.Sprintf("%slocation/current.gpx", currentDir)

	err = readLocationFromGPX()
	if err != nil {
		fmt.Println(err)
		return
	}
	longitude, _ = strconv.ParseFloat(longitudeStr, 64)
	latitude, _ = strconv.ParseFloat(latitudeStr, 64)

	fmt.Println("Use the WASD keys to move arround, have fun!")

	lastCommand := time.Now()

	var command keyCommand
	for {
		c := getch()
		switch string(c) {
		case "w":
			command = forwardsCMD
		case "a":
			command = turnLeftCMD
		case "s":
			command = turnArroundCMD
		case "d":
			command = turnRightCMD

		case "W":
			command = forwardsCMD
		case "A":
			command = turnLeftCMD
		case "S":
			command = turnArroundCMD
		case "D":
			command = turnRightCMD

		case "q":
			return
		}

		now := time.Now().Add(-time.Millisecond * 500)
		if now.After(lastCommand) {
			go dispatch(command)
			lastCommand = time.Now()
		}
	}
}

func dispatch(cmd keyCommand) {
	switch cmd {
	case forwardsCMD:
		w()
	case turnLeftCMD:
		a()
	case turnArroundCMD:
		s()
	case turnRightCMD:
		d()
	}
}

func w() {
	move(0.01)
}

func a() {
	turnLeft()
	move(0.001)
}

func s() {
	turnArround()
	move(0.001)
}

func d() {
	turnRight()
	move(0.001)
}

func move(distance float64) {

	latitude, longitude = destination(latitude, longitude, heading, distance)

	latitudeStr := strconv.FormatFloat(latitude, 'f', -1, 64)
	longitudeStr := strconv.FormatFloat(longitude, 'f', -1, 64)
	save(latitudeStr, longitudeStr)
}

func turnArround() {
	heading += 180
	if heading > 360 {
		heading -= 360
	}
}

func turnLeft() {
	heading -= 10
	if heading < 0 {
		heading += 360
	}
}

func turnRight() {
	heading += 10
	if heading > 360 {
		heading -= 360
	}
}

func save(latitude string, longitude string) {

	fileStr := fmt.Sprintf("<?xml version=\"1.0\"?>\n<gpx version=\"1.1\" creator=\"Xcode\">\n    <wpt lat=\"%s\" lon=\"%s\">\n    </wpt>\n</gpx>", latitude, longitude)
	fileBytes := []byte(fileStr)

	os.Remove(gpx)

	var file, err = os.Create(gpx)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	err = ioutil.WriteFile(gpx, fileBytes, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

	update()
}

func update() {

	cmd := exec.Command("osascript", currentDir+"updateXcodeLocSim.app")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()

}

func loadGPX() (*string, error) {

	if _, err := os.Stat(gpx); err != nil {
		return nil, err
	}

	b, err := ioutil.ReadFile(gpx)
	if err != nil {
		return nil, err
	}

	gpxStr := string(b)

	return &gpxStr, nil

}

func readLocationFromGPX() error {

	gpxStr, err := loadGPX()
	if err != nil {
		return err
	}

	latRegExp, err := regexp.Compile(`lat="(.*?)"`)
	lonRegExp, err := regexp.Compile(`lon="(.*?)"`)
	numberRegExp, err := regexp.Compile(`"(.*?)"`)
	if err != nil {
		return err
	}
	latitudeStr = latRegExp.FindString(*gpxStr)
	latitudeStr = numberRegExp.FindString(latitudeStr)
	latitudeStr = strings.Replace(latitudeStr, `"`, "", -1)
	longitudeStr = lonRegExp.FindString(*gpxStr)
	longitudeStr = numberRegExp.FindString(longitudeStr)
	longitudeStr = strings.Replace(longitudeStr, `"`, "", -1)

	return nil
}

func getch() []byte {
	t, _ := term.Open("/dev/tty")
	term.RawMode(t)
	bytes := make([]byte, 3)
	numRead, err := t.Read(bytes)
	t.Restore()
	t.Close()
	if err != nil {
		return nil
	}
	return bytes[0:numRead]
}

func destination(latitude float64, longitude float64, bearing float64, distance float64) (float64, float64) {
	distanceRadians := distance / 6371.0
	bearingRadians := degreesToRadians(bearing)
	fromLatRadians := degreesToRadians(latitude)
	fromLonRadians := degreesToRadians(longitude)

	toLatRadians := math.Asin(math.Sin(fromLatRadians)*math.Cos(distanceRadians) + math.Cos(fromLatRadians)*math.Sin(distanceRadians)*math.Cos(bearingRadians))

	toLonRadians := fromLonRadians + math.Atan2(math.Sin(bearingRadians)*math.Sin(distanceRadians)*math.Cos(fromLatRadians), math.Cos(distanceRadians)-math.Sin(fromLatRadians)*math.Sin(toLatRadians))

	lat := radiansToDegrees(toLatRadians)
	long := radiansToDegrees(toLonRadians)

	for long > 180 {
		long -= 180
	}
	for long < -180 {
		long += 180
	}

	return lat, long
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

func radiansToDegrees(radians float64) float64 {
	return radians * 180.0 / math.Pi
}
