package main

import (
  "bufio"
  "fmt"
  "io"
  "io/ioutil"
  "log"
  "os"
  "os/user"
  "strings"
)

// Check if a slice contains a given value.
func sliceContains(slice []string, value string) bool {
  for _, v := range slice {
    if v == value {
      return true
    }
  }

  return false
}

// Adds the items from 'new' to 'existing' without causing duplicates.
func joinSlices(new []string, existing []string) []string {
  for _, i := range new {
    if !sliceContains(existing, i) {
      existing = append(existing, i)
    }
  }

  return existing
}

// Opens a file at the given path.
// Creates if it doesn't exist.
func openFile(filePath string) *os.File {
  f, err := os.OpenFile(filePath, os.O_APPEND|os.O_RDWR, 0755)
  if err != nil {
    if os.IsNotExist(err) {
      f, err := os.Create(filePath)
      if err != nil {
        panic(err)
      }
      return f
    } else {
      panic(err)
    }
  }

  return f
}

// Parses the content of a file at a given path and return them
// as a slice.
func parseFileLinesToSlice(filePath string) []string {
  f := openFile(filePath)
  defer f.Close()

  var lines []string
  scanner := bufio.NewScanner(f)
  for scanner.Scan() {
    lines = append(lines, scanner.Text())
  }

  if err := scanner.Err(); err != nil {
    if err != io.EOF {
      panic(err)
    }
  }

  return lines
}

// Writes the content in given slice to lines on a file.
func dumpStringsSliceToFile(repos []string, filePath string) {
  content := strings.Join(repos, "\n")
  ioutil.WriteFile(filePath, []byte(content), 0755)
}

// Adds the new repos to the file.
func addNewSliceElementsToFile(filePath string, newRepos []string) {
  existingRepos := parseFileLinesToSlice(filePath)
  repos := joinSlices(newRepos, existingRepos)
  dumpStringsSliceToFile(repos, filePath)
}

// Returns a list of subdirectories of 'directory' ending with '.git'.
// The returned directories are the base directories of a repo, which means
// they contain a '.git' directory.
func scanGitDirectories(directories []string, directory string) []string {
  // trim the trailing slash '/'
  directory = strings.TrimSuffix(directory, "/")

  f, err := os.Open(directory)
  if err != nil {
    log.Fatal(err)
    // TODO quit here?
  }

  files, err := f.Readdir(-1)
  f.Close()
  if err != nil {
    log.Fatal(err)
    // TODO quit here?
  }

  var path string

  for _, file := range files {
    if file.IsDir() {
      path = directory + "/" + file.Name()
      if file.Name() == ".git" { // TODO move to const
        path = strings.TrimSuffix(path, "/.git")
        fmt.Println(path)
        directories = append(directories, path)
        continue
      }
      // TODO get the ignored files list from .gitignore, if present
      if file.Name() == "node_modules" {
        continue
      }

      directories = scanGitDirectories(directories, path)
    }
  }

  return directories
}

// Starts the recursive search of directories from the given directory.
func recursiveScanDirectory(directory string) []string {
  return scanGitDirectories(make([]string, 0), directory)
}

// Returns the dotfile for the repos list.
func getDotfilePath() string {
  usr, err := user.Current()
  if err != nil {
    log.Fatal(err)
    // TODO quit here?
  }

  dotfile := usr.HomeDir + "/.config/.gostat"

  return dotfile
}

// Scans a directory for git repositories.
func scan(directory string) {
  fmt.Println("Directories found:")
  repositories := recursiveScanDirectory(directory)
  filePath := getDotfilePath()
  addNewSliceElementsToFile(filePath, repositories)
  fmt.Println("Successfully added")
}
