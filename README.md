# Revogi library

This package provides a way to read and control a Revogi Wi-Fi Smart Powerstrip (SOW323).
It uses the HTTP API on server.revogi.net.

This library was tested with firmware version 5.12.

## Features
* Retrieve devices and their statistics (e.g. current power consumption or switch states) from your Revogi account
* Control switch states per port

## Example
```
package main

import (
    "log"
    "github.com/LitecoinTim/revogi"
)

func main() {
    r := revogi.NewClient(&http.Client{Timeout: 3}, revogi.Config{Username: "user@host.com", Password: "mypass"})

    if err := r.Login(); err != nil {
        log.Fatalf("login failed: %v", err)
    }
    
    devices, err := r.GetDevices()
    if err != nil {
        log.Fatalf("failed to retrieve devices: %v", err)
    }
    
    // ... 
}
```
