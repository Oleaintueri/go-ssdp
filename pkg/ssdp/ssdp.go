// SSDP (Simple Service Discovery Protocol) package provides an implementation of the SSDP
// specification.
package ssdp

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type options struct {
	// The port for SSDP discovery
	port int
	// The IP for SSDP broadcast
	broadcastIp string
	// timeout in milliseconds
	timeout time.Duration
}

type OptionSSDP interface {
	apply(*options)
}

type portOption int

func (p portOption) apply(opts *options) {
	opts.port = int(p)
}

type broadcastOption string

func (b broadcastOption) apply(opts *options) {
	opts.broadcastIp = string(b)
}

type timeoutOption int

func (t timeoutOption) apply(opts *options) {
	opts.timeout = time.Duration(t) * time.Millisecond
}

func WithPort(port int) OptionSSDP {
	return portOption(port)
}

func WithBroadcast(broadcast string) OptionSSDP {
	return broadcastOption(broadcast)
}

func WithTimeout(timeout int) OptionSSDP {
	return timeoutOption(timeout)
}

type SSDP struct {
	*options
}

func NewSSDP(opts ...OptionSSDP) *SSDP {
	options := &options{
		port:        9000,
		broadcastIp: "239.235.255.250",
	}

	for _, o := range opts {
		o.apply(options)
	}

	return &SSDP{options}
}

// The search response from a device implementing SSDP.
type SearchResponse struct {
	Control      string
	Server       string
	ST           string
	Ext          string
	USN          string
	Location     *url.URL
	Date         time.Time
	ResponseAddr *net.UDPAddr
}

type Device struct {
	SpecVersion      SpecVersion `xml:"specVersion"`
	URLBase          string      `xml:"URLBase"`
	DeviceType       string      `xml:"device>deviceType"`
	FriendlyName     string      `xml:"device>friendlyName"`
	Manufacturer     string      `xml:"device>manufacturer"`
	ManufacturerURL  string      `xml:"device>manufacturerURL"`
	ModelDescription string      `xml:"device>modelDescription"`
	ModelName        string      `xml:"device>modelName"`
	ModelNumber      string      `xml:"device>modelNumber"`
	ModelURL         string      `xml:"device>modelURL"`
	SerialNumber     string      `xml:"device>serialNumber"`
	UDN              string      `xml:"device>UDN"`
	UPC              string      `xml:"device>UPC"`
	PresentationURL  string      `xml:"device>presentationURL"`
	Icons            []Icon      `xml:"device>iconList>icon"`
}

type SpecVersion struct {
	Major int `xml:"major"`
	Minor int `xml:"minor"`
}

type Icon struct {
	MIMEType string `xml:"mimetype"`
	Width    int    `xml:"width"`
	Height   int    `xml:"height"`
	Depth    int    `xml:"depth"`
	URL      string `xml:"url"`
}

// The search reader interface to read UDP packets on the wire with a timeout
// period specified.
type searchReader interface {
	ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error)
	SetReadDeadline(t time.Time) error
}

// Search the network for SSDP devices using the given search string and duration
// to discover new devices. This function will return an array of SearchReponses
// discovered.
func (ssdp *SSDP) Search(search string) ([]SearchResponse, error) {
	conn, err := ssdp.listenForSearchResponses()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	searchBytes, broadcastAddr, err := ssdp.buildSearchRequest(search)

	if err != nil {
		return nil, err
	}

	// Write search bytes on the wire so all devices can respond
	_, err = conn.WriteTo(searchBytes, broadcastAddr)
	if err != nil {
		return nil, err
	}

	return ssdp.readSearchResponses(conn)
}

func (ssdp *SSDP) SearchDevices(search string) ([]Device, error) {
	responses, err := ssdp.Search(search)

	if err != nil {
		return nil, err
	}

	uniqueLocations := make(map[url.URL]bool)

	for _, response := range responses {
		uniqueLocations[*response.Location] = true
	}

	locations := make([]url.URL, 0, len(uniqueLocations))
	for location, _ := range uniqueLocations {
		locations = append(locations, location)
	}

	devices := make([]Device, 0, len(locations))
	for _, location := range locations {
		device, err := parseDescriptionXml(location)
		if err != nil {
			return nil, err
		}
		devices = append(devices, *device)
	}

	return devices, nil
}

func (ssdp *SSDP) listenForSearchResponses() (*net.UDPConn, error) {
	serverAddr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", ssdp.port))
	return net.ListenUDP("udp", serverAddr)
}

func (ssdp *SSDP) buildSearchRequest(st string) ([]byte, *net.UDPAddr, error) {
	// Placeholder to replace with * later on
	// replaceMePlaceHolder := "/replacemewithstar"

	broadcastAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ssdp.broadcastIp, ssdp.port))

	if err != nil {
		return nil, nil, err
	}

	request, _ := http.NewRequest("M-SEARCH",
		fmt.Sprintf("http://%s/*", broadcastAddr.String()), strings.NewReader(""))

	headers := request.Header
	headers.Set("User-Agent", "")
	headers.Set("st", st)
	headers.Set("man", `"ssdp:discover"`)
	headers.Set("mx", strconv.Itoa(int(ssdp.timeout/time.Second)))

	searchBytes := make([]byte, 0, 1024)
	buffer := bytes.NewBuffer(searchBytes)
	err = request.Write(buffer)

	if err != nil {
		return nil, nil, fmt.Errorf("error writing to buffer")
	}

	searchBytes = buffer.Bytes()

	// Replace placeholder with *. Needed because request always escapes * when it shouldn't
	// searchBytes = bytes.Replace(searchBytes, []byte(replaceMePlaceHolder), []byte("*"), 1)

	return searchBytes, broadcastAddr, nil
}

func (ssdp *SSDP) readSearchResponses(reader searchReader) ([]SearchResponse, error) {
	responses := make([]SearchResponse, 0, 10)
	// Only listen for responses for duration amount of time.
	err := reader.SetReadDeadline(time.Now().Add(ssdp.timeout))

	if err != nil {
		return nil, err
	}

	buf := make([]byte, 1024)
	for {
		rlen, addr, err := reader.ReadFromUDP(buf)
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			break // duration reached, return what we've found
		}
		if err != nil {
			return nil, err
		}

		response, err := parseSearchResponse(bytes.NewReader(buf[:rlen]), addr)
		if err != nil {
			return nil, err
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

func parseSearchResponse(httpResponse io.Reader, responseAddr *net.UDPAddr) (*SearchResponse, error) {
	reader := bufio.NewReader(httpResponse)
	request := &http.Request{} // Needed for ReadResponse but doesn't have to be real
	response, err := http.ReadResponse(reader, request)
	if err != nil {
		return nil, err
	}
	headers := response.Header

	res := &SearchResponse{}

	res.Control = headers.Get("cache-control")
	res.Server = headers.Get("server")
	res.ST = headers.Get("st")
	res.Ext = headers.Get("ext")
	res.USN = headers.Get("usn")
	res.ResponseAddr = responseAddr

	if headers.Get("location") != "" {
		res.Location, err = response.Location()
		if err != nil {
			return nil, err
		}
	}

	date := headers.Get("date")
	if date != "" {
		res.Date, err = http.ParseTime(date)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func parseDescriptionXml(url url.URL) (*Device, error) {
	response, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	decoder := xml.NewDecoder(response.Body)

	device := &Device{}

	err = decoder.Decode(device)

	if err != nil {
		return nil, err
	}

	return device, nil
}
