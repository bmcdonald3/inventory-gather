// pkg/inventory/collector.go

package inventory

import (
	"fmt"
	"bytes"
	"encoding/json"
	"net/http"
	"crypto/tls"
	"io"
)

// Placeholder for the actual API server address
const InventoryAPIHost = "http://localhost:8081"
const DefaultUsername = "root"
const DefaultPassword = "password" // Security note: Hardcoding credentials is not recommended in production.

func CollectAndPost(bmcIP string) error {
	// 1. Initialize Redfish Client
	rfClient, err := NewRedfishClient(bmcIP, DefaultUsername, DefaultPassword)
	if err != nil {
		return fmt.Errorf("failed to initialize Redfish client: %w", err)
	}

	// 2. REDFISH DISCOVERY (Placeholder for Step 3)
	// For testing the client, we can make a simple call to the root service entry point.
	body, err := rfClient.Get("") // Getting the root path (https://<ip>/redfish/v1)
	if err != nil {
		return fmt.Errorf("redfish client test failed: %w", err)
	}
	
	fmt.Printf("Redfish Client Test: Successfully connected to %s. Response size: %d bytes.\n", rfClient.BaseURL, len(body))

	// devices, err := discoverDevices(rfClient) // The next step's function

	// --- A. REDFISH DISCOVERY (Using simulated data for now) ---
	devices := []DiscoveredDevice{
		{DeviceType: "Node", Manufacturer: "HPE (Simulated)", SerialNumber: "ABC0001", RedfishURI: "/Systems/1"},
		{DeviceType: "CPU", Manufacturer: "Intel (Simulated)", SerialNumber: "CPU0002", RedfishURI: "/Systems/1/Processors/1"},
		{DeviceType: "DIMM", Manufacturer: "Micron (Simulated)", SerialNumber: "DIMM0003", RedfishURI: "/Systems/1/Memory/1"},
	}
	fmt.Println("Redfish Discovery Simulated: Found 3 devices.")

	// B. API POSTING
	// The API requires a two-step process: Create the resource, then update its status.
	// We will use a map to store the relationship between the temporary name and the API-assigned UID.
	nameToUID := make(map[string]string)
	
	// C. PROCESS DEVICE HIERARCHY (BMC/Node first)
	// Process the main system device first, as others are children.
	// In a real implementation, you'd process parent devices before children.
	
	for i, dev := range devices {
		// 1. Create the Device Resource Envelope (POST /devices)
		tempName := fmt.Sprintf("%s-%d-%s", dev.DeviceType, i, dev.SerialNumber)
		
		fmt.Printf("-> Creating resource envelope for %s (%s)...\n", tempName, dev.DeviceType)
		uid, err := createDeviceEnvelope(tempName)
		if err != nil {
			return fmt.Errorf("failed to create API envelope for %s: %w", tempName, err)
		}
		nameToUID[tempName] = uid
		
		// 2. Map the Redfish data to the API's Status struct
		// (A much cleaner way would be to unmarshal directly from Redfish and then transform).
		statusData := map[string]interface{}{
			"deviceType":   dev.DeviceType,
			"manufacturer": dev.Manufacturer,
			"partNumber":   dev.PartNumber,
			"serialNumber": dev.SerialNumber,
			// parentID would need to be resolved to a UUID here using the nameToUID map
			// For simplicity and focusing on top-level for now, we leave it out.
			// "parentID": resolveParentID(dev.ParentID, nameToUID),
			"properties": map[string]interface{}{
				"redfish_uri": dev.RedfishURI,
			},
		}

		// 3. Update the Status (PUT /devices/{uid}/status)
		fmt.Printf("-> Updating status for %s (UID: %s)...\n", tempName, uid)
		err = updateDeviceStatus(uid, statusData)
		if err != nil {
			return fmt.Errorf("failed to update status for %s: %w", tempName, err)
		}
		fmt.Printf("-> Successfully posted device %s\n", uid)
	}

	return nil
}

// pkg/inventory/collector.go

// discoverDevices is a simplified placeholder for the Redfish client logic.
func discoverDevices(bmcIP string) ([]DiscoveredDevice, error) {
	// **THIS IS THE CRITICAL LOGIC YOU WILL NEED TO DEVELOP FULLY.**
	//
	// You will need a proper Redfish client to:
	// 1. Connect to "https://" + bmcIP + "/redfish/v1"
	// 2. Authenticate (Basic or Session-based).
	// 3. Start traversal from /redfish/v1/Systems/ or /redfish/v1/Chassis/
	// 4. Follow links to /Systems/X/Processors, /Systems/X/Memory, etc.
	
	// Example of a minimal set of data you would gather:
	var devices []DiscoveredDevice

	// --- 1. System/Node ---
	devices = append(devices, DiscoveredDevice{
		DeviceType: "Node",
		Manufacturer: "HPE (Redfish System)",
		PartNumber: "ProLiant-BL460c-Gen10",
		SerialNumber: "ABC0001",
		RedfishURI: "/redfish/v1/Systems/1",
	})
	
	// --- 2. CPU/Processor ---
	devices = append(devices, DiscoveredDevice{
		DeviceType: "CPU",
		Manufacturer: "Intel",
		PartNumber: "Xeon-Gold-6240",
		SerialNumber: "CPU0002",
		ParentID: "/redfish/v1/Systems/1", // Placeholder
		RedfishURI: "/redfish/v1/Systems/1/Processors/1",
	})

	// --- 3. DIMM/Memory ---
	devices = append(devices, DiscoveredDevice{
		DeviceType: "DIMM",
		Manufacturer: "Micron",
		PartNumber: "32GB-DDR4-2666",
		SerialNumber: "DIMM0003",
		ParentID: "/redfish/v1/Systems/1", // Placeholder
		RedfishURI: "/redfish/v1/Systems/1/Memory/1",
	})
	
	fmt.Println("Redfish Discovery Simulated: Found 3 devices.")
	return devices, nil
}

// pkg/inventory/collector.go

type Metadata struct {
	UID string `json:"uid"`
}

type DeviceResponse struct {
	Metadata Metadata `json:"metadata"`
}

// createDeviceEnvelope POSTs to /devices to create the resource and get its UID.
func createDeviceEnvelope(name string) (string, error) {
	// Fabrica requires a name and optional metadata to create the envelope.
	payload := map[string]interface{}{
		"name": name,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(InventoryAPIHost+"/devices", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to post device envelope: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api returned status code %d when creating envelope", resp.StatusCode)
	}

	var deviceResponse DeviceResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceResponse); err != nil {
		return "", fmt.Errorf("failed to decode API response for UID: %w", err)
	}

	if deviceResponse.Metadata.UID == "" {
		return "", fmt.Errorf("API response did not contain a UID in the metadata")
	}

	return deviceResponse.Metadata.UID, nil
}

// updateDeviceStatus PUTs the observed state to /devices/{uid}/status.
func updateDeviceStatus(uid string, statusData map[string]interface{}) error {
	// The Fabrica PUT /status endpoint expects a body corresponding to the DeviceStatus struct.
	// We wrap the map in a "status" key because the API is structured as spec/status.
	payload := map[string]interface{}{
		"status": statusData,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal status payload: %w", err)
	}

	// Create an HTTP client for the PUT request
	url := InventoryAPIHost + "/devices/" + uid + "/status"
	fmt.Printf("Attempting PUT to: %s\n", url)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create PUT request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute PUT request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("api returned status code %d when updating status", resp.StatusCode)
	}

	// If the API call succeeded (Status 200), the 'payload' is now used.
	return nil
}

type RedfishClient struct {
	BaseURL  string
	Username string
	Password string
	HTTPClient *http.Client
}

// NewRedfishClient initializes the client with a specified BMC IP.
func NewRedfishClient(bmcIP, username, password string) (*RedfishClient, error) {
	// Redfish requires HTTPS and starts at the /redfish/v1 path.
	baseURL := fmt.Sprintf("https://%s/redfish/v1", bmcIP)

	// Create a custom HTTP client that trusts the BMC's self-signed certificate.
	// NOTE: In production, you would use proper certificate validation.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	
	return &RedfishClient{
		BaseURL:  baseURL,
		Username: username,
		Password: password,
		HTTPClient: &http.Client{Transport: tr},
	}, nil
}

func (c *RedfishClient) Get(path string) ([]byte, error) {
	// The path can be a full URI or a relative path (e.g., /Systems/1).
	targetURL := c.BaseURL + path
	
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redfish request for %s: %w", targetURL, err)
	}

	// Add Basic Authentication header
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Add("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute Redfish request for %s: %w", targetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Redfish API returned status code %d for %s", resp.StatusCode, targetURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}


func mapCommonProperties(rfProps CommonRedfishProperties, deviceType, redfishURI, parentURI string) DiscoveredDevice {
    // Redfish often uses Model if PartNumber is unavailable or vice-versa.
    // Prioritize PartNumber, fallback to Model.
    partNum := rfProps.PartNumber
    if partNum == "" {
        partNum = rfProps.Model
    }

	return DiscoveredDevice{
		DeviceType:   deviceType,
		Manufacturer: rfProps.Manufacturer,
		PartNumber:   partNum,
		SerialNumber: rfProps.SerialNumber,
		ParentID:     parentURI, // This will be resolved to a UUID later (Step 5)
		RedfishURI:   redfishURI,
	}
}

func getSystemInventory(c *RedfishClient, systemURI string) (*SystemInventory, error) {
	inv := &SystemInventory{CPUs: make([]DiscoveredDevice, 0), DIMMs: make([]DiscoveredDevice, 0)}

    // --- 1. Get and Map System (Node) Details ---
	systemBody, err := c.Get(systemURI)
	if err != nil {
		return nil, err
	}
    
    var systemData RedfishSystem
    if err := json.Unmarshal(systemBody, &systemData); err != nil {
        return nil, fmt.Errorf("failed to decode system data from %s: %w", systemURI, err)
    }

    // Map Node Data
    inv.Node = mapCommonProperties(
        systemData.CommonRedfishProperties, 
        "Node", 
        systemURI, 
        "", // Parent will be resolved later (Chassis/Rack, etc.)
    )
    
    // --- 2. Get Processors (CPUs) ---
    if cpuCollectionURI := systemData.Processors.ODataID; cpuCollectionURI != "" {
        cpuDevices, err := getCollectionDevices(c, cpuCollectionURI, "CPU", systemURI, &RedfishProcessor{})
        if err != nil {
            fmt.Printf("Warning: Failed to retrieve CPU inventory from %s: %v\n", cpuCollectionURI, err)
        } else {
            inv.CPUs = cpuDevices
        }
    }

    // --- 3. Get Memory (DIMMs) ---
    if dimmCollectionURI := systemData.Memory.ODataID; dimmCollectionURI != "" {
        dimmDevices, err := getCollectionDevices(c, dimmCollectionURI, "DIMM", systemURI, &RedfishMemory{})
        if err != nil {
            fmt.Printf("Warning: Failed to retrieve DIMM inventory from %s: %v\n", dimmCollectionURI, err)
        } else {
            inv.DIMMs = dimmDevices
        }
    }

	return inv, nil
}

// getCollectionDevices retrieves a collection, iterates over members, and maps them to DiscoveredDevice.
// componentTypeExample is an empty struct pointer (*RedfishProcessor, *RedfishMemory) used for typing.
func getCollectionDevices(c *RedfishClient, collectionURI, deviceType, parentURI string, componentTypeExample interface{}) ([]DiscoveredDevice, error) {
	var devices []DiscoveredDevice

	collectionBody, err := c.Get(collectionURI)
	if err != nil {
		return nil, err
	}

	var collection RedfishCollection
	if err := json.Unmarshal(collectionBody, &collection); err != nil {
		return nil, fmt.Errorf("failed to decode collection from %s: %w", collectionURI, err)
	}

	for _, member := range collection.Members {
		memberBody, err := c.Get(member.ODataID)
		if err != nil {
			fmt.Printf("Warning: Failed to get member %s: %v\n", member.ODataID, err)
			continue
		}

		// Create a new instance of the correct component type for unmarshaling
        // We rely on the CommonRedfishProperties being the first embedded field
		component := reflect.New(reflect.TypeOf(componentTypeExample).Elem()).Interface()

		if err := json.Unmarshal(memberBody, &component); err != nil {
			fmt.Printf("Warning: Failed to unmarshal component %s: %v\n", member.ODataID, err)
			continue
		}
        
        // Use reflection to access the CommonRedfishProperties, which is the 0th field.
        // This is necessary because the component (Processor/Memory) struct is generic here.
		rfProps := reflect.ValueOf(component).Elem().Field(0).Interface().(CommonRedfishProperties)

		devices = append(devices, mapCommonProperties(rfProps, deviceType, member.ODataID, parentURI))
	}
	
	return devices, nil
}
