package main

import (
  "encoding/json"
  "os"
  "net/http"
  "io/ioutil"
  "io"
  "log"
  "time"
  "strings"
  "bytes"
  "crypto/rand"
  "encoding/hex"
)

type PollResponseServer struct {
  UserKey string `json:"userKey,omitempty"`
  ServerID string `json:"serverId,omitempty"`
}

type PollResponse struct {
  Server PollResponseServer `json:"server,omitempty"`
}

type UpdateResponse struct {
  Server PollResponseServer `json:"server,omitempty"`
}

type PollRequest struct {
  Name string `json:"name,omitempty"`
  Description string `json:"description,omitempty"`
  MinimumPollTimeSec int `json:"minimumPollTimeSec,omitempty"`
  ServerAPIKey string `json:"serverApiKey,omitempty"`
  ServerID string `json:"serverId,omitempty"`
}

type UpdateRequest struct {
  Name string `json:"name,omitempty"`
  Description string `json:"description,omitempty"`
  MinimumPollTimeSec int `json:"minimumPollTimeSec,omitempty"`
  ServerAPIKey string `json:"serverApiKey,omitempty"`
  ServerID string `json:"serverId,omitempty"`
}

type Config struct {
  Name string `json:"name,omitempty"`
  Description string `json:"description,omitempty"`
  MinimumPollTimeSec int `json:"minimumPollTimeSec,omitempty"`
  MinimumUpdateTimeSec int `json:"minimumUpdateTimeSec,omitempty"`
  ServerAPIKey string `json:"serverApiKey,omitempty"`
  ServerID string `json:"serverId,omitempty"`
  BiboopServer string `json:"biboopServer,omitempty"`
}
var contentType = "application/json; charset=UTF-8"
var config *Config

func startPolling() {
  log.Println("Starting to poll")
  pollUrl := strings.Join([]string{config.BiboopServer, "api", "server", "poll"}, "/")
  updateUrl := strings.Join([]string{config.BiboopServer, "api", "server", "update"}, "/")

  initRequest(updateUrl)

  pollTicker := time.NewTicker(time.Duration(config.MinimumPollTimeSec) * time.Second)
  updateTicker := time.NewTicker(time.Duration(config.MinimumUpdateTimeSec) * time.Second)

  for {
    select {
    case <- pollTicker.C:
      var pollResponse PollResponse
      executeRequest(pollUrl, buildPollRequestBody(), &pollResponse)
    case <- updateTicker.C:
      var updateResponse UpdateResponse
      executeRequest(updateUrl, buildUpdateRequestBody(), &updateResponse)
    }
  }
}

func initRequest(updateUrl string) {
  var updateResponse UpdateResponse
  executeRequest(updateUrl, buildUpdateRequestBody(), &updateResponse)
}

func executeRequest(url string, request *bytes.Buffer, result interface{}) {
  var resp *http.Response
  var body []byte
  var err error

  if resp, err = http.Post(url, contentType, request); err != nil {
    log.Println("Request failed", err)
    return
  }

  defer resp.Body.Close()
  body, err = ioutil.ReadAll(resp.Body)
  if err != nil || (resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated) {
    log.Println("Request failed", err, resp.StatusCode, string(body))
    return
  }

  if err = json.Unmarshal(body, &result); err != nil {
    log.Println("Failed to parse response body")
  }
}

func buildPollRequestBody() *bytes.Buffer {
  var requestBytes []byte
  var err error
  pollRequest := PollRequest{
    Name: config.Name,
    Description: config.Description,
    MinimumPollTimeSec: config.MinimumPollTimeSec,
    ServerAPIKey: config.ServerAPIKey,
    ServerID: config.ServerID }
  if requestBytes, err = json.Marshal(pollRequest); err != nil {
    log.Println("Failed to marshal request JSON")
    panic(err)
  }

  return bytes.NewBuffer(requestBytes)
}

func buildUpdateRequestBody() *bytes.Buffer {
  var requestBytes []byte
  var err error
  updateRequest := UpdateRequest{
    Name: config.Name,
    Description: config.Description,
    MinimumPollTimeSec: config.MinimumPollTimeSec,
    ServerAPIKey: config.ServerAPIKey,
    ServerID: config.ServerID }
  if requestBytes, err = json.Marshal(updateRequest); err != nil {
    log.Println("Failed to marshal request JSON")
    panic(err)
  }

  return bytes.NewBuffer(requestBytes)
}

func writeOutConfiguration(filename string) {
  var configBytes []byte
  var err error
  if configBytes, err = json.MarshalIndent(config, "", "  "); err != nil {
    panic(err)
  }
  if err = ioutil.WriteFile(filename, configBytes, 0600); err != nil {
    panic(err)
  }
}

func UID() []byte {
  buf := make([]byte, 16)
  io.ReadFull(rand.Reader, buf)
  buf[6] = (buf[6] & 0x0f) | 0x40
  buf[8] = (buf[8] & 0x3f) | 0x80
  return buf
}

func UIDString() string {
  return hex.EncodeToString(UID())
}

func main() {
  var configFile []byte
  var err error
  var configInHome = true
  var home = os.Getenv("HOME")
  var homeFileName = strings.Join([]string{ home, ".biboop/config.json" }, "/")
  var etcFileName = "/etc/biboop/config.json"

  if configFile, err = ioutil.ReadFile(homeFileName); err != nil {
    configInHome = false
    if configFile, err = ioutil.ReadFile(etcFileName); err != nil {
      panic(err)
    }
  }

  config = new(Config)
  if err = json.Unmarshal(configFile, config); err != nil {
    panic(err)
  }

  if config.ServerID == "" {
    config.ServerID = UIDString()
    if configInHome {
      writeOutConfiguration(homeFileName)
    } else {
      writeOutConfiguration(etcFileName)
    }
  }

  startPolling()
}
