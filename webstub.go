package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "strings"
  "time"
)

type Response map[string]interface{}

var global struct {
  Gists map[string] *Gist
}

type Gist struct {
  Id string
  Description string
  User User
  FileIndex int         // Not derived from JSON, used to loop over gist file content
  RefreshedAt time.Time // Not derived from JSON, used to expire Gist cache
  Files map[string]File
}

type User struct {
  Login string
}

type File struct {
  Filename string
  Type string
  Content string
}

func parseRequest(r *http.Request) (gistId string) {
  path := strings.Split(r.URL.Path, "/")
  return strings.Join(path, "")
}

func getGist(id string) (*Gist, error) {
  fmt.Println("Making API request")
  url := "https://api.github.com/gists/" + id

  resp, err := http.Get(url)
  panicError(err)
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  panicError(err)

  var data Gist
  err = json.Unmarshal(body, &data)
  panicError(err)

  // Cache gist so that we don't have to
  // use tons of GitHub API requests
  data.RefreshedAt = time.Now()
  global.Gists[id] = &data

  return &data, err
}

func panicError(err error) {
  if err != nil {
    panic(err.Error())
  }
}

func handlerIcon(w http.ResponseWriter, r *http.Request) {}

func handlerRoot(w http.ResponseWriter, r *http.Request) {
  if r.Method == "GET" {
    gistId := parseRequest(r)

    var hasKey bool
    var err error
    output, hasKey := global.Gists[gistId]

    if hasKey {
      now := time.Now()
      delta := now.Sub(output.RefreshedAt)

      // blow away the cached Gist
      // if it's older than 2 hours
      if (int(delta.Minutes()) > 120) {
        delete(global.Gists, gistId)

        // refresh the data and cache it
        output, err = getGist(gistId)
        panicError(err)
      }
    }

    if !hasKey {
      output, err = getGist(gistId)
      panicError(err)
    }

    w.Header().Set("Content-Type", "application/json")

    length := len(output.Files)
    if length == 0 {
      fmt.Fprint(w, Response{"message": "This Gist doesn't have any files"})
    } else if hasNonJsonFile(output.Files) {
      fmt.Fprint(w, Response{"message": "One or more of the files in this Gist aren't JSON"})
    } else {
      output.FileIndex = (output.FileIndex + 1) % length
      fmt.Fprint(w, currentFileContent(output.Files, output.FileIndex))
    }
  }
}

func currentFileContent(files map[string]File, index int) string {
  output := ""
  i := 0

  for _, v := range files {
    if index == i {
      output = v.Content
    }

    i += 1
  }

  return output
}

func hasNonJsonFile(files map[string]File) bool {
  output := false

  for _, v := range files {
    if v.Type != "application/json" {
      output = true
    }
  }

  return output
}

func port() string {
  if os.Getenv("PORT") == "" {
    return "8888"
  }

  return os.Getenv("PORT")
}

func (r Response) String() string {
  b, err := json.Marshal(r)
  if err != nil {
    return ""
  }
  return string(b)
}

func main() {
  port := port()

  global.Gists = make(map[string]*Gist, 100)

  http.HandleFunc("/favicon.ico", handlerIcon)
  http.HandleFunc("/", handlerRoot)

  log.Println("Listening on port " + port)
  log.Fatal(http.ListenAndServe(":" + port, nil))
}