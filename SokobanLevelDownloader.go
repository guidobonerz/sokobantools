package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

type SokobanLevels struct {
	XMLName         xml.Name `xml:"SokobanLevels"`
	Text            string   `xml:",chardata"`
	Xsi             string   `xml:"xsi,attr"`
	SchemaLocation  string   `xml:"schemaLocation,attr"`
	Title           string   `xml:"Title"`
	Description     string   `xml:"Description"`
	Email           string   `xml:"Email"`
	LevelCollection struct {
		Text      string `xml:",chardata"`
		Copyright string `xml:"Copyright,attr"`
		MaxWidth  int    `xml:"MaxWidth,attr"`
		MaxHeight int    `xml:"MaxHeight,attr"`
		Level     []struct {
			Text   string   `xml:",chardata"`
			ID     string   `xml:"Id,attr"`
			Width  int      `xml:"Width,attr"`
			Height int      `xml:"Height,attr"`
			L      []string `xml:"L"`
		} `xml:"Level"`
	} `xml:"LevelCollection"`
}

func buildData(levels *SokobanLevels) {
	for _, level := range levels.LevelCollection.Level {

		if level.Width < 16 && level.Height < 16 {
			var levelData = make([]uint8, 0)
			var walls = make([]uint8, 0)
			var goals = make([]uint8, 0)
			var boxes = make([]uint8, 0)
			var player = make([]uint8, 0)
			for i, line := range level.L {
				fmt.Println(line)

				examineRow(alignRow(line, level.Width), uint8(i), &walls, &goals, &boxes, &player)
			}
			levelData = append(levelData, uint8(len(walls)))
			levelData = append(levelData, walls[:]...)
			levelData = append(levelData, uint8(len(boxes)))
			levelData = append(levelData, boxes[:]...)
			levelData = append(levelData, goals[:]...)
			levelData = append(levelData, player[:]...)
			fmt.Println(levelData)
			var dataLine string = "1000 DATA "
			for i, v := range levelData {
				dataLine += strconv.Itoa(int(v))
				if i < len(levelData)-1 {
					dataLine += ","
				}
			}
			dataLine += "\n1001 DATA -1"
			fmt.Println(dataLine)
			fmt.Println("--------------------------------")
		}
	}
}

func alignRow(line string, width int) string {
	for i := 0; i <= width-len(line); i++ {
		line += " "
	}
	return line
}

func examineRow(row string, rowNum uint8, walls *[]uint8, goals *[]uint8, boxes *[]uint8, player *[]uint8) {
	var b uint8 = 0
	var v uint8 = 128
	for i, c := range row {
		var value = ((uint8(i) << 4) | (rowNum & 15))
		switch c {
		case '.': // goal
			*goals = append(*goals, value)
		case '*': // box on goal
			*boxes = append(*boxes, value)
			*goals = append(*goals, value)
		case '$': // box
			*boxes = append(*boxes, value)
		case '@': // player
			*player = append(*player, value)
		case '+': // player on goal
			*player = append(*player, value)
			*goals = append(*goals, value)
		case ' ': // floor
		}
		if i > 0 && i%8 == 0 {
			*walls = append(*walls, b)
			b = 0
			v = 128
		}
		if c == '#' {
			b |= v
		}
		v >>= 1
	}
	*walls = append(*walls, b)
}

func main() {
	resp, err := http.Get("http://www.sourcecode.se/sokoban/levels")
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	sb := string(body)

	mainPattern := "(\\?act=info&id=[0-9]*&nr=[0-9]*)"
	downloadPattern := "(/sokoban/download/([a-zA-Z0-9_-]*)\\.slc)"

	re, err := regexp.Compile(mainPattern)
	results := re.FindAllString(sb, -1)
	for i, v := range results {
		fmt.Printf("MainMatch %d: %s\n", i, v)
		resp, err := http.Get("http://www.sourcecode.se/sokoban/levels" + v)
		body, err := ioutil.ReadAll(resp.Body)
		sb := string(body)

		re, err := regexp.Compile(downloadPattern)
		results := re.FindAllString(sb, -1)
		r := re.FindStringSubmatch(sb)
		fmt.Println(r)

		for j, v := range results {
			fmt.Printf("SubMatch %d: %s\n", j, v)
			resp, err := http.Get("http://www.sourcecode.se/" + v)
			if err != nil {
				log.Fatalln(err)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			}
			fileName := fmt.Sprintf("%s.xml", r[2])
			fmt.Println(fileName)
			err = ioutil.WriteFile(fileName, body, 0644)

			sokobanLevels := &SokobanLevels{}

			_ = xml.Unmarshal([]byte(body), &sokobanLevels)

			//buildData(sokobanLevels)
			if err != nil {
				log.Fatalln(err)
			}
		}

		if err != nil {
			log.Fatalln(err)
		}
	}

}
