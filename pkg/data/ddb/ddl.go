package ddb

import (
    "context"
    "errors"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
    "marketplace-platform/pkg/constant"
    "time"
)

// ListingTableExists checks if the Listing table exists
func (d DynamoDataAccess) ListingTableExists() (bool, error) {
    _, err := d.client.DescribeTable(
        context.TODO(), &dynamodb.DescribeTableInput{TableName: aws.String(constant.TableName)},
    )
    if err != nil {
        var notFoundEx *types.ResourceNotFoundException
        if errors.As(err, &notFoundEx) {
            return false, nil
        } else {
            return false, err
        }
    }

    return true, nil
}

// CreateListingTable creates the Listing table
func (d DynamoDataAccess) CreateListingTable() (*types.TableDescription, error) {
    // ignored for DDB Local
    provisionedThroughput := &types.ProvisionedThroughput{
        ReadCapacityUnits:  aws.Int64(5),
        WriteCapacityUnits: aws.Int64(5),
    }

    table, err := d.client.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
        AttributeDefinitions: []types.AttributeDefinition{{
            AttributeName: aws.String(constant.ListingTablePartitionKeyName),
            AttributeType: types.ScalarAttributeTypeN,
        }, {
            AttributeName: aws.String(constant.ListingTableSortKeyName),
            AttributeType: types.ScalarAttributeTypeS,
        }, {
            AttributeName: aws.String("Category"),
            AttributeType: types.ScalarAttributeTypeS,
        }, {
            AttributeName: aws.String("CategoryCount"),
            AttributeType: types.ScalarAttributeTypeN,
        }, {
            AttributeName: aws.String("Price"),
            AttributeType: types.ScalarAttributeTypeN,
        }, {
            AttributeName: aws.String("CreatedAt"),
            AttributeType: types.ScalarAttributeTypeN,
        }, {
            AttributeName: aws.String(constant.ListingIdIndexPartitionKeyName),
            AttributeType: types.ScalarAttributeTypeN,
        }},

        KeySchema: []types.KeySchemaElement{{
            AttributeName: aws.String(constant.ListingTablePartitionKeyName),
            KeyType:       types.KeyTypeHash,
        }, {
            AttributeName: aws.String(constant.ListingTableSortKeyName),
            KeyType:       types.KeyTypeRange,
        }},

        LocalSecondaryIndexes: []types.LocalSecondaryIndex{{
            IndexName: aws.String(constant.CategoryCountIndex),
            KeySchema: []types.KeySchemaElement{{
                AttributeName: aws.String(constant.ListingTablePartitionKeyName),
                KeyType:       types.KeyTypeHash,
            }, {
                AttributeName: aws.String("CategoryCount"),
                KeyType:       types.KeyTypeRange,
            }},
            Projection: &types.Projection{
                ProjectionType: types.ProjectionTypeAll,
            },
        }},

        GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
            {
                IndexName: aws.String(constant.CategoryPriceIndex),
                KeySchema: []types.KeySchemaElement{{
                    AttributeName: aws.String("Category"),
                    KeyType:       types.KeyTypeHash,
                }, {
                    AttributeName: aws.String("Price"),
                    KeyType:       types.KeyTypeRange,
                }},
                Projection: &types.Projection{
                    ProjectionType: types.ProjectionTypeAll,
                },
                ProvisionedThroughput: provisionedThroughput,
            },
            {
                IndexName: aws.String(constant.CategoryCreatedAtIndex),
                KeySchema: []types.KeySchemaElement{{
                    AttributeName: aws.String("Category"),
                    KeyType:       types.KeyTypeHash,
                }, {
                    AttributeName: aws.String("CreatedAt"),
                    KeyType:       types.KeyTypeRange,
                }},
                Projection: &types.Projection{
                    ProjectionType: types.ProjectionTypeAll,
                },
                ProvisionedThroughput: provisionedThroughput,
            },
            {
                IndexName: aws.String(constant.ListingIdIndex),
                KeySchema: []types.KeySchemaElement{{
                    AttributeName: aws.String(constant.ListingIdIndexPartitionKeyName),
                    KeyType:       types.KeyTypeHash,
                }, {
                    AttributeName: aws.String(constant.ListingTablePartitionKeyName),
                    KeyType:       types.KeyTypeRange,
                }},
                Projection: &types.Projection{
                    ProjectionType: types.ProjectionTypeAll,
                },
                ProvisionedThroughput: provisionedThroughput,
            },
        },

        TableName:             aws.String(constant.TableName),
        ProvisionedThroughput: provisionedThroughput,
    })
    if err != nil {
        d.log.Fatalf("Got error calling CreateTable: %s", err)
    }

    waiter := dynamodb.NewTableExistsWaiter(d.client)
    err = waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
        TableName: aws.String(constant.TableName)}, 5*time.Minute)
    if err != nil {
        d.log.Fatalf("Got error waiting for table to exist: %s", err)
    }

    return table.TableDescription, nil
}

// DeleteTable deletes the DynamoDB Listing table and all its data
func (d DynamoDataAccess) DeleteTable() error {
    _, err := d.client.DeleteTable(context.TODO(), &dynamodb.DeleteTableInput{
        TableName: aws.String(constant.TableName)})
    if err != nil {
        d.log.Errorf("Got error calling DeleteTable: %s", err)
    }
    return err
}
