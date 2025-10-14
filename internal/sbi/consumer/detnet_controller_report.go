package consumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/HanHongChen/tsctsf/internal/context"
	"github.com/HanHongChen/tsctsf/internal/logger"
)

// DetNet Controller API request structure
type DetNetInterfaceRequest struct {
	IetfInterfaces struct {
		Interface []DetNetInterface `json:"interface"`
	} `json:"ietf-interfaces:interfaces"`
}

type DetNetInterface struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	PhysAddress string `json:"physAddress"`
	OperStatus  string `json:"oper-status"`
	Statistics  struct {
		DiscontinuityTime string `json:"discontinuity-time"`
	} `json:"statistics"`
}

// SendToDetNetController sends interface information to DetNet Controller
func SendToDetNetController() error {
	// Get router information from context
	tsnContext := context.GetSelf()
	if tsnContext == nil {
		return fmt.Errorf("TSCTSF context is not initialized")
	}

	// Generate API request
	request, err := generateInterfaceRequest(tsnContext)
	logger.ConsumerLog.Warnf("Generated DetNet Interface Request: %+v", request)
	if err != nil {
		logger.ConsumerLog.Errorf("Failed to generate interface request: %v", err)
		return err
	}

	// Send to DetNet Controller
	return sendRequest(tsnContext, request)
}

// generateInterfaceRequest generates interface request based on router information
func generateInterfaceRequest(tsnContext *context.TSCTSFContext) (DetNetInterfaceRequest, error) {
	var request DetNetInterfaceRequest

	interfaceIndex := 0
	for _, router := range tsnContext.Routers {
		// Validate MAC address
		macAddr, err := formatMacAddress(router.MacAddr)
		if err != nil {
			logger.ConsumerLog.Errorf("Router %s has invalid MAC address: %v", router.UpNodeId, err)
			return request, fmt.Errorf("router %s has invalid MAC address: %v", router.UpNodeId, err)
		}

		iface := DetNetInterface{
			Name: fmt.Sprintf("nic-%d", interfaceIndex),
			Description: fmt.Sprintf("Interface for %s (UE IP: %s, Port: %d, MTU: %d)",
				router.UpNodeId, router.UeIp, router.Port, router.Mtu),
			Type:        "iana-if-type:ethernetCsmacd",
			PhysAddress: macAddr,
			OperStatus:  "up",
		}

		iface.Statistics.DiscontinuityTime = time.Now().Format("2006-01-02T15:04:05-07:00")
		request.IetfInterfaces.Interface = append(request.IetfInterfaces.Interface, iface)
		interfaceIndex++
	}

	return request, nil
}

// formatMacAddress validates and formats MAC address
func formatMacAddress(macAddr string) (string, error) {
	if macAddr == "" {
		return "", fmt.Errorf("MAC address is empty")
	}
	return macAddr, nil
}

// sendRequest sends the request to DetNet Controller with Basic Auth
func sendRequest(tsnContext *context.TSCTSFContext, request DetNetInterfaceRequest) error {
	jsonData, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		logger.ConsumerLog.Errorf("Failed to marshal JSON: %v", err)
		return err
	}

	logger.ConsumerLog.Infof("Sending to DetNet Controller:\n%s", string(jsonData))

	// Build request URL
	url := fmt.Sprintf("%s://%s:%d/%s",
		tsnContext.DetNetController.Scheme,
		tsnContext.DetNetController.IPv4,
		tsnContext.DetNetController.Port,
		tsnContext.DetNetController.Uri)

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.ConsumerLog.Errorf("Failed to create request: %v", err)
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add Basic Auth
	req.SetBasicAuth(tsnContext.DetNetController.User, tsnContext.DetNetController.Password)

	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Send request
	logger.ConsumerLog.Infof("Sending POST request to %s with Basic Auth (user: %s)",
		url, tsnContext.DetNetController.User)

	resp, err := client.Do(req)
	if err != nil {
		logger.ConsumerLog.Errorf("Failed to send request: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.ConsumerLog.Errorf("Failed to read response: %v", err)
		return err
	}

	logger.ConsumerLog.Infof("DetNet Controller response status: %s", resp.Status)
	logger.ConsumerLog.Debugf("DetNet Controller response body: %s", string(body))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API request failed with status: %s, body: %s", resp.Status, string(body))
	}

	logger.ConsumerLog.Infof("Successfully sent interface information to DetNet Controller")
	return nil
}
