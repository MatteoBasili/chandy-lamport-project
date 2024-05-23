package utils_test

import (
	"os"
	"reflect"
	"sdccProject/src/utils"
	"testing"
)

func TestReadConfig(t *testing.T) {
	// Before running the test, create a temporary configuration file
	// with known data to be read by the ReadConfig function
	testConfigData := `{
        "nodes": [
            {
                "idx": 1,
                "name": "node1",
                "ip": "127.0.0.1",
                "port": 8080,
                "appPort": 9001
            },
            {
                "idx": 2,
                "name": "node2",
                "ip": "127.0.0.1",
                "port": 8081,
                "appPort": 9002
            }
        ],
        "initialBalance": 100,
        "sendAttempts": 3
    }`

	// Write the data to the temporary file for the test
	tempConfigFile := "temp_config.json"
	err := os.WriteFile(tempConfigFile, []byte(testConfigData), 0644)
	if err != nil {
		t.Fatalf("Error creating temporary config file: %v", err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatalf("Error removing temporary config file: %v", err)
		}
	}(tempConfigFile) // Remove the temporary file at the end of the test

	// Read the configuration from the temporary file using the ReadConfig function
	netLayout := utils.ReadConfig(tempConfigFile)

	// Set up the expected NetLayout
	expectedNetLayout := utils.NetLayout{
		Nodes: []utils.Node{
			{Idx: 1, Name: "node1", IP: "127.0.0.1", Port: 8080, AppPort: 9001},
			{Idx: 2, Name: "node2", IP: "127.0.0.1", Port: 8081, AppPort: 9002},
		},
		InitialBalance: 100,
		SendAttempts:   3,
	}

	// Compare the NetLayout read from the file with the expected one
	if !reflect.DeepEqual(netLayout, expectedNetLayout) {
		t.Errorf("ReadConfig() returned unexpected result.\nExpected: %+v\nGot: %+v", expectedNetLayout, netLayout)
	}
}

func TestReadConfig_FileNotFound(t *testing.T) {
	// Test scenario where the config file is not found
	// Expected behavior: ReadConfig should panic with an error
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected ReadConfig to panic with file not found error, but it did not panic")
		}
	}()
	utils.ReadConfig("nonexistent_file.json")
}

func TestReadConfig_InvalidJSON(t *testing.T) {
	// Test scenario where the config file contains invalid JSON
	// Expected behavior: ReadConfig should panic with an error
	invalidJSON := `{"nodes": []}}` // Invalid JSON with extra closing bracket
	tempConfigFile := "invalid_config.json"
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatalf("Error removing temporary config file: %v", err)
		}
	}(tempConfigFile)
	err := os.WriteFile(tempConfigFile, []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatalf("Error creating temporary config file: %v", err)
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected ReadConfig to panic with invalid JSON error, but it did not panic")
		}
	}()
	utils.ReadConfig(tempConfigFile)
}
