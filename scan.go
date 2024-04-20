package main

import (
  "flag"
  "fmt"
  "strings"
  "os"
  "log"
)

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

// Scans a directory for git repositories
func scan(directory string) {
  fmt.Println("Directories found:")
  repositories := recursiveScanDirectory(directory)
  filePath := getDotfilePath()
  addNewSliceElementsToFile(filePath, repositories)
  fmt.Println("Successfully added")
}

func stats(email string) {
  fmt.Println("stats %v", email)
}

func main() {
  var directory string
  var email string

  flag.StringVar(&directory, "add", "", "add a directory to scan for Git repos")
  flag.StringVar(&email, "email", "your@email.com", "the email to scan")
  flag.Parse()

  if directory != "" {
    scan(directory)
    return
  }

  stats(email)
}
