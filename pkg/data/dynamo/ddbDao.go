package dynamo

import (
    "context"
    "errors"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
    "go.uber.org/zap"
    "time"
)

type DynamoDataAccess struct {
    client *dynamodb.Client
    log    *zap.SugaredLogger
}

var (
    tableName              = "Listing"
    CategoryPriceIndex     = "CategoryPriceIndex"
    CategoryCreatedAtIndex = "CategoryCreatedAtIndex"
    CategoryCountIndex     = "CategoryCountIndex"
)

func NewDynamoDataAccess(log *zap.SugaredLogger) DynamoDataAccess {
    // Create a custom configuration
    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion("eu-west-1"),
        config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
            func(service, region string, options ...interface{}) (aws.Endpoint, error) {
                return aws.Endpoint{URL: "http://localhost:8000",
                    SigningRegion: "eu-west-1",
                }, nil
            })),
    )
    if err != nil {
        log.Fatalf("unable to load SDK config, %v", err)
    }

    // Create a DynamoDB client
    client := dynamodb.NewFromConfig(cfg)

    return DynamoDataAccess{
        client: client,
        log:    log,
    }
}

// TableExists determines whether the Listing table exists
func (d DynamoDataAccess) ListingTableExists() (bool, error) {
    _, err := d.client.DescribeTable(
        context.TODO(), &dynamodb.DescribeTableInput{TableName: aws.String(tableName)},
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
            AttributeName: aws.String("Username"),
            AttributeType: types.ScalarAttributeTypeS,
        }, {
            AttributeName: aws.String("ListingId"),
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
        }},

        KeySchema: []types.KeySchemaElement{{
            AttributeName: aws.String("Username"),
            KeyType:       types.KeyTypeHash,
        }, {
            AttributeName: aws.String("ListingId"),
            KeyType:       types.KeyTypeRange,
        }},

        LocalSecondaryIndexes: []types.LocalSecondaryIndex{{
            IndexName: aws.String(CategoryCountIndex),
            KeySchema: []types.KeySchemaElement{{
                AttributeName: aws.String("Username"),
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
                IndexName: aws.String(CategoryPriceIndex),
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
                IndexName: aws.String(CategoryCreatedAtIndex),
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
        },

        TableName:             aws.String(tableName),
        ProvisionedThroughput: provisionedThroughput,
    })
    if err != nil {
        d.log.Fatalf("Got error calling CreateTable: %s", err)
    }

    waiter := dynamodb.NewTableExistsWaiter(d.client)
    err = waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
        TableName: aws.String(tableName)}, 5*time.Minute)
    if err != nil {
        d.log.Fatalf("Got error waiting for table to exist: %s", err)
    }

    return table.TableDescription, nil
}
