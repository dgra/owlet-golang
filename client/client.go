package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"time"
)

type Application struct {
	ID     string `json:"app_id"`
	Secret string `json:"app_secret"`
}

type User struct {
	Email       string      `json:"email"`
	Password    string      `json:"password"`
	Application Application `json:"application"`
}

type Payload struct {
	User User `json:"user"`
}

type Device struct {
	DSN              string    `json:"dsn"`
	ProductName      string    `json:"product_name"`
	Model            string    `json:"model"`
	ConnectionStatus string    `json:"connection_status"`
	DeviceType       string    `json:"device_type"`
	SWVersion        string    `json:"sw_version"`
	Mac              string    `json:"mac"`
	ConnectedAt      time.Time `json:"connected_at"`
}

type DeviceRoot struct {
	Device Device `json:"device"`
}

// TODO: Make multiple properties for each value type?
type Property struct {
	Key         int       `json:"key"`
	BaseType    string    `json:"base_type"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Value       FlexValue `json:"value"`
	UpdatedAt   FlexTime  `json:"data_updated_at"`
}

type FlexValue string

type FlexTime struct {
	time.Time
}

type PropertyRoot struct {
	Property *Property `json:"property"`
}

type Authentication struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Role         string `json:"role"`
}

type Client struct {
	implementationClient *http.Client
	Email                string
	Password             string
	ActivePropID         int
	Auth                 *Authentication
	Device               *Device
}

type IntDatapoint struct {
	Value     int                  `json:"value"`
	Metadata  map[string]FlexValue `json:"metadata"`
	UpdatedAt time.Time            `json:"updated_at"`
}

type datapointRequest struct {
	Datapoint IntDatapoint `json:"datapoint"`
}

func (fv *FlexValue) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		return json.Unmarshal(b, (*string)(fv))
	}

	if string(b) == "null" {
		*fv = FlexValue("")
		return nil
	}

	*fv = FlexValue(fmt.Sprintf("\"%s\"", string(b)))
	return nil
}

func (ft *FlexTime) UnmarshalJSON(b []byte) error {
	if bytes.Compare(b, []byte{'"', 'n', 'u', 'l', 'l', '"'}) == 0 {
		*ft = FlexTime{}
		return nil
	}

	var currTime time.Time
	err := json.Unmarshal(b, &currTime)
	if err != nil {
		return err
	}
	*ft = FlexTime{currTime}
	return nil
}

func logRequest(req *http.Request) {
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		fmt.Println("Error dumping request details:", err)
	}
	fmt.Printf("Outgoing request:\n%s\nEnd request\n", string(dump))
}

func (c *Client) Post(subdomain, endpoint string, data io.Reader, v interface{}) error {
	return c.MakeRequest("POST", subdomain, endpoint, data, v)
}

func (c *Client) Get(subdomain, endpoint string, v interface{}) error {
	return c.MakeRequest("GET", subdomain, endpoint, nil, v)
}

func (c *Client) MakeRequest(method, subdomain, endpoint string, data io.Reader, v interface{}) error {
	req, err := NewRequestWithAuthoriztion(c.Auth, method, subdomain, endpoint, data)
	if err != nil {
		return err
	}

	logRequest(req)

	resp, err := c.implementationClient.Do(req)
	if err != nil {
		// Err is 401? Refresh auth and try again.
		fmt.Println("401?", err)
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(endpoint, string(body))
	return json.Unmarshal(body, v)
}

func (c *Client) authenticate() error {
	if c.Auth == nil {
		return c.login()
	}
	return nil
}

func (c *Client) SetFirstDevice() error {
	devices, err := c.GetDevices()
	if err != nil {
		return err
	}

	c.Device = &devices[0]
	return nil
}

func (c *Client) GetDevices() ([]Device, error) {
	deviceRoots := make([]DeviceRoot, 0)
	err := c.Get("ads-field", "apiv1/devices.json", &deviceRoots)
	if err != nil {
		return []Device{}, err
	}

	devices := make([]Device, len(deviceRoots))
	for i, v := range deviceRoots {
		devices[i] = v.Device
	}
	return devices, nil
}

func (c *Client) GetPropertyByName(deviceID, name string) (*Property, error) {
	endpoint := fmt.Sprintf("apiv1/dsns/%s/properties/%s", deviceID, name)
	propertyRoot := &PropertyRoot{}
	err := c.Get("ads-field", endpoint, propertyRoot)
	return propertyRoot.Property, err
}

func (c *Client) GetProperties(deviceID string) (map[string]*Property, error) {
	endpoint := fmt.Sprintf("apiv1/dsns/%s/properties.json", deviceID)

	propertyRoots := make([]PropertyRoot, 0)
	err := c.Get("ads-field", endpoint, &propertyRoots)
	if err != nil {
		return make(map[string]*Property), err
	}
	properties := make(map[string]*Property)
	for _, v := range propertyRoots {
		property := v.Property
		properties[v.Property.Name] = property
	}

	return properties, nil
}

func (c *Client) SetAppActiveStatus(deviceID string) (bool, error) {
	endpoint := fmt.Sprintf("apiv1/dsns/%s/properties/APP_ACTIVE/datapoints.json", deviceID)

	reqDP := datapointRequest{
		Datapoint: IntDatapoint{
			Value: 1,
		},
	}
	sendingBody, err := StructToReader(reqDP)
	if err != nil {
		return false, err
	}

	respDP := &datapointRequest{}
	err = c.Post("ads-field", endpoint, sendingBody, respDP)
	if err != nil {
		return false, err
	}
	fmt.Printf("%+v\n", respDP)
	return true, nil
}

func (c *Client) login() error {
	fmt.Println("Logging in...")
	data := Payload{
		User: User{
			Email:    c.Email,
			Password: c.Password,
			Application: Application{
				ID:     "OWL-id",
				Secret: "OWL-4163742",
			},
		},
	}

	sendingBody, err := StructToReader(data)
	if err != nil {
		return err
	}

	req, err := NewRequest("POST", "user-field", "users/sign_in", sendingBody)
	if err != nil {
		return err
	}

	resp, err := c.implementationClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	auth := &Authentication{}
	err = json.Unmarshal(body, auth)
	fmt.Println(auth)

	c.Auth = auth
	return nil
}

func New(email, password string) (*Client, error) {
	c := &Client{
		Email:                email,
		Password:             password,
		implementationClient: http.DefaultClient,
	}

	err := c.authenticate()
	return c, err
}

func NewRequestWithAuthoriztion(auth *Authentication, method, subdomain, endpoint string, data io.Reader) (*http.Request, error) {
	req, err := NewRequest(method, subdomain, endpoint, data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("auth_token %s", auth.AccessToken))
	return req, nil
}

func NewRequest(method, subdomain, endpoint string, data io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("https://%s.aylanetworks.com/%s", subdomain, endpoint)
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func StructToReader(data interface{}) (*bytes.Reader, error) {
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(payloadBytes), nil
}

//
// func (c *Client) refreshAuth(auth *Authentication) {
//
// }
