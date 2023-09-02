package model

import (
    "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
    "marketplace-platform/pkg/constant"
    "strconv"
    "time"
)

type Listing struct {
    ListingId   int       `dynamodbav:"ListingId" validate:"gt=100000"` // partition key
    Username    string    `dynamodbav:"Username" validate:"required"`   // sort key
    Title       string    `dynamodbav:"Title" validate:"required"`
    Description string    `dynamodbav:"Description"`
    Price       int       `dynamodbav:"Price" validate:"gte=0"`
    Category    string    `dynamodbav:"Category" validate:"required"`
    CreatedAt   time.Time `dynamodbav:"CreatedAt,unixtime"`
}

func NewListing(listingId int, username string, title string, description string, price int, category string) (Listing, error) {
    listing := Listing{
        ListingId:   listingId,
        Username:    username,
        Title:       title,
        Description: description,
        Price:       price,
        Category:    category,
        CreatedAt:   time.Now(),
    }

    err := validate.Struct(listing)
    if err != nil {
        return Listing{}, err
    }

    return listing, nil
}

func (l Listing) Validate() error {
    return validate.Struct(l)
}

func (l Listing) DdbMarshalMap() (map[string]types.AttributeValue, error) {
    av, err := attributevalue.MarshalMap(l)
    if err != nil {
        return av, err
    }

    // Add GSI attribute
    av[constant.ListingIdIndexPartitionKeyName] = &types.AttributeValueMemberN{
        Value: strconv.Itoa(constant.ListingIdIndexPartitionKey),
    }

    return av, nil
}

func (l Listing) String() string {
    // print listing in the format:
    // "<title>|<description>|<price>|<created_at>|<category>|<username>"
    return l.Title + "|" + l.Description + "|" + strconv.Itoa(l.Price/100) + "|" + l.CreatedAt.Format("2006-01-02 15:04:05") + "|" + l.Category + "|" + l.Username
}
