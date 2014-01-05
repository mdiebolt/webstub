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
  Files File
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

    output, _ := getGist(gistId)
    fmt.Printf("%+v", output)
  }
}

func port() string {
  if os.Getenv("PORT") == "" {
    return "8888"
  }

  return os.Getenv("PORT")
}

func main() {
  port := port()

  http.HandleFunc("/favicon.ico", handlerIcon)
  http.HandleFunc("/", handlerRoot)

  log.Println("Listening on port " + port)
  log.Fatal(http.ListenAndServe(":" + port, nil))
}