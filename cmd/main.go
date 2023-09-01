package main

import (
    "bufio"
    "fmt"
    "go.uber.org/zap"
    "marketplace-platform/pkg/data/ddb"
    "marketplace-platform/pkg/data/model"
    "marketplace-platform/pkg/logger"
    "marketplace-platform/pkg/util"
    "os"
    "os/signal"
    "strconv"
    "strings"
    "syscall"
)

var (
    log = logger.NewLogger()
    dao = ddb.NewDynamoDataAccess(log)
)

func main() {
    exists, err := dao.ListingTableExists()
    if err != nil {
        log.Fatalf("Error checking if listing table exists: %v", err)
        return
    }
    if exists {
        log.Info("Listing table exists before initialization. Cleaning up")
        err := dao.DeleteTable()
        if err != nil {
            log.Fatalf("Error deleting listing table: %v", err)
            return
        }
    }

    _, err = dao.CreateListingTable()
    if err != nil {
        log.Fatalf("Error creating Listing table: %v", err)
        return
    }

    exists, err = dao.ListingTableExists()
    if err != nil {
        log.Fatalf("Error checking if listing table exists: %v", err)
        return
    }
    log.Info("Empty Listing table initialized")

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
        _, err = fmt.Fprint(os.Stderr, "# ")
        if err != nil {
            log.Errorf("Error printing prompt: %v", err)
            return
        }

        input, err := reader.ReadString('\n')
        if err != nil {
            log.Errorf("Error reading input: %v", err)
            return
        }

        input = strings.TrimSpace(input)
        if input == "" {
            continue
        }
        args := util.SplitArgs(input)
        cmd := args[0]
        args = args[1:]

        // Check the command and execute the corresponding function. Additional arguments ignored.
        switch cmd {
        case "REGISTER":
            if len(args) < 1 {
                fmt.Println("Error - invalid number of arguments")
                return
            }
            username := args[0]
            register(username)
        case "CREATE_LISTING":
            if len(args) < 5 {
                fmt.Println("Error - invalid number of arguments")
                return
            }
            username := args[0]
            title := args[1]
            description := args[2]
            price := args[3]
            category := args[4]

            user, err := authUser(username)
            if err != nil {
                log.Errorf("Error authenticating user '%s': %v", username, err)
                return
            }
            if user == nil {
                fmt.Println("Error - unknown user")
                return
            }

            createListing(username, title, description, price, category)
        case "GET_LISTING":
            if len(args) < 2 {
                fmt.Println("Error - invalid number of arguments")
                return
            }
            username := args[0]
            // convert listingId to int
            listingId, err := strconv.Atoi(args[1])
            if err != nil {
                log.Errorf("Error converting listingId '%s' to int: %v", args[1], err)
                fmt.Println("Error - invalid listingId")
                return
            }

            user, err := authUser(username)
            if err != nil {
                log.Errorf("Error authenticating user '%s': %v", username, err)
                fmt.Println("Error - internal server error")
                return
            }
            if user == nil {
                fmt.Println("Error - unknown user")
                return
            }

            getListing(listingId)
        case "GET_CATEGORY":
            if len(args) < 2 || len(args) == 3 {
                fmt.Println("Error - invalid number of arguments")
                return
            }
            username := args[0]
            category := args[1]

            user, err := authUser(username)
            if err != nil {
                log.Errorf("Error authenticating user '%s': %v", username, err)
                fmt.Println("Error - internal server error")
                return
            }
            if user == nil {
                fmt.Println("Error - unknown user")
                return
            }

            if len(args) >= 4 {
                sortKey := args[2]
                sortOrder := args[3]
                getCategoryWithSort(category, sortKey, sortOrder)
            } else {
                getCategory(category)
            }
        case "GET_TOP_CATEGORY":
            if len(args) < 1 {
                fmt.Println("Error - invalid number of arguments")
                return
            }
            username := args[0]

            user, err := authUser(username)
            if err != nil {
                log.Errorf("Error authenticating user '%s': %v", username, err)
                fmt.Println("Error - internal server error")
                return
            }
            if user == nil {
                fmt.Println("Error - unknown user")
                return
            }

            getTopCategory()

        case "DELETE_LISTING":
            if len(args) < 2 {
                fmt.Println("Error - invalid number of arguments")
                return
            }
            username := args[0]
            listingId := args[1]

            user, err := authUser(username)
            if err != nil {
                log.Errorf("Error authenticating user '%s': %v", username, err)
                fmt.Println("Error - internal server error")
                return
            }
            if user == nil {
                fmt.Println("Error - unknown user")
                return
            }

            deleteListing(username, listingId)
        default:
            log.Error("Unknown command", zap.String("command", args[0]))
            fmt.Println("Unknown command", args[0])
        }
    }
}

func register(username string) {
    user, err := dao.PutUser(username)
    if err != nil {
        log.Errorf("Error registering user '%s': %v", username, err)
        fmt.Println("Error - internal server error")
        return
    }

    if user == nil {
        fmt.Println("Error - user already existing")
    } else {
        fmt.Println("Success")
    }
}

func createListing(username string, title string, description string, price string, category string) {
    // validate price is a float with 2 decimal places and convert to int
    priceInt, err := util.ConvertPriceStringToInt(price)
    if err != nil {
        log.Errorf("Error converting price '%s' to int: %v", price, err)
        fmt.Println("Error - invalid price")
        return
    }
    listing, err := dao.PutListing(username, title, description, priceInt, category)
    if err != nil {
        log.Errorf("Error creating listing: %v", err)
        fmt.Println("Error - internal server error")
        return
    }
    if listing == nil {
        fmt.Println("Error - listing already existing")
    } else {
        fmt.Println(listing.ListingId)
    }
}

func getListing(listingId int) {
    listing, err := dao.GetListing(listingId)
    if err != nil {
        log.Errorf("Error getting listing '%s': %v", listingId, err)
        fmt.Println("Error - internal server error")
        return
    }
    if listing == nil {
        fmt.Println("Error - not found")
    } else {
        fmt.Println(listing)
    }
}

func getCategory(category string) {

}

func getCategoryWithSort(category string, sortKey string, sortOrder string) {
    // Implement the GET_CATEGORY command here
}

func getTopCategory() {
    // Implement the GET_TOP_CATEGORY command here
}

func deleteListing(username string, listingId string) {
    // Implement the DELETE_LISTING command here
}

// authUser determines if the user is authorized to perform the action
// Returns the user if authorized, otherwise return nil
func authUser(username string) (*model.User, error) {
    user, err := dao.GetUser(username)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, nil
    }
    return user, nil
}
