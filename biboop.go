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

type PollRequest struct {
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
  ServerAPIKey string `json:"serverApiKey,omitempty"`
  ServerID string `json:"serverId,omitempty"`
  BiboopServer string `json:"biboopServer,omitempty"`
}
var contentType = "application/json; charset=UTF-8"
var config *Config

func startPolling() {
  log.Println("Starting to poll")
  var pollTime = time.Duration(config.MinimumPollTimeSec) * time.Second
  var pollUrl = strings.Join([]string{config.BiboopServer, "api", "server", "poll"}, "/")
  for {
    executeRequest(pollUrl)
    time.Sleep(pollTime)
  }
}

func executeRequest(pollUrl string) {
  var resp *http.Response
  var body []byte
  var err error

  if resp, err = http.Post(pollUrl, contentType, buildPollRequestBody()); err != nil {
    log.Println("Request failed", err)
    return
  }

  body, err = ioutil.ReadAll(resp.Body)
  defer resp.Body.Close()
  if err != nil || resp.StatusCode != http.StatusOK || resp.StatusCode != http.StatusCreated {
    log.Println("Request failed", err, resp.StatusCode, string(body))
    return
  }

  var pollResponse PollResponse
  if err = json.Unmarshal(body, &pollResponse); err != nil {
    log.Println("Failed to parse response body")
  } else {
    log.Println("Request succeeded")
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
