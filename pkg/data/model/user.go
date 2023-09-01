package model

import (
    "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
    "marketplace-platform/pkg/constant"
    "strconv"
)

type User struct {
    Username string `dynamodbav:"Username"`
}

func (u User) Validate() error {
    return validate.Struct(u)
}

func (u User) DdbMarshalMap() (map[string]types.AttributeValue, error) {
    av, err := attributevalue.MarshalMap(u)
    if err != nil {
        return av, err
    }

    // fix the partition key
    av[constant.ListingTablePartitionKeyName] = &types.AttributeValueMemberN{
        Value: strconv.Itoa(constant.UserRootRecordPartitionKey),
    }

    return av, nil
}
