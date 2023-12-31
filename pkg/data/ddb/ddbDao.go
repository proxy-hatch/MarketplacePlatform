package ddb

import (
    "context"
    "errors"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
    "github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
    "go.uber.org/zap"
    "marketplace-platform/pkg/constant"
    "marketplace-platform/pkg/data/model"
    "marketplace-platform/pkg/data/model/enum"
    "marketplace-platform/pkg/exception"
    "marketplace-platform/pkg/util"
    "os"
    "strconv"
)

type DynamoDataAccess struct {
    client *dynamodb.Client
    log    *zap.SugaredLogger
}

func NewDynamoDataAccess(log *zap.SugaredLogger) DynamoDataAccess {
    // try to get the endpoint from the environment variable
    var endpoint string
    var endpointExists bool
    endpoint, endpointExists = os.LookupEnv(constant.DynamoDbEndpointEnvKey)
    if !endpointExists {
        endpoint = "http://localhost:8000"
    }
    log.Debug("DynamoDB endpoint: " + endpoint)

    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion("eu-west-1"),
        config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
            func(service, region string, options ...interface{}) (aws.Endpoint, error) {
                return aws.Endpoint{URL: endpoint,
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

// PutUser a new user
// Returns nil if the user already exists
func (d DynamoDataAccess) PutUser(username string) (*model.User, error) {
    user := model.User{
        Username: username,
    }

    // Marshal the User struct to a DynamoDB attribute value map
    av, err := user.DdbMarshalMap()
    if err != nil {
        return nil, fmt.Errorf("failed to marshal User struct to attribute value map: %w", err)
    }

    expr, err := expression.NewBuilder().WithCondition(
        expression.Name(constant.ListingTablePartitionKeyName).AttributeNotExists().And(
            expression.Name(constant.ListingTableSortKeyName).AttributeNotExists()),
    ).Build()
    if err != nil {
        return nil, err
    }

    // Create a PutItemInput struct
    input := &dynamodb.PutItemInput{
        Item:                      av,
        ExpressionAttributeNames:  expr.Names(),
        ExpressionAttributeValues: expr.Values(),
        ConditionExpression:       expr.Condition(),
        TableName:                 aws.String(constant.TableName),
    }
    _, err = d.client.PutItem(context.TODO(), input)
    if err != nil {
        var conditionCheckFailedErr *types.ConditionalCheckFailedException
        if errors.As(err, &conditionCheckFailedErr) {
            return nil, nil
        }
        return nil, err
    }

    return &user, nil
}

// GetUser retrieves a user by username
// Returns nil if the user does not exist
func (d DynamoDataAccess) GetUser(username string) (*model.User, error) {
    av, err := model.User{
        Username: username,
    }.DdbMarshalMap()
    if err != nil {
        return nil, err
    }

    input := &dynamodb.GetItemInput{
        Key:       av,
        TableName: aws.String(constant.TableName),
    }
    output, err := d.client.GetItem(context.TODO(), input)
    if err != nil {
        d.log.Errorw("failed to get user", "username", username, "error", err)
        return nil, err
    }

    if output.Item == nil {
        return nil, nil
    }

    var user model.User
    err = attributevalue.UnmarshalMap(output.Item, &user)
    if err != nil {
        d.log.Errorf("failed to unmarshal user: %v", err)
        return nil, err
    }

    return &user, nil
}

func (d DynamoDataAccess) getNextListingId() (int, error) {
    expr, err := expression.NewBuilder().WithKeyCondition(expression.Key(constant.ListingIdIndexPartitionKeyName).Equal(expression.Value(constant.ListingIdIndexPartitionKey))).Build()
    if err != nil {
        return -1, err
    }

    input := &dynamodb.QueryInput{
        KeyConditionExpression:    expr.KeyCondition(),
        ProjectionExpression:      expr.Projection(),
        ExpressionAttributeNames:  expr.Names(),
        ExpressionAttributeValues: expr.Values(),
        ScanIndexForward:          aws.Bool(false),
        Limit:                     aws.Int32(1),
        TableName:                 aws.String(constant.TableName),
        IndexName:                 aws.String(constant.ListingIdIndex),
    }
    output, err := d.client.Query(context.TODO(), input)
    if err != nil {
        return -1, err
    }

    // unmarshall the listing id if exists, otherwise return 100001
    if output.Count == 0 {
        return 100001, nil
    }

    var listing model.Listing
    err = attributevalue.UnmarshalMap(output.Items[0], &listing)
    if err != nil {
        return 0, err
    }

    return listing.ListingId + 1, nil
}

// PutListing puts a listing item to the database
func (d DynamoDataAccess) PutListing(
    username string,
    title string,
    description string,
    price int,
    category string,
) (*model.Listing, error) {
    listingId, err := d.getNextListingId()
    if err != nil {
        d.log.Error("failed to get next listing id: ", err)
        return nil, err
    }

    listing, err := model.NewListing(listingId, username, title, description, price, category)
    if err != nil {
        d.log.Error("failed to create new listing: ", err)
        return nil, err
    }

    // Marshal the Listing struct to a DynamoDB attribute value map
    av, err := listing.DdbMarshalMap()
    if err != nil {
        d.log.Errorf("failed to marshal Listing struct %s to attribute value map: %v", util.AnyToJsonString(listing), err)
        return nil, err
    }

    putListingExpr, err := expression.NewBuilder().WithCondition(
        expression.Name(constant.ListingTablePartitionKeyName).AttributeNotExists()).Build()
    if err != nil {
        return nil, err
    }
    incrementCategoryCountExpr, err := expression.NewBuilder().WithUpdate(expression.Add(expression.Name("CategoryCount"), expression.Value(1))).Build()
    if err != nil {
        return nil, err
    }
    input := &dynamodb.TransactWriteItemsInput{
        TransactItems: []types.TransactWriteItem{
            {
                Put: &types.Put{
                    Item:                      av,
                    TableName:                 aws.String(constant.TableName),
                    ExpressionAttributeNames:  putListingExpr.Names(),
                    ExpressionAttributeValues: putListingExpr.Values(),
                    ConditionExpression:       putListingExpr.Condition(),
                },
            },
            {
                Update: &types.Update{
                    Key:                       buildCategoryMetricKey(category),
                    ExpressionAttributeNames:  incrementCategoryCountExpr.Names(),
                    ExpressionAttributeValues: incrementCategoryCountExpr.Values(),
                    UpdateExpression:          incrementCategoryCountExpr.Update(),
                    TableName:                 aws.String(constant.TableName),
                },
            },
        },
    }
    _, err = d.client.TransactWriteItems(context.TODO(), input)
    if err != nil {
        var txCanceledErr *types.TransactionCanceledException
        if errors.As(err, &txCanceledErr) {
            for idx, reason := range txCanceledErr.CancellationReasons {
                if *reason.Code != "None" {
                    d.log.Errorf("Transaction cancelled at index %d with reason: %v", idx, reason)
                }
                if idx == 0 && *reason.Code == "ConditionalCheckFailed" {
                    return nil, nil
                }
            }
        }
        d.log.Errorf("TransactWriteItems failed with unhandled error: %v", err)
        return nil, err
    }

    return &listing, nil
}

// GetListing retrieves a listing by listingId
func (d DynamoDataAccess) GetListing(listingId int) (*model.Listing, error) {
    expr, err := expression.NewBuilder().WithKeyCondition(expression.Key(constant.ListingTablePartitionKeyName).Equal(expression.Value(listingId))).Build()
    if err != nil {
        return nil, err
    }

    input := &dynamodb.QueryInput{
        KeyConditionExpression:    expr.KeyCondition(),
        ProjectionExpression:      expr.Projection(),
        ExpressionAttributeNames:  expr.Names(),
        ExpressionAttributeValues: expr.Values(),
        TableName:                 aws.String(constant.TableName),
    }

    output, err := d.client.Query(context.TODO(), input)
    if err != nil {
        d.log.Errorf("failed to query listing with listingId %s: %v", listingId, err)
        return nil, err
    }

    if len(output.Items) == 0 {
        return nil, nil
    }
    if len(output.Items) > 1 {
        d.log.Errorf("more than one listing with listingId '%s'. Returning the 1st one.", listingId)
    }

    var listing model.Listing
    err = attributevalue.UnmarshalMap(output.Items[0], &listing)
    if err != nil {
        d.log.Errorf("failed to unmarshal listing with listingId %s: %v", listingId, err)
        return nil, err
    }

    return &listing, nil
}

// GetCategory retrieves all listings of a specified category and sorts them by price or creation time
func (d DynamoDataAccess) GetCategory(category string, sortBy enum.SortBy, order enum.OrderBy) ([]model.Listing, error) {
    indexName := constant.CategoryPriceIndex
    if sortBy == enum.SortByCreatedAt {
        indexName = constant.CategoryCreatedAtIndex
    }

    expr, err := expression.NewBuilder().
        WithKeyCondition(expression.Key("Category").Equal(expression.Value(category))).
        Build()
    if err != nil {
        return nil, err
    }

    input := &dynamodb.QueryInput{
        KeyConditionExpression:    expr.KeyCondition(),
        ExpressionAttributeNames:  expr.Names(),
        ExpressionAttributeValues: expr.Values(),
        TableName:                 aws.String(constant.TableName),
        IndexName:                 aws.String(indexName),
        ScanIndexForward:          aws.Bool(order == enum.OrderByAscending),
    }
    output, err := d.client.Query(context.TODO(), input)
    if err != nil {
        d.log.Errorf("failed to query category %s: %v", category, err)
        return nil, err
    }

    var listings []model.Listing
    err = attributevalue.UnmarshalListOfMaps(output.Items, &listings)
    if err != nil {
        d.log.Errorf("failed to unmarshal listings: %v", err)
        return nil, err
    }

    return listings, nil
}

// GetTopCategory retrieves the category with the highest total number of listings
func (d DynamoDataAccess) GetTopCategory() (string, error) {
    expr, err := expression.NewBuilder().
        WithKeyCondition(expression.Key(constant.ListingTablePartitionKeyName).Equal(expression.Value(constant.CategoryMetricRecordPartitionKey))).
        Build()
    if err != nil {
        return "", err
    }

    // Create a QueryInput struct
    input := &dynamodb.QueryInput{
        KeyConditionExpression:    expr.KeyCondition(),
        ExpressionAttributeNames:  expr.Names(),
        ExpressionAttributeValues: expr.Values(),
        TableName:                 aws.String(constant.TableName),
        IndexName:                 aws.String(constant.CategoryCountIndex),
        ScanIndexForward:          aws.Bool(false), // descending order
        Limit:                     aws.Int32(1),
    }
    output, err := d.client.Query(context.TODO(), input)
    if err != nil {
        d.log.Errorf("failed to query top category: %v", err)
        return "", err
    }

    d.log.Debugf("Query output items: %s", util.AnyToJsonString(output.Items))
    var categoryMetric model.CategoryMetric
    err = attributevalue.UnmarshalMap(output.Items[0], &categoryMetric)
    if err != nil {
        d.log.Errorf("failed to unmarshal category metric: %v", err)
        return "", err
    }

    return categoryMetric.Category, nil
}

// DeleteListing deletes a listing and updates the CategoryMetric
func (d DynamoDataAccess) DeleteListing(username string, listingId int) error {
    // Get the listing to be deleted
    listing, err := d.GetListing(listingId)
    if err != nil {
        d.log.Errorf("failed to get listing with listingId %d: %v", listingId, err)
        return err
    }

    // Check if the listing exists
    if listing == nil {
        return exception.NewListingDoesNotExistException(fmt.Sprintf("listing with listingId %d does not exist", listingId), nil)
    }

    // Check if the username matches the owner of the listing
    if listing.Username != username {
        return exception.NewOwnershipMismatchException(fmt.Sprintf("listing with listingId %d is not owned by %s", listingId, username), nil)
    }

    // Prepare the TransactWriteItems input
    deleteListingExpr, err := expression.NewBuilder().WithCondition(
        expression.Name(constant.ListingTablePartitionKeyName).Equal(expression.Value(listingId)).And(
            expression.Name(constant.ListingTableSortKeyName).Equal(expression.Value(username))),
    ).Build()
    if err != nil {
        return err
    }
    decrementCategoryCountExpr, err := expression.NewBuilder().WithUpdate(expression.Add(expression.Name("CategoryCount"), expression.Value(-1))).Build()
    if err != nil {
        return err
    }
    input := &dynamodb.TransactWriteItemsInput{
        TransactItems: []types.TransactWriteItem{
            {
                Delete: &types.Delete{
                    Key: map[string]types.AttributeValue{
                        constant.ListingTablePartitionKeyName: &types.AttributeValueMemberN{Value: strconv.Itoa(listingId)},
                        constant.ListingTableSortKeyName:      &types.AttributeValueMemberS{Value: username},
                    },
                    TableName:                 aws.String(constant.TableName),
                    ExpressionAttributeNames:  deleteListingExpr.Names(),
                    ExpressionAttributeValues: deleteListingExpr.Values(),
                    ConditionExpression:       deleteListingExpr.Condition(),
                },
            },
            {
                Update: &types.Update{
                    Key:                       buildCategoryMetricKey(listing.Category),
                    ExpressionAttributeNames:  decrementCategoryCountExpr.Names(),
                    ExpressionAttributeValues: decrementCategoryCountExpr.Values(),
                    UpdateExpression:          decrementCategoryCountExpr.Update(),
                    TableName:                 aws.String(constant.TableName),
                },
            },
        },
    }

    // Execute the transaction
    _, err = d.client.TransactWriteItems(context.TODO(), input)
    if err != nil {
        var txCanceledErr *types.TransactionCanceledException
        if errors.As(err, &txCanceledErr) {
            for idx, reason := range txCanceledErr.CancellationReasons {
                if *reason.Code != "None" {
                    d.log.Errorf("Transaction cancelled at index %d with reason: %v", idx, reason)
                }
            }
        }
        d.log.Errorf("TransactWriteItems failed with unhandled error: %v", err)
        return err
    }

    return nil
}

func buildCategoryMetricKey(category string) map[string]types.AttributeValue {
    return map[string]types.AttributeValue{
        constant.ListingTablePartitionKeyName: &types.AttributeValueMemberN{Value: strconv.Itoa(constant.CategoryMetricRecordPartitionKey)},
        constant.ListingTableSortKeyName:      &types.AttributeValueMemberS{Value: category},
    }
}
