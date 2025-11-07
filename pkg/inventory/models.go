package inventory

// Note: This is an internal representation for gathering, before posting.
type DiscoveredDevice struct {
	DeviceType   string
	Manufacturer string
	PartNumber   string
	SerialNumber string
	ParentID     string
	RedfishURI   string
}

// CommonRedfishProperties contains the fields required by the Device model,
// found in various Redfish resources (System, Processor, Memory).
type CommonRedfishProperties struct {
	// Note: We use the actual JSON keys, which start with capital letters in Redfish.
	Manufacturer string `json:"Manufacturer,omitempty"`
	Model        string `json:"Model,omitempty"`      // Often the part number/product name
	PartNumber   string `json:"PartNumber,omitempty"`
	SerialNumber string `json:"SerialNumber,omitempty"`
}

// RedfishSystem defines the structure for a System resource (the Node).
type RedfishSystem struct {
	CommonRedfishProperties // Embeds the common fields
    
    // Links to collections needed for traversal
	Processors struct {
		ODataID string `json:"@odata.id"`
	} `json:"Processors"`
	Memory struct {
		ODataID string `json:"@odata.id"`
	} `json:"Memory"`
    // You could include other collections like NetworkInterfaces, etc., here.
}

// RedfishProcessor defines the structure for a Processor resource (the CPU).
type RedfishProcessor struct {
	CommonRedfishProperties // Embeds the common fields
}

// RedfishMemory defines the structure for a Memory resource (the DIMM).
type RedfishMemory struct {
	CommonRedfishProperties // Embeds the common fields
    // Memory resources sometimes use 'PartNumber' on different keys, 
    // but the spec usually provides PartNumber/Manufacturer.
}