// Parses the dataset of LC questions given at
// https://www.kaggle.com/datasets/gzipchrist/leetcode-problem-dataset
// Parses it in objects to more easily import them at-will to the
// database, if it has no problem set at that time.

package main

import (
	"encoding/csv"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type CsvProblem struct {
	Id                 int
	Title              string
	Description        string
	Is_premium         bool
	Difficulty         int // Easy = 1, Medium = 2, Hard = 3
	Solution_link      string
	Acceptance_rate_th int // as 1/1000, so 1000 is 100%, and 573 is 57.3%
	Frequency          int // as 1/1000, so 1000 is 100%, and 573 is 57.3%
	Url                string
	Discuss_count      int
	Accepted           int
	Submissions        int
	Companies          []string
	Related_topics     []string
	Likes              int
	Dislikes           int
	Rating             int // as 1/1000, so 1000 is 100%, and 573 is 57.3%
	Asked_by_faang     bool
	Similar_questions  []string
}

func strColumn(newLine bool, name, text string) string {
	str := ""
	if newLine {
		str += "\n"
	}
	str += name + ": " + text

	return str
}

func (p *CsvProblem) String() string {
	str := "Problem " + strconv.Itoa(p.Id)

	str += strColumn(true, "Title", p.Title)
	desc := p.Description[:70]
	i := strings.Index(desc, "\n\n")
	if i > 0 {
		desc = desc[:i]
	}
	// Remove trailing newlines
	// for strings.HasSuffix(desc, "\n") {
	// 	desc = desc[:len(desc)-1]
	// }
	str += strColumn(true, "Description", desc)
	diff := "Easy"
	if p.Difficulty == 2 {
		diff = "Medium"
	} else if p.Difficulty == 3 {
		diff = "Hard"
	}
	str += strColumn(true, "Difficulty", diff)
	str += strColumn(true, "URL", p.Url)

	return str
}

func parseProblems(rows [][]string) []CsvProblem {
	problems := make([]CsvProblem, 0, len(rows))

	for _, row := range rows {
		if len(row) < 19 {
			//continue // skip incomplete rows
			panic("What")
		}

		id, _ := strconv.Atoi(row[0])
		isPremium, _ := strconv.ParseBool(row[3])
		difficulty_str := row[4]
		difficulty := 0
		if difficulty_str == "Easy" {
			difficulty = 1
		} else if difficulty_str == "Medium" {
			difficulty = 2
		} else if difficulty_str == "Hard" {
			difficulty = 3
		} else {
			panic("Got " + difficulty_str + " where one from [Easy, Medium, Hard] expected")
		}

		acceptanceRate, _ := strconv.Atoi(row[6])
		frequency, _ := strconv.Atoi(row[7])
		discussCount, _ := strconv.Atoi(row[9])
		accepted, _ := strconv.Atoi(row[10])
		submissions, _ := strconv.Atoi(row[11])
		likes, _ := strconv.Atoi(row[14])
		dislikes, _ := strconv.Atoi(row[15])
		rating, _ := strconv.Atoi(row[16])
		askedByFaang, _ := strconv.ParseBool(row[17])

		companies := strings.Split(row[12], ",")
		relatedTopics := strings.Split(row[13], ",")

		// Parse similar questions
		similarRaw := row[18]
		similar := []string{}
		matches := regexp.MustCompile(`\[([^\]]+)\]`).FindAllStringSubmatch(similarRaw, -1)
		for _, match := range matches {
			similar = append(similar, match[1]) // e.g., "3Sum, /problems/3sum/, Medium"
		}

		problem := CsvProblem{
			Id:                 id,
			Title:              row[1],
			Description:        row[2],
			Is_premium:         isPremium,
			Difficulty:         difficulty,
			Solution_link:      row[5],
			Acceptance_rate_th: acceptanceRate,
			Frequency:          frequency,
			Url:                row[8],
			Discuss_count:      discussCount,
			Accepted:           accepted,
			Submissions:        submissions,
			Companies:          companies,
			Related_topics:     relatedTopics,
			Likes:              likes,
			Dislikes:           dislikes,
			Rating:             rating,
			Asked_by_faang:     askedByFaang,
			Similar_questions:  similar,
		}

		problems = append(problems, problem)
	}

	return problems
}

// Expects a csv file like the one in
// https://www.kaggle.com/datasets/gzipchrist/leetcode-problem-dataset
func ParseProblemsFromReader(r io.Reader) []CsvProblem {
	csv_r := csv.NewReader(r)

	values, err := csv_r.ReadAll()
	if err != nil {
		panic(err)
	}

	problems := parseProblems(values[1:])

	return problems
}
