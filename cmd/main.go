package main

import (
    "bufio"
    "fmt"
    "github.com/go-playground/validator/v10"
    "go.uber.org/zap"
    "marketplace-platform/pkg/data/ddb"
    "marketplace-platform/pkg/data/model"
    "marketplace-platform/pkg/data/model/enum"
    "marketplace-platform/pkg/exception"
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
    } else {
        log.Info("Listing table does not exist before initialization")
    }

    log.Info("Initializing empty Listing table")
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

        log.Info("Received command: " + cmd)
        log.Info("Received arguments: " + strings.Join(args, ", "))

        // Check the command and execute the corresponding function. Additional arguments ignored.
        switch cmd {
        case "REGISTER":
            if len(args) < 1 {
                fmt.Println("Error - invalid number of arguments")
                continue
            }
            username := args[0]
            register(username)
        case "CREATE_LISTING":
            if len(args) < 5 {
                fmt.Println("Error - invalid number of arguments")
                continue
            }
            username := args[0]
            title := args[1]
            description := args[2]
            price := args[3]
            category := args[4]

            user, err := authUser(username)
            if err != nil {
                log.Errorf("Error authenticating user '%s': %v", username, err)
                continue
            }
            if user == nil {
                fmt.Println("Error - unknown user")
                continue
            }

            createListing(username, title, description, price, category)
        case "GET_LISTING":
            if len(args) < 2 {
                fmt.Println("Error - invalid number of arguments")
                continue
            }
            username := args[0]
            // convert listingId to int
            listingId, err := strconv.Atoi(args[1])
            if err != nil {
                log.Errorf("Error converting listingId '%s' to int: %v", args[1], err)
                fmt.Println("Error - invalid input")
                continue
            }

            user, err := authUser(username)
            if err != nil {
                log.Errorf("Error authenticating user '%s': %v", username, err)
                fmt.Println("Error - internal server error")
                continue
            }
            if user == nil {
                fmt.Println("Error - unknown user")
                continue
            }

            getListing(listingId)
        case "GET_CATEGORY":
            if len(args) < 2 || len(args) == 3 {
                fmt.Println("Error - invalid number of arguments")
                continue
            }
            username := args[0]
            category := args[1]

            user, err := authUser(username)
            if err != nil {
                log.Errorf("Error authenticating user '%s': %v", username, err)
                fmt.Println("Error - internal server error")
                continue
            }
            if user == nil {
                fmt.Println("Error - unknown user")
                continue
            }
            //
            // // DEBUG
            // fmt.Println("args: ", args)

            if len(args) >= 4 {
                sortKeyStr := args[2]
                sortOrderStr := args[3]

                sortBy, err := parseSortBy(sortKeyStr)
                if err != nil {
                    log.Errorf("Error parsing sort key '%s': %v", sortKeyStr, err)
                    fmt.Println("Error - invalid sort key")
                    continue
                }
                orderBy, err := parseOrderBy(sortOrderStr)
                if err != nil {
                    log.Errorf("Error parsing sort order '%s': %v", sortOrderStr, err)
                    fmt.Println("Error - invalid sort order")
                    continue
                }

                getCategory(category, &sortBy, &orderBy)
            } else {
                getCategory(category, nil, nil)
            }

        case "GET_TOP_CATEGORY":
            if len(args) < 1 {
                fmt.Println("Error - invalid number of arguments")
                continue
            }
            username := args[0]

            user, err := authUser(username)
            if err != nil {
                log.Errorf("Error authenticating user '%s': %v", username, err)
                fmt.Println("Error - internal server error")
                continue
            }
            if user == nil {
                fmt.Println("Error - unknown user")
                continue
            }

            getTopCategory()

        case "DELETE_LISTING":
            if len(args) < 2 {
                fmt.Println("Error - invalid number of arguments")
                continue
            }
            username := args[0]
            listingId, err := strconv.Atoi(args[1])
            if err != nil {
                log.Errorf("Error converting listingId '%s' to int: %v", args[1], err)
                fmt.Println("Error - invalid input")
                continue
            }

            user, err := authUser(username)
            if err != nil {
                log.Errorf("Error authenticating user '%s': %v", username, err)
                fmt.Println("Error - internal server error")
                continue
            }
            if user == nil {
                fmt.Println("Error - unknown user")
                continue
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
    priceInt, err := util.ConvertPriceStringToInt(price)
    if err != nil {
        log.Errorf("Error converting price '%s' to int: %v", price, err)
        fmt.Println("Error - invalid price")
        return
    }
    listing, err := dao.PutListing(username, title, description, priceInt, category)
    if err != nil {
        log.Errorf("Error creating listing: %v", err)

        if _, ok := err.(validator.ValidationErrors); ok {
            fmt.Println("Error - invalid input")
        } else {
            fmt.Println("Error - internal server error")
        }
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

func getCategory(category string, sortKey *enum.SortBy, sortOrder *enum.OrderBy) {
    var listings []model.Listing
    var err error
    if sortKey == nil && sortOrder == nil {
        // default sort by descending created time
        listings, err = dao.GetCategory(category, enum.SortByCreatedAt, enum.OrderByDescending)
    } else {
        listings, err = dao.GetCategory(category, *sortKey, *sortOrder)
    }

    if err != nil {
        log.Errorf("Error getting category '%s': %v", category, err)
        fmt.Println("Error - internal server error")
        return
    }

    if len(listings) == 0 {
        fmt.Println("Error - category not found")
    } else {
        for _, listing := range listings {
            fmt.Println(listing)
        }
    }
}

func getTopCategory() {
    category, err := dao.GetTopCategory()
    if err != nil {
        log.Errorf("Error getting top category: %v", err)
        fmt.Println("Error - internal server error")
        return
    }
    if category == "" {
        fmt.Println("Error - no category found")
    } else {
        fmt.Println(category)
    }
}

func deleteListing(username string, listingId int) {
    err := dao.DeleteListing(username, listingId)
    if err != nil {
        // handle exception.OwnershipMismatchException and exception.ListingNotFoundException
        log.Errorf("Error deleting listing '%d': %v", listingId, err)
        switch err.(type) {
        case *exception.OwnershipMismatchException:
            fmt.Println("Error - listing owner mismatch")
        case *exception.ListingDoesNotExistException:
            fmt.Println("Error - listing does not exist")
        default:
            fmt.Println("Error - internal server error")
        }
        return
    }

    fmt.Println("Success")
}

// authUser determines if the user is authorized to perform the action
// Returns the user if authorized, otherwise return nil
func authUser(username string) (*model.User, error) {
    user, err := dao.GetUser(username)
    if err != nil {
        log.Debugf("Error getting user '%s': %v", username, err)
        return nil, err
    }
    if user == nil {
        log.Debugf("User '%s' does not exist", username)
        return nil, nil
    }
    log.Debugf("Retrieved user '%+v'", user)
    log.Debugf("User with username '%s' exist", username)
    return user, nil
}

func parseSortBy(s string) (enum.SortBy, error) {
    switch s {
    case "sort_time":
        return enum.SortByCreatedAt, nil
    case "sort_price":
        return enum.SortByPrice, nil
    default:
        return 0, fmt.Errorf("invalid SortBy value: %s", s)
    }
}

func parseOrderBy(s string) (enum.OrderBy, error) {
    switch s {
    case "dsc":
        return enum.OrderByDescending, nil
    case "asc":
        return enum.OrderByAscending, nil
    default:
        return 0, fmt.Errorf("invalid OrderBy value: %s", s)
    }
}
