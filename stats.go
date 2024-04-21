package main

import (
  "fmt"
  "sort"
  "time"

  "gopkg.in/src-d/go-git.v4"
  "gopkg.in/src-d/go-git.v4/plumbing/object"
)

type column []int

const daysInLastSixMonths = 183
const outOfRange = 99999
const weeksInLastSixMonths = 26

// Given a time.Time, returns the same date but with time set to 00:00:00.
func getBeginningOfDay(t time.Time) time.Time {
  year, month, day := t.Date()
  startOfDay := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
  return startOfDay
}

// Returns how many days have passed since the given date.
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

// Populates a map with commit data for a given author, for a repo at
// the given path.
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

// Function for testing purposes.
// TODO move this to a separate file.
func printIntIntMap(m map[int]int) {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	fmt.Println("Key\tValue")
	fmt.Println("---------------")
	for _, k := range keys {
		fmt.Printf("%d\t%d\n", k, m[k])
	}
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

// Generates a map with rows and columns with commit info.
func buildCols(keys []int, commits map[int]int) map[int]column {
  cols := make(map[int]column)
  col := column{}

  for _, k := range keys {
    week := int(k/7)    // 26, 25, .., 1
    dayInWeek := k % 7  // 0, 1, 2, .., 6

    // Start a new column
    if dayInWeek == 0 {
      col = column{}
    }

    col = append(col, commits[k])

    if dayInWeek == 6 {
      // HACK: must figure out the issue and fix this
      if len(col) < 7 {
        col = append(col, 0)
      }
      // END HACK
      cols[week] = col
    }
  }

  return cols
}

// Prints the month names in the first line.
func printMonths() {
  week := getBeginningOfDay(time.Now()).Add(-(daysInLastSixMonths * time.Hour * 24))
  month := week.Month()
  fmt.Printf("        ")
  for {
    if week.Month() != month {
      fmt.Printf("%s", week.Month().String()[:3])
      month = week.Month()
    } else {
      fmt.Printf("    ")
    }

    week = week.Add(7 * time.Hour * 24)
    if week.After(time.Now()) {
      break
    }
  }
  fmt.Printf("\n")
}

// Given the day number (starting from 0), prints the day name.
func printDayCol(day int) {
  out := "     "
  switch day {
    case 1:
      out = " Mon "
    case 3:
      out = " Wed "
    case 5:
      out = " Fri "
  }

  fmt.Printf(out)
}

// Prints the formatted cell data.
func printCell(val int, today bool) {
  escape := "\033[0;37;30m"
  switch {
    case val > 0 && val < 5:
      escape = "\033[1;30;47m"
    case val >= 5 && val < 10:
      escape = "\033[1;30;43m"
    case val >= 10:
      escape = "\033[1;30;42m"
  }

  if today {
    escape = "\033[1;37;45m"
  }

  if val == 0 {
    fmt.Printf(escape + "  - " + "\033[0m")
    return
  }

  str := "  %d "
  switch {
    case val >= 10:
      str = " %d "
    case val >= 100:
      str = "%d "
  }

  fmt.Printf(escape + str + "\033[0m", val)
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
        // Handling 'today'
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
