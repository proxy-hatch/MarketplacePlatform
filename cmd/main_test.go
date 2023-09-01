package main

import (
    "bytes"
    "fmt"
    "os"
    "os/exec"
    "testing"
)

func TestIntegration(t *testing.T) {
    // Start the program as a separate process
    cmd := exec.Command("./main")
    stdin, err := cmd.StdinPipe()
    if err != nil {
        t.Fatalf("could not get stdin pipe: %v", err)
    }
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        t.Fatalf("could not get stdout pipe: %v", err)
    }
    err = cmd.Start()
    if err != nil {
        t.Fatalf("could not start command: %v", err)
    }

    // Define the test cases
    testCases := []struct {
        input    string
        expected string
    }{
        {"REGISTER user1\n", "Success\n"},
        {"CREATE_LISTING user1 'Phone model 8' 'Black color, brand new' 1000 'Electronics'\n", "100001\n"},
        {"GET_LISTING user1 100001\n", "Phone model 8|Black color, brand new|1000|2019-02-22 12:34:56|Electronics|user1\n"},
        // {"CREATE_LISTING user1 'Black shoes' 'Training shoes' 100 'Sports'\n", "100002\n"},
        // {"REGISTER user2\n", "Success\n"},
        // {"REGISTER user2\n", "Error - user already existing\n"},
        // {"CREATE_LISTING user2 'T-shirt' 'White color' 20 'Sports'\n", "100003\n"},
        // {"GET_LISTING user1 100003\n", "Error - category not found\n"},
        // {"GET_CATEGORY user1 'Fashion' sort_time asc\n", "Error - category not found\n"},
        // {"GET_CATEGORY user1 'Sports' sort_time dsc\n", "T-shirt|White color|20|2019-02-22 12:34:58|Sports|user2\n"},
        // {"GET_CATEGORY user1 'Sports' sort_price dsc\n", "Black shoes|Training shoes|100|2019-02-22 12:34:57|Sports|user1\n"},
        // {"GET_TOP_CATEGORY user1\n", "Black shoes|Training shoes|100|2019-02-22 12:34:57|Sports|user1\n"},
        // {"DELETE_LISTING user1 100003\n", "Error - listing owner mismatch\n"},
        // {"DELETE_LISTING user2 100003\n", "Success\n"},
        // {"GET_TOP_CATEGORY user2\n", "Sports\n"},
        // {"DELETE_LISTING user1 100002\n", "Success\n"},
        // {"GET_TOP_CATEGORY user1\n", "Electronics\n"},
        // {"GET_TOP_CATEGORY user3\n", "Error - unknown user\n"},
    }

    // Create a buffer to hold the output
    var output bytes.Buffer

    // Execute each test case
    fmt.Println("====Start test case: ", testCases)

    for _, tc := range testCases {
        // Write the input to stdin
        _, err = stdin.Write([]byte(tc.input + "\n"))
        if err != nil {
            t.Fatalf("could not write to stdin: %v", err)
        }

        // Read the output from stdout
        buf := make([]byte, len(tc.expected))
        _, err = stdout.Read(buf)
        if err != nil {
            t.Fatalf("could not read from stdout: %v", err)
        }
        output.Write(buf)

        // Compare the output to the expected output
        if output.String() != tc.expected {
            fmt.Println("====Test case failed: ")
            fmt.Println("Input: ", tc.input)
            t.Fatalf("expected %q, got %q", tc.expected, output.String())
        }

        fmt.Println("====Test case passed: ")
        fmt.Println("Input: ", tc.input)
        fmt.Println("Output: ", output.String())

        // Clear the output buffer for the next test case
        output.Reset()
    }

    // Send a SIGTERM signal to the process to stop it
    err = cmd.Process.Signal(os.Interrupt)
    if err != nil {
        t.Fatalf("could not send SIGTERM signal to process: %v", err)
    }

    // Wait for the process to exit
    err = cmd.Wait()
    if err != nil && err.Error() != "signal: interrupt" {
        t.Fatalf("process did not exit cleanly: %v", err)
    }

}
