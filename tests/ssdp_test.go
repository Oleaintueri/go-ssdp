package tests

import (
	"testing"
)

func Test_SsdpDevices(t *testing.T) {
	ssdpClient := ssdp.NewSSDP()

	devices, err := ssdpClient.SearchDevices(ssdp.ALL.String())

	if err != nil {
		t.Error(err)
	}

	for i := range devices {
		t.Logf("Device: %v", devices[i])
	}
}

func Test_SsdpResponses(t *testing.T) {
	ssdpClient := ssdp.NewSSDP(ssdp.WithTimeout(2000))

	responses, err := ssdpClient.Search(ssdp.ALL.String())

	if err != nil {
		t.Error(err)
	}

	for i := range responses {
		t.Logf("Response: %v", responses[i])
	}
}
