package main

import (
  "gopkg.in/src-d/go-git.v4"
  "gopkg.in/src-d/go-git.v4/plumbing/object"
)

type column []int

const daysInLastSixMonths = 183
const outOfRange = 99999

// Given a time.Time, returns the same date but with time set to 00:00:00.
func getBeginningOfDay(t time.Time) time.Time {
  year, month, day := t.Date()
  startOfDay := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
  return startOfDay
}

// Returns how many days have passed since the given date
func countDaysSinceDate(date time.Time) int {
  days := 0
  now := getBeginningOfDay(time.Now())
  for date.Before(now) {
    date = date.Add(time.Hour * 24)
    days++
    if days > daysInLastSixMonths {
      return outOfRange
    }
  }
  return days
}

// Determines the amount of days needed to fill the last row of the stats graph.
// TODO: Start the week with Sunday or Monday?
func calcOffset() int {
  var offset int
  weekday := time.Now().Weekday()

  switch weekday {
  case time.Sunday:
    offset = 7
  case time.Monday:
    offset = 6
  case time.Tuesday:
    offset = 5
  case time.Wednesday:
    offset = 4
  case time.Thursday:
    offset = 3
  case time.Friday:
    offset = 2
  case time.Saturday:
    offset = 1
  }

  return offset
}

func fillCommits(email string, path string, commits map[int]int) map[int]int {
  // Create a git repo object
  repo, err := git.PlainOpen(path)
  if err != nil {
    panic(err)
  }

  // Get the HEAD reference
  ref, err := repo.Head()
  if err != nil {
    panic(err)
  }

  // Get the commit history starting from HEAD
  iterator, err := repo.Log(&git.LogOptions{From: ref.Hash()})
  if err != nil {
    panic(err)
  }

  // Iterate over the commits
  offset := calcOffset()
  err = iterator.ForEach(func(c *object.Commit) error {
    daysAgo := countDaysSinceDate(c.Author.When) + offset
    if c.Author.Email != email {
      return nil
    }
    if daysAgo != outOfRange {
      commits[daysAgo]++
    }

    return nil
  })
  if err != nil {
    panic(err)
  }

  return commits
}

// Given an email, returns the commits made in last 6 months.
func processRepositories(email string) map[int]int {
  filePath := getDotfilePath()
  repos := parseFileLinesToSlice(filePath)
  daysInMap := daysInLastSixMonths

  commits := make(map[int]int, daysInMap)
  for i := daysInMap; i > 0; i-- {
    commits[i] = 0
  }

  for _, path := range repos {
    commits = fillCommits(email, path, commits)
  }

  return commits
}

// Returns a slice of indexes of a map, ordered.
func sortMapIntoSlice(m map[int]int) []int {
  var keys []int
  for k := range m {
    keys = append(keys, k)
  }
  sort.Ints(keys)
  return keys
}

// Generates a map with rows and columns with commit info
func buildCols(keys []int, commits map[int]int) map[int]column {
  cols := make(map[int]column)
  col := column{}

  for _, k := range keys {
    week := int(k/7)    // 26, 25, .., 1
    dayInWeek := k % 7  // 0, 1, 2, .., 6

    if dayInWeek == 0 {
      col = column{}
    }

    col = append(col, commits[k])

    if dayInWeek == 6 {
      cols[week] = col
    }
  }

  return cols
}

// Prints the cells.
func printCells(cols map[int]column) {
  printMonths()
  for j := 6; j >= 0; j-- {
    for i := weeksInLastSixMonths + 1; i >= 0; i-- {
      if i == weeksInLastSixMonths + 1 {
        printDayCol(j)
      }
      if col, ok := cols[i]; ok {
        if i == 0 && j == calcOffset() - 1 {
          printCell(col[j], true)
          continue
        } else {
          if len(col) > j {
            printCell(col[j], false)
            continue
          }
        }
      }
      printCell(0, false)
    }
    fmt.Printf("\n")
  }
}

// Prepare the commit data for printing to the screen.
func printCommitStats(commits map[int]int) {
  keys := sortMapIntoSlice(commits)
  cols := buildCols(keys, commits)
  printCells(cols)
}

// Entry point for the stats calculation and printing.
func stats(email string) {
  commits := processRepositories(email)
  printCommitStats(commits)
}
