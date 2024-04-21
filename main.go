package main

import ("flag")

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
