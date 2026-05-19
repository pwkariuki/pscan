package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"pScan/scan"
	"strconv"
	"strings"
	"testing"
)

func setup(t *testing.T, hosts []string, initList bool) (string, func()) {
	// Create temp file
	tf, err := os.CreateTemp("", "pScan")
	if err != nil {
		t.Fatal(err)
	}
	tf.Close()

	// Initialize list if needed
	if initList {
		hl := &scan.HostsList{}

		for _, h := range hosts {
			hl.Add(h)
		}

		if err := hl.Save(tf.Name()); err != nil {
			t.Fatal(err)
		}
	}

	// Return temp file name and cleanup functions
	return tf.Name(), func() {
		os.Remove(tf.Name())
	}
}

func TestHostActions(t *testing.T) {
	// Define hosts for Actions test
	hosts := []string{
		"host1",
		"host2",
		"host3",
	}

	testCases := []struct {
		name           string
		args           []string
		expectedOut    string
		initList       bool
		actionFunction func(io.Writer, string, []string) error
	}{
		{
			name:           "AddAction",
			args:           hosts,
			expectedOut:    "Added host: host1\nAdded host: host2\nAdded host: host3\n",
			initList:       false,
			actionFunction: addAction,
		},
		{
			name:           "listAction",
			expectedOut:    "host1\nhost2\nhost3\n",
			initList:       true,
			actionFunction: listAction,
		},
		{
			name:           "DeleteAction",
			args:           []string{"host1", "host2"},
			expectedOut:    "Deleted host: host1\nDeleted host: host2\n",
			initList:       true,
			actionFunction: deleteAction,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup Action test
			tf, cleanup := setup(t, hosts, tc.initList)
			defer cleanup()

			// Buffer to capture Action output
			var out bytes.Buffer

			// Execute Action and capture output
			if err := tc.actionFunction(&out, tf, tc.args); err != nil {
				t.Fatalf("Expected no error, got %q\n", err)
			}

			// Test Action output
			if out.String() != tc.expectedOut {
				t.Errorf("Expected output %q, got %q\n", tc.expectedOut, out.String())
			}
		})
	}
}

func TestIntegration(t *testing.T) {
	// Define hosts for integration test
	hosts := []string{
		"host1",
		"host2",
		"host3",
	}

	// Setup integration test
	tf, cleanup := setup(t, hosts, false)
	defer cleanup()

	delHost := "host2"
	hostsEnd := []string{
		"host1",
		"host3",
	}

	// Buffer to capture Action output
	var out bytes.Buffer

	// Define expected output for all actions
	expectedOutput := ""
	for _, v := range hosts {
		expectedOutput += fmt.Sprintf("Added host: %s\n", v)
	}
	expectedOutput += strings.Join(hosts, "\n")
	expectedOutput += fmt.Sprintln()
	expectedOutput += fmt.Sprintf("Deleted host: %s\n", delHost)
	expectedOutput += strings.Join(hostsEnd, "\n")
	expectedOutput += fmt.Sprintln()
	for _, v := range hostsEnd {
		expectedOutput += fmt.Sprintf("%s: Host not found\n", v)
		expectedOutput += fmt.Sprintln()
	}

	// Add hosts to the list
	if err := addAction(&out, tf, hosts); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	// List hosts
	if err := listAction(&out, tf, nil); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	// Delete host2
	if err := deleteAction(&out, tf, []string{delHost}); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	// List hosts after deletion
	if err := listAction(&out, tf, nil); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	// Scan hosts
	if err := scanAction(&out, tf, nil); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	// Test integration output
	if out.String() != expectedOutput {
		t.Errorf("Expected output %q, got %q\n", expectedOutput, out.String())
	}
}

func TestScanAction(t *testing.T) {
	// Define hosts for scan test
	hosts := []string{
		"localhost",
		"unknownhostoutthere",
	}

	// Setup scan test
	tf, cleanup := setup(t, hosts, true)
	defer cleanup()

	ports := []int{}

	// Init ports, 1 open, 1 closed
	for i := 0; i < 2; i++ {
		ln, err := net.Listen("tcp", net.JoinHostPort("localhost", "0"))
		if err != nil {
			t.Fatal(err)
		}

		defer ln.Close()

		_, portStr, err := net.SplitHostPort(ln.Addr().String())
		if err != nil {
			t.Fatal(err)
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			t.Fatal(err)
		}

		ports = append(ports, port)

		if i == 1 {
			ln.Close()
		}
	}

	// Define expected output for scan action
	expectedOutput := fmt.Sprintln("localhost:")
	expectedOutput += fmt.Sprintf("\t%d: open\n", ports[0])
	expectedOutput += fmt.Sprintf("\t%d: closed\n", ports[1])
	expectedOutput += fmt.Sprintln()
	expectedOutput += fmt.Sprintln("unknownhostoutthere: Host not found")
	expectedOutput += fmt.Sprintln()

	// Buffer to capture scan output
	var out bytes.Buffer

	// Execute scan and capture output
	if err := scanAction(&out, tf, ports); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	// Test scan output
	if out.String() != expectedOutput {
		t.Errorf("Expected output %q, got %q\n", expectedOutput, out.String())
	}
}
