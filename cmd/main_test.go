package main

import (
    "bytes"
    "fmt"
    "os"
    "os/exec"
    "strings"
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
        // authentication errors
        {"CREATE_LISTING user1 'Phone model 8' 'Black color, brand new' 1000 'Electronics'\n", "Error - unknown user\n"},
        {"DELETE_LISTING user1 100001\n", "Error - unknown user\n"},
        {"GET_LISTING user1 100001\n", "Error - unknown user\n"},
        {"GET_CATEGORY user1 'Electronics'\n", "Error - unknown user\n"},
        {"GET_TOP_CATEGORY user1\n", "Error - unknown user\n"},

        // input validation errors
        {"REGISTER \n", "Error - invalid number of arguments\n"},
        {"CREATE_LISTING user1 'Phone model 8' 'Black color, brand new' 1000 \n", "Error - invalid number of arguments\n"},
        {"DELETE_LISTING user1 \n", "Error - invalid number of arguments\n"},
        {"GET_LISTING user1 \n", "Error - invalid number of arguments\n"},
        {"GET_CATEGORY user1 \n", "Error - invalid number of arguments\n"},
        {"GET_CATEGORY user1 'Electronics' sort_price \n", "Error - invalid number of arguments\n"},
        {"GET_TOP_CATEGORY \n", "Error - invalid number of arguments\n"},

        // happy path
        {"REGISTER user1\n", "Success\n"},
        {"CREATE_LISTING user1 'Phone model 8' 'Black color, brand new' 1000 'Electronics'\n", "100001\n"},
        {"GET_LISTING user1 100001\n", "Phone model 8|Black color, brand new|1000|2019-02-22 12:34:56|Electronics|user1\n"},

        // title cannot be empty
        {"CREATE_LISTING user1 '' 'Black color, brand new' 1000 'Electronics'\n", "Error - invalid input\n"},
        // category cannot be empty
        {"CREATE_LISTING user1 'Phone model 8' 'Black color, brand new' 1000 ''\n", "Error - invalid input\n"},
        // listing ID validation error
        {"GET_LISTING user1 100xxx\n", "Error - invalid input\n"},
        // listing not found
        {"GET_LISTING user1 900001\n", "Error - not found\n"},

        // unquoted non-whitespace string field should still work
        {"CREATE_LISTING user1 'Black shoes' 'Training shoes' 100 Sports\n", "100002\n"},

        {"REGISTER user2\n", "Success\n"},
        {"REGISTER user2\n", "Error - user already existing\n"},
        {"CREATE_LISTING user2 'T-shirt' 'White color' 20 'Sports'\n", "100003\n"},
        {"GET_LISTING user1 100003\n", "T-shirt|White color|20|2019-02-22 12:34:58|Sports|user2\n"},

        // invalid sort key and order
        {"GET_CATEGORY user1 'Electronics' zzzsort_price dsc \n", "Error - invalid sort key\n"},
        {"GET_CATEGORY user1 'Electronics' sort_price zzzdsc \n", "Error - invalid sort order\n"},

        // category not found
        {"GET_CATEGORY user1 'Fashion' sort_time asc\n", "Error - category not found\n"},

        // unquoted number category should still work
        {"GET_CATEGORY user1 123 sort_time asc\n", "Error - category not found\n"},

        {"GET_CATEGORY user1 'Sports' sort_time dsc\n", "T-shirt|White color|20|2019-02-22 12:34:58|Sports|user2\nBlack shoes|Training shoes|100|2019-02-22 12:34:57|Sports|user1\n"},
        {"GET_CATEGORY user1 'Sports' sort_price dsc\n", "Black shoes|Training shoes|100|2019-02-22 12:34:57|Sports|user1\nT-shirt|White color|20|2019-02-22 12:34:58|Sports|user2\n"},

        // default sort order is sort_time dsc
        {"GET_CATEGORY user1 'Sports'\n", "T-shirt|White color|20|2019-02-22 12:34:58|Sports|user2\nBlack shoes|Training shoes|100|2019-02-22 12:34:57|Sports|user1\n"},

        {"GET_TOP_CATEGORY user1\n", "Sports\n"},
        {"DELETE_LISTING user1 100003\n", "Error - listing owner mismatch\n"},
        {"DELETE_LISTING user2 100003\n", "Success\n"},
        {"GET_TOP_CATEGORY user2\n", "Sports\n"},
        {"DELETE_LISTING user1 100002\n", "Success\n"},
        {"GET_TOP_CATEGORY user1\n", "Electronics\n"},

        // create listing with empty description should still work
        {"CREATE_LISTING user1 'Black shoes' '' 100 Sports\n", "100002\n"},

        // listing ID validation error
        {"DELETE_LISTING user1 100xxx\n", "Error - invalid input\n"},
    }

    // Create a buffer to hold the output
    var outputBuffer bytes.Buffer

    // Execute each test case
    fmt.Println("====Begin Test====")

    for _, tc := range testCases {
        // Write the input to stdin
        _, err = stdin.Write([]byte(tc.input + "\n"))
        if err != nil {
            t.Fatalf("could not write to stdin: %v", err)
        }

        // Read the outputBuffer from stdout
        buf := make([]byte, len(tc.expected))
        _, err = stdout.Read(buf)
        if err != nil {
            t.Fatalf("could not read from stdout: %v", err)
        }
        outputBuffer.Write(buf)

        // Compare the outputBuffer with the expected outputBuffer
        // Ignore the timestamp field for listings
        outputFields := strings.Split(outputBuffer.String(), "|")
        expectedFields := strings.Split(tc.expected, "|")
        if isMultilineListings(len(expectedFields)) && len(outputFields) == len(expectedFields) {
            for i := range outputFields {
                if isTimestampField(i) {
                    continue
                }
                if outputFields[i] != expectedFields[i] {
                    fmt.Println("\n====Test case failed: ")
                    fmt.Println(fmt.Sprintf("Field %d: expected %q, got %q", i, expectedFields[i], outputFields[i]))
                    fmt.Println("Input: ", tc.input)
                    t.Fatalf("expected %q, got %q", tc.expected, outputBuffer.String())
                }
            }
        } else if outputBuffer.String() != tc.expected {
            fmt.Println("\n====Test case failed: ")
            fmt.Println("Input: ", tc.input)
            t.Fatalf("expected %q, got %q", tc.expected, outputBuffer.String())
        }

        fmt.Println("\n====Test case passed: ")
        fmt.Println("Input: ", tc.input)
        fmt.Println("Output: ", outputBuffer.String())

        // Clear the outputBuffer buffer for the next test case
        outputBuffer.Reset()
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

func isTimestampField(index int) bool {
    return index == 3 || index == 8
}

func isMultilineListings(fieldCount int) bool {
    return fieldCount >= 6 && (fieldCount-6)%5 == 0
}
