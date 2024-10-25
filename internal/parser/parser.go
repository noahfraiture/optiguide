package parser

import (
	"fmt"
	"strings"

	"github.com/tealeg/xlsx"
)

type Task struct {
	Title       string
	Description string
}

type Card struct {
	Step         int
	Level        int
	Info         string
	Tasks        [2]Task
	Achievements []string
	Dungeon1     []string
	Dungeon2     []string
	Dungeon3     []string
	Spell        string
}

// Level, first column
func parseLevel(row *xlsx.Row, prevLevel *int) (int, error) {
	var level int
	var err error
	if row.Cells[0].Value == "" {
		level = *prevLevel
	} else {
		level, err = row.Cells[0].Int()
		*prevLevel = level
	}
	return level, err
}

func parseInfo(row *xlsx.Row, prevInfo *string, prevInfoCounter *int) (string, error) {
	var info string
	var err error
	switch {
	case row.Cells[2].Value == "" && *prevInfoCounter == 0:
	case row.Cells[2].Value != "" && *prevInfoCounter == 0:
		info = row.Cells[2].Value
		*prevInfo = info
		*prevInfoCounter = row.Cells[2].VMerge + 1
	case row.Cells[2].Value == "" && *prevInfoCounter != 0:
		info = *prevInfo
		*prevInfoCounter--
	case row.Cells[2].Value != "" && *prevInfoCounter != 0:
		return "", fmt.Errorf("Cell not empty but overlap with previous")
	}
	return info, err
}

var globalCards []Card

func GetCard(i int) (Card, error) {
	if len(globalCards) == 0 {
		err := parse("guide.xlsx")
		if err != nil {
			return Card{}, err
		}
	}
	if i >= len(globalCards) {
		return Card{}, fmt.Errorf("Index out of range")
	}
	return globalCards[i], nil
}

func parse(fileName string) error {
	var err error
	wb, err := xlsx.OpenFile(fileName)
	if err != nil {
		return err
	}
	progression, ok := wb.Sheet["Progression Optimisée"]
	if !ok {
		return fmt.Errorf("Sheet not found")
	}

	cards := make([]Card, 0)

	var prevLevel int
	var prevInfo string
	var prevInfoCounter int // When row merged
	step_number := 0
	for _, row := range progression.Rows[4:] {
		if prevInfoCounter > 0 {
			prevInfoCounter--
		}

		left := strings.Trim(row.Cells[6].Value, " ↓\n")
		center := strings.Trim(row.Cells[7].Value, " ↓\n")
		right := strings.Trim(row.Cells[8].Value, " ↓\n")
		if center == "↓" || left == "" && right == "" {
			continue
		}

		level, err := parseLevel(row, &prevLevel)
		if err != nil {
			return err
		}

		// Info supplementaire
		info, err := parseInfo(row, &prevInfo, &prevInfoCounter)
		if err != nil {
			return err
		}

		// Schema Tasks
		var task1 Task
		if left != "" {
			taskOneText := strings.SplitN(row.Cells[6].Value, "\n", 2)
			if len(taskOneText) != 2 {
				taskOneText = append(taskOneText, "")
			}
			task1 = Task{
				Title:       taskOneText[0],
				Description: taskOneText[1],
			}
		}
		var task2 Task
		if right != "" {
			taskTwoText := strings.SplitN(row.Cells[8].Value, "\n", 2)
			if len(taskTwoText) != 2 {
				taskTwoText = append(taskTwoText, "")
			}
			task2 = Task{
				Title:       taskTwoText[0],
				Description: taskTwoText[1],
			}
		}
		tasks := [2]Task{task1, task2}

		// Succes concerne
		achievements := strings.Split((row.Cells[10].Value), "\n")

		// Tour du monde
		dungeon1 := strings.Split((row.Cells[12].Value), "\n")
		// Tornade de donjons
		dungeon2 := strings.Split((row.Cells[14].Value), "\n")
		// Autre donjons
		dungeon3 := strings.Split((row.Cells[16].Value), "\n")
		// Sorts communs
		spells := row.Cells[18].Value

		cards = append(cards, Card{
			Step:         step_number,
			Level:        level,
			Info:         info,
			Tasks:        tasks,
			Achievements: achievements,
			Dungeon1:     dungeon1,
			Dungeon2:     dungeon2,
			Dungeon3:     dungeon3,
			Spell:        spells,
		})
		step_number++
	}
	globalCards = cards
	return nil
}
