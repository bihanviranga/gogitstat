package main

import (
  "flag"
  "fmt"
)

func scan(path string) {
  fmt.Println("scan %v", path)
}

func stats(email string) {
  fmt.Println("stats %v", email)
}

func main() {
  var folder string
  var email string

  flag.StringVar(&folder, "add", "", "add a new folder to scan for Git repos")
  flag.StringVar(&email, "email", "your@email.com", "the email to scan")
  flag.Parse()

  if folder != "" {
    scan(folder)
    return
  }

  stats(email)
}
