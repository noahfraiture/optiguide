package parser

import (
	"fmt"
	"strings"

	"github.com/tealeg/xlsx"
)

// TODO : change to array of string instead of simple string
// Currently this is string because of database library hard to understand
type Card struct {
	ID           int
	Level        int
	Info         string
	TaskOne      string
	TaskTwo      string
	Achievements string
	DungeonOne   string
	DungeonTwo   string
	DungeonThree string
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

func Parse(fileName string) ([]Card, error) {
	var err error
	wb, err := xlsx.OpenFile(fileName)
	if err != nil {
		return nil, err
	}
	progression, ok := wb.Sheet["Progression Optimisée"]
	if !ok {
		return nil, fmt.Errorf("Sheet not found")
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
			return nil, err
		}

		// Info supplementaire
		info, err := parseInfo(row, &prevInfo, &prevInfoCounter)
		if err != nil {
			return nil, err
		}

		cards = append(cards, Card{
			ID:           step_number,
			Level:        level,
			Info:         info,
			TaskOne:      row.Cells[6].Value,
			TaskTwo:      row.Cells[8].Value,
			Achievements: row.Cells[10].Value,
			DungeonOne:   row.Cells[12].Value,
			DungeonTwo:   row.Cells[14].Value,
			DungeonThree: row.Cells[16].Value,
			Spell:        row.Cells[18].Value,
		})
		step_number++
	}
	return cards, nil
}
