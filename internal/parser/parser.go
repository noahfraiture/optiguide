package parser

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/tealeg/xlsx"
)

// TODO : change to array of string instead of simple string
// Currently this is string because of database library hard to understand
type Card struct {
	ID             uuid.UUID
	Idx            int
	Level          string
	Info           string // NOTE : we could move that in a different sql table
	TaskTitleOne   string
	TaskTitleTwo   string
	TaskContentOne string
	TaskContentTwo string
	Achievements   []string // Different row
	DungeonOne     []string // Same row but cleaner if it follows same pattern
	DungeonTwo     []string
	DungeonThree   []string
	Spell          string
}

const (
	LEVEL               = 0
	INFO                = 2
	TASKCHECKBOX        = 4
	TASKTITLEONE        = 4
	TASKTITLETWO        = 8
	TASKCONTENTONE      = 6
	TASKCONTENTTWO      = 8
	ARROWLEFT           = 6
	ARROWCENTER         = 7
	ARROWRIGHT          = 8
	ACHIEVEMENTCHECKBOX = 11
	ACHIEVEMENTNAME     = 12
	DUNGEONSONE         = 14
	DUNGEONSTWO         = 16
	DUNGEONSTHREE       = 18
	SPELLS              = 20
)

func newCard(cardCounter int) *Card {
	return &Card{
		ID:           uuid.New(),
		Idx:          cardCounter,
		Achievements: []string{},
		DungeonOne:   []string{},
		DungeonTwo:   []string{},
		DungeonThree: []string{},
	}
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
	cardCounter := 0

	var prevInfoCounter int // When row merged. Can cross multiple cards
	var prevInfo string
	var prevLevel string

	var card *Card
	for _, row := range progression.Rows[10:102] {
		left := strings.Trim(row.Cells[ARROWLEFT].Value, " \n")
		center := strings.Trim(row.Cells[ARROWCENTER].Value, " \n")
		right := strings.Trim(row.Cells[ARROWRIGHT].Value, " \n")
		if center == "↓" || (left == "↓" && right == "↓") {
			continue
		}

		// Should trigger on first iteration
		if row.Cells[TASKTITLEONE].Value != "" &&
			row.Cells[TASKTITLEONE].Value != "0" &&
			card != nil {
			cards = append(cards, *card)
			cardCounter++
			card = newCard(cardCounter)
		}

		if card == nil {
			card = newCard(cardCounter)
		}

		if row.Cells[LEVEL].Value != "" {
			prevLevel = strings.ReplaceAll(row.Cells[LEVEL].Value, ".0", "")
		}
		card.Level = prevLevel

		if row.Cells[INFO].Value != "" {
			prevInfo = row.Cells[INFO].Value
			prevInfoCounter = row.Cells[INFO].VMerge
			// We must have this line because of there's no merge, the value isn't 1 but null
			card.Info = prevInfo
		}
		if prevInfoCounter > 0 {
			card.Info = prevInfo
		}
		prevInfoCounter--

		if row.Cells[TASKTITLEONE].Value != "" && row.Cells[TASKTITLEONE].Value != "0" {
			card.TaskTitleOne = row.Cells[TASKTITLEONE].Value
			card.TaskTitleTwo = row.Cells[TASKTITLETWO].Value
		}

		if row.Cells[TASKCONTENTONE].Value != "" {
			card.TaskContentOne = row.Cells[TASKCONTENTONE].Value
		}
		if row.Cells[TASKCONTENTTWO].Value != "" {
			card.TaskContentTwo = row.Cells[TASKCONTENTTWO].Value
		}

		if row.Cells[ACHIEVEMENTNAME].Value != "" {
			card.Achievements = append(card.Achievements, row.Cells[ACHIEVEMENTNAME].Value)
		}

		if strings.Trim(row.Cells[DUNGEONSONE].Value, " -\n\t") != "" {
			card.DungeonOne = append(card.DungeonOne, strings.Trim(row.Cells[DUNGEONSONE].Value, " -\n\t"))
		}
		if strings.Trim(row.Cells[DUNGEONSTWO].Value, " -\n\t") != "" {
			card.DungeonTwo = append(card.DungeonTwo, strings.Trim(row.Cells[DUNGEONSTWO].Value, " -\n\t"))
		}
		if strings.Trim(row.Cells[DUNGEONSTHREE].Value, " -\n\t") != "" {
			card.DungeonThree = append(card.DungeonThree, strings.Trim(row.Cells[DUNGEONSTHREE].Value, " -\n\t"))
		}

		if row.Cells[SPELLS].Value != "" {
			card.Spell = row.Cells[SPELLS].Value
		}
	}
	return cards, nil
}

func (c Card) prettyPrint() string {
	var builder strings.Builder

	builder.WriteString("Card Details:\n")
	builder.WriteString(fmt.Sprintf("ID: %s\n", c.ID))
	builder.WriteString(fmt.Sprintf("Index: %d\n", c.Idx))
	builder.WriteString(fmt.Sprintf("Level: %s\n", c.Level))
	builder.WriteString(fmt.Sprintf("Info: %s\n", c.Info))
	builder.WriteString(fmt.Sprintf("Task Titles: %s, %s\n", c.TaskTitleOne, c.TaskTitleTwo))
	builder.WriteString(fmt.Sprintf("Task Contents: %s\n", c.TaskContentOne))
	builder.WriteString(fmt.Sprintf("             : %s\n", c.TaskContentTwo))

	// Format achievements
	if len(c.Achievements) > 0 {
		builder.WriteString("Achievements:\n")
		for i, achievement := range c.Achievements {
			builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, achievement))
		}
	}

	// Format dungeons
	if len(c.DungeonOne) > 0 || len(c.DungeonTwo) > 0 || len(c.DungeonThree) > 0 {
		builder.WriteString("Dungeons:\n")
		formatDungeon := func(name string, dungeon []string) {
			if len(dungeon) > 0 {
				builder.WriteString(fmt.Sprintf("  %s:\n", name))
				for i, item := range dungeon {
					builder.WriteString(fmt.Sprintf("    %d. %s\n", i+1, item))
				}
			}
		}
		formatDungeon("Dungeon One", c.DungeonOne)
		formatDungeon("Dungeon Two", c.DungeonTwo)
		formatDungeon("Dungeon Three", c.DungeonThree)
	}

	builder.WriteString(fmt.Sprintf("Spell: %s\n", c.Spell))

	return builder.String()
}
