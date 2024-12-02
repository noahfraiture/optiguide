package parser

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/tealeg/xlsx/v3"
)

type Achievement struct {
	Value string
	Link  string
}

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
	Achievements   []Achievement // Different row
	DungeonOne     []string      // Same row but cleaner if it follows same pattern
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

const CHECKBOXTEXT = "FALSE"

func newCard(cardCounter int) *Card {
	return &Card{
		ID:           uuid.New(),
		Idx:          cardCounter,
		Achievements: []Achievement{},
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
	for i := 9; i < 102; i++ {
		row, err := progression.Row(i)
		if err != nil {
			return nil, err
		}
		left := strings.Trim(row.GetCell(ARROWLEFT).String(), " \n")
		center := strings.Trim(row.GetCell(ARROWCENTER).String(), " \n")
		right := strings.Trim(row.GetCell(ARROWRIGHT).String(), " \n")
		if center == "↓" || (left == "↓" && right == "↓") {
			continue
		}

		// Should trigger on first iteration
		if row.GetCell(TASKTITLEONE).String() != "" &&
			row.GetCell(TASKTITLEONE).String() != CHECKBOXTEXT &&
			card != nil {
			cards = append(cards, *card)
			cardCounter++
			card = newCard(cardCounter)
		}

		if card == nil {
			card = newCard(cardCounter)
		}

		if row.GetCell(LEVEL).String() != "" {
			prevLevel = strings.ReplaceAll(row.GetCell(LEVEL).String(), ".0", "")
		}
		card.Level = prevLevel

		if row.GetCell(INFO).String() != "" {
			prevInfo = row.GetCell(INFO).String()
			prevInfoCounter = row.GetCell(INFO).VMerge
			// We must have this line because of there's no merge, the value isn't 1 but null
			card.Info = prevInfo
		}
		if prevInfoCounter > 0 {
			card.Info = prevInfo
		}
		prevInfoCounter--

		if row.GetCell(TASKTITLEONE).String() != "" &&
			row.GetCell(TASKTITLEONE).String() != CHECKBOXTEXT {
			card.TaskTitleOne = row.GetCell(TASKTITLEONE).String()
			card.TaskTitleTwo = row.GetCell(TASKTITLETWO).String()
		}

		if row.GetCell(TASKCONTENTONE).String() != "" {
			card.TaskContentOne = row.GetCell(TASKCONTENTONE).String()
		}
		if row.GetCell(TASKCONTENTTWO).String() != "" {
			card.TaskContentTwo = row.GetCell(TASKCONTENTTWO).String()
		}

		if row.GetCell(ACHIEVEMENTNAME).String() != "" {
			achievement := Achievement{
				Value: row.GetCell(ACHIEVEMENTNAME).String(),
				Link:  row.GetCell(ACHIEVEMENTNAME).Hyperlink.Link,
			}
			card.Achievements = append(card.Achievements, achievement)
		}

		if strings.Trim(row.GetCell(DUNGEONSONE).String(), " -\n\t") != "" {
			card.DungeonOne = append(
				card.DungeonOne,
				strings.Trim(row.GetCell(DUNGEONSONE).String(), " -\n\t"),
			)
		}
		if strings.Trim(row.GetCell(DUNGEONSTWO).String(), " -\n\t") != "" {
			card.DungeonTwo = append(
				card.DungeonTwo,
				strings.Trim(row.GetCell(DUNGEONSTWO).String(), " -\n\t"),
			)
		}
		if strings.Trim(row.GetCell(DUNGEONSTHREE).String(), " -\n\t") != "" {
			card.DungeonThree = append(
				card.DungeonThree,
				strings.Trim(row.GetCell(DUNGEONSTHREE).String(), " -\n\t"),
			)
		}

		if row.GetCell(SPELLS).String() != "" {
			card.Spell = row.GetCell(SPELLS).String()
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
