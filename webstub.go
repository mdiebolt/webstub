package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "strings"
)

type Gist struct {
  Id string
  Description string
  User User
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

func getGist(id string) (Gist, error) {
  url := "https://api.github.com/gists/" + id

  resp, err := http.Get(url)
  panicError(err)

  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  panicError(err)

  var data Gist
  err = json.Unmarshal(body, &data)
  if err != nil {
    fmt.Printf("%T\n%s\n%#v\n", err, err, err)
    switch v := err.(type) {
      case *json.SyntaxError:
        fmt.Println(string(body[v.Offset-40:v.Offset]))
    }
  }

  return data, err
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
    nonJSONFile := false

    if _, okay := global.FileIndex[gistId]; !okay {
      global.FileIndex[gistId] = 0
    }

    output, err := getGist(gistId)
    panicError(err)

    res := ""
    index := 0
    for _, v := range output.Files {
      if v.Type != "JSON" {
        nonJSONFile = true
      }

      if global.FileIndex[gistId] == index {
        res = v.Content
      }

      index += 1
    }

    w.Header().Set("Content-Type", "application/json")

    if nonJSONFile {
      fmt.Fprint(w, Response{"message": "One or more of the files in this Gist aren't JSON"})
    } else if len(output.Files) == 0 {
      fmt.Fprint(w, Response{"message": "This Gist doesn't have any files"})
    } else {
      global.FileIndex[gistId] = (global.FileIndex[gistId] + 1) % (len(output.Files))
      fmt.Fprint(w, res)
    }
  }
}

func port() string {
  if os.Getenv("PORT") == "" {
    return "8888"
  }

  return os.Getenv("PORT")
}

var global struct {
  FileIndex map[string]int
}

type Response map[string]interface{}

func (r Response) String() (s string) {
  b, err := json.Marshal(r)
  if err != nil {
    s = ""
    return
  }
  s = string(b)
  return
}

func main() {
  port := port()

  global.FileIndex = make(map[string]int, 100)

  http.HandleFunc("/favicon.ico", handlerIcon)
  http.HandleFunc("/", handlerRoot)

  log.Println("Listening on port " + port)
  log.Fatal(http.ListenAndServe(":" + port, nil))
}