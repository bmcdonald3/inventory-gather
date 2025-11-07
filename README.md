# redfish-inventory-collector

The `redfish-inventory-collector` is a command-line tool designed to populate an OpenCHAMI-compliant hardware inventory API by actively discovering Field Replaceable Units (FRUs) from a remote machine's Baseboard Management Controller (BMC) via Redfish.

This tool enforces the core inventory contract: it discovers devices, transforms the Redfish hierarchy into UUID-linked parent/child relationships, and submits the collected data to the target API using the two-step Fabrica workflow (create resource envelope, then update status).

## Core Functionality

The collector executes the following workflow:

1. **Redfish Connection:** Connects via HTTPS using basic authentication, skipping certificate verification (self-signed certs).

2. **Hierarchical Traversal:** Navigates the Redfish `/redfish/v1/Systems` collection to discover the Node, Processors (CPUs), and Memory (DIMMs).

3. **Data Mapping:** Extracts the core `Manufacturer`, `PartNumber`, and `SerialNumber` fields from heterogeneous Redfish JSON responses.

4. **API Posting:** Posts the Node first, captures its API-assigned UUID, and uses that UUID to correctly populate the `parentID` field for all child devices before posting them.

## Prerequisites

To run this tool, you must have:

1. **Go Runtime:** Go 1.20+ installed.

2. **Inventory API:** The Fabrica-generated OpenCHAMI Inventory API must be running and accessible at `http://localhost:8081`.

3. **Target BMC:** An accessible BMC device running a Redfish service (e.g., `172.24.0.2`). The tool is configured for Basic Authentication using hardcoded credentials.

## Usage

Run the collector from the project root using `go run` and provide the target BMC IP via the `--ip` flag.

**Command Syntax:**

```bash
go run ./cmd/gather/main.go --ip <BMC_IP_ADDRESS>
```

**Example Execution:**

```bash
go run ./cmd/gather/main.go --ip 172.24.0.2
```

## Verification and Results

The tool successfully connected to the Redfish endpoint at `172.24.0.2`, discovered 7 devices, and posted them hierarchically to the inventory API.

### Execution Log (Verified Success)

The following output confirms that all steps, including authentication, discovery, and hierarchical API posting, completed successfully:

```
Starting inventory collection for BMC IP: 172.24.0.2
Redfish Client Test: Successfully connected to https://172.24.0.2/redfish/v1. Response size: 981 bytes.
Redfish Discovery Complete: Found 7 total devices.
-> Creating resource envelope for Node-QSBP82909274 (Node)...
-> Updating status for Node-QSBP82909274 (UID: dev-741037d9)...
-> Successfully posted parent device dev-741037d9
-> Creating resource envelope for CPU--1 (CPU)...
-> Updating status for CPU--1 (UID: dev-fb78d886)...
-> Successfully posted child device dev-fb78d886
... (5 more child devices posted) ...
Inventory collection and posting completed successfully.
```

### API Data Structure Example

This section illustrates the complete JSON structure of the **Node** device (`Node-QSBP82909274`) posted by this tool, demonstrating how the Redfish data is mapped into the Fabrica resource envelope.

```json
{
  "apiVersion": "v1",
  "kind": "Device",
  "schemaVersion": "v1",
  "metadata": {
    "name": "Node-QSBP82909274",
    "uid": "dev-741037d9",
    "createdAt": "2025-11-06T19:46:26.92214513-06:00",
    "updatedAt": "2025-11-06T19:46:26.923852596-06:00"
  },
  "spec": {},
  "status": {
    "deviceType": "Node",
    "manufacturer": "Intel Corporation",
    "partNumber": "102072300",
    "serialNumber": "QSBP82909274",
    "properties": {
      "redfish_uri": "/Systems/QSBP82909274"
    }
  }
}
```

### Data Posted to API

The successful posting included the following devices, demonstrating correct Redfish data extraction and UUID resolution for the `parentID` field:

| Device Type | API UID (Example) | Core Data | ParentID (Resolved UUID) | 
 | ----- | ----- | ----- | ----- | 
| **Node** | `dev-741037d9` | Manufacturer: Intel Corporation, Serial: QSBP82909274 | (Empty, Top-Level) | 
| **CPU** | `dev-fb78d886` | Manufacturer: Intel(R) Corporation, PartNumber: Intel Xeon processor | `dev-741037d9` | 
| **DIMM** | `dev-1d04a77d` | Manufacturer: Hynix, Serial: 3128C51A | `dev-741037d9` | 
| **DIMM** | `dev-cdbbc1c4` | Manufacturer: Hynix, Serial: 10CD71D4 | `dev-741037d9` |