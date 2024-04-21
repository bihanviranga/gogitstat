package main

import (
  "gopkg.in/src-d/go-git.v4"
  "gopkg.in/src-d/go-git.v4/plumbing/object"
)

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

// Calculates and prints stats.
func stats(email string) {
  commits := processRepositories(email)
  printCommitsStats(commits)
}
