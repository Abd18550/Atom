package database

import (
	"backend/models"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type questionMeta struct {
	Title string `json:"title"`
	Type  string `json:"type"`
}

// loadAllQuestions reads the questions/ directory structure and returns
// a slice of StageQuestion models, correctly ordered and mapped to stages.
func loadAllQuestions(stages []models.Stage) []models.StageQuestion {
	questionsDir := "questions"
	var questions []models.StageQuestion

	// Read stage directories
	stageDirs, err := os.ReadDir(questionsDir)
	if err != nil {
		log.Fatalf("Failed to read questions directory: %v", err)
	}

	// Sort by name to ensure ordering (s01_, s02_, ...)
	sort.Slice(stageDirs, func(i, j int) bool {
		return stageDirs[i].Name() < stageDirs[j].Name()
	})

	stageIndex := 0
	for _, stageDir := range stageDirs {
		if !stageDir.IsDir() || !strings.HasPrefix(stageDir.Name(), "s") {
			continue
		}
		if stageIndex >= len(stages) {
			log.Printf("Warning: more stage directories than stages in DB, skipping %s", stageDir.Name())
			break
		}

		stagePath := filepath.Join(questionsDir, stageDir.Name())
		questionDirs, err := os.ReadDir(stagePath)
		if err != nil {
			log.Printf("Warning: failed to read stage directory %s: %v", stagePath, err)
			stageIndex++
			continue
		}

		// Sort by name to ensure ordering (q01_, q02_, ...)
		sort.Slice(questionDirs, func(i, j int) bool {
			return questionDirs[i].Name() < questionDirs[j].Name()
		})

		orderIndex := 1
		for _, qDir := range questionDirs {
			if !qDir.IsDir() || !strings.HasPrefix(qDir.Name(), "q") {
				continue
			}

			qPath := filepath.Join(stagePath, qDir.Name())

			// Read meta.json
			metaBytes, err := os.ReadFile(filepath.Join(qPath, "meta.json"))
			if err != nil {
				log.Printf("Warning: failed to read meta.json in %s: %v", qPath, err)
				continue
			}
			var meta questionMeta
			if err := json.Unmarshal(metaBytes, &meta); err != nil {
				log.Printf("Warning: failed to parse meta.json in %s: %v", qPath, err)
				continue
			}

			// Read description.md
			desc, err := os.ReadFile(filepath.Join(qPath, "description.md"))
			if err != nil {
				log.Printf("Warning: failed to read description.md in %s: %v", qPath, err)
				continue
			}

			// Read solution.txt
			solution, err := os.ReadFile(filepath.Join(qPath, "solution.txt"))
			if err != nil {
				log.Printf("Warning: failed to read solution.go in %s: %v", qPath, err)
				continue
			}

			// Read test.txt
			test, err := os.ReadFile(filepath.Join(qPath, "test.txt"))
			if err != nil {
				log.Printf("Warning: failed to read test.go in %s: %v", qPath, err)
				continue
			}

			questions = append(questions, models.StageQuestion{
				Title:         meta.Title,
				Description:   strings.TrimSpace(string(desc)),
				Type:          meta.Type,
				RightSolution: strings.TrimSpace(string(solution)),
				Test:          strings.TrimSpace(string(test)),
				StageID:       stages[stageIndex].ID,
				OrderIndex:    orderIndex,
			})
			orderIndex++
		}
		stageIndex++
	}

	return questions
}
