<h1 align="center">Go SSDP client</h1>

A library that implements the client side of SSDP (Simple Service Discovery Protocol).

This is an altered version from the original by [Ben Curren](https://github.com/bcurren/go-ssdp).

Build with :heart: by

<a href="https://oleaintueri.com"><img src="https://oleaintueri.com/images/oliv.svg" width="60px"/><img width="200px" style="padding-bottom: 10px" src="https://oleaintueri.com/images/oleaintueri.svg"/></a>

[Oleaintueri](https://oleaintueri.com) is sponsoring the development and maintenance of this project within their organisation.


## Getting Started

### Installation

    GOMODULE111=on go get github.com/Oleaintueri/gossdp


### Usage

```go
package main

import (
	"fmt"
	"github.com/Oleaintueri/gossdp"
	"github.com/Oleaintueri/gossdp/pkg/ssdp"
	"time"
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

### How to contribute

* Fork the repository
* Create an issue with your desired update
* Write tests and code
* Submit a pull request
