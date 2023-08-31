package main

import (
    "bufio"
    "fmt"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "go.uber.org/zap"
    "marketplace-platform/pkg/logger"
    "os"
    "os/signal"
    "strings"
    "syscall"
)

func main() {
    log := logger.NewLogger()

    // awsSession :=
    session.Must(session.NewSession(&aws.Config{
        Region:   aws.String("eu-west-1"),
        Endpoint: aws.String("http://dynamodb-local:8000"),
    }))

    // Create a new DynamoDB service client

    // Create a channel to receive the SIGTERM signal
    c := make(chan os.Signal, 1)
    signal.Notify(c, syscall.SIGTERM)

    // Create a goroutine to handle the SIGTERM signal
    go func() {
        <-c
        fmt.Println("Received SIGTERM, exiting...")
        os.Exit(0)
    }()

    // Create a new reader to read input from the command line
    reader := bufio.NewReader(os.Stdin)

    for {
        // Print the prompt
        fmt.Print("# ")

        // Read the input from the command line
        input, _ := reader.ReadString('\n')

        // Remove the newline character from the input
        input = strings.TrimSpace(input)

        // Split the input into arguments
        args := strings.Split(input, " ")

        // Check the command and execute the corresponding function
        switch args[0] {
        case "REGISTER":
            // Execute the REGISTER command
            register(args[1:])
        case "CREATE_LISTING":
            // Execute the CREATE_LISTING command
            createListing(args[1:])
        case "GET_LISTING":
            // Execute the GET_LISTING command
            getListing(args[1:])
        case "GET_CATEGORY":
            // Execute the GET_CATEGORY command
            getCategory(args[1:])
        case "GET_TOP_CATEGORY":
            // Execute the GET_TOP_CATEGORY command
            getTopCategory(args[1:])
        case "DELETE_LISTING":
            // Execute the DELETE_LISTING command
            deleteListing(args[1:])
        default:
            log.Error("Unknown command", zap.String("command", args[0]))
            // fmt.Println("Unknown command", args[0])
        }
    }
}

// Define your functions for each command here
func register(args []string) {
    // Implement the REGISTER command here
}

func createListing(args []string) {
    // Implement the CREATE_LISTING command here
}

func getListing(args []string) {
    // Implement the GET_LISTING command here
}

func getCategory(args []string) {
    // Implement the GET_CATEGORY command here
}

func getTopCategory(args []string) {
    // Implement the GET_TOP_CATEGORY command here
}

func deleteListing(args []string) {
    // Implement the DELETE_LISTING command here
}
