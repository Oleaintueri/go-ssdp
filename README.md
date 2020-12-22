<h1 align="center">go-ssdp</h1>

A library that implements the client side of SSDP (Simple Service Discovery Protocol).

Please see [godoc.org](http://godoc.org/github.com/bcurren/go-ssdp) for a detailed API description.

This is an altered version from the original by [Ben Curren](https://github.com/bcurren/go-ssdp).

## Getting Started

### Installation

    GOMODULE111=on go get github.com/Oleaintueri/go-ssdp


### Usage

```go
package main

import (
	"github.com/Oleaintueri/go-ssdp"
	"github.com/Oleaintueri/go-ssdp/pkg/ssdp"
	"time"
	"fmt"
)

func main() {
	ssdpClient := ssdp.NewSSDP(ssdp.WithTimeout(1000))

	// Get the devices on the network directly
	devices, err := ssdpClient.SearchDevices("upnp:rootdevice")

	if err != nil {
		panic(err)
	}

	// Get the general responses on the network
	responses, err := ssdpClient.Search("upnp:rootdevice")

	if err != nil {
		panic(err)
	}

	for _, device := range devices {
		fmt.Printf("Device: %v", device)
	}

	for _, resp := range responses {
		fmt.Printf("Response: %v", resp)
	}
}
```

## How to contribute

* Fork
* Write tests and code
* Run go fmt
* Submit a pull request

