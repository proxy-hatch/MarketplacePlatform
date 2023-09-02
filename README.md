# MarketplacePlatform

A basic marketplace platform which allows users to ‘buy’ and ‘sell’ items. Carousell offline assignment.

## Prerequisites
- Docker
- Go 1.19 or above (provided by Docker)

## Running the application

```
./build.sh && ./run.sh
```

## Running without Docker

Must ensure that DynamoDB local is running on port 8000.

```
go run main.go
```

## Running integration test

```
cd cmd
go build -o main && go test
```
# Application Design

## Requirements
- Username and list ID are unique and case insensitive.
- listing ID is sequential starting from 100001.
- Only registered users should be allowed to buy or sell items.
- Each listing can be associated with only 1 user and 1 category.
- Category reads can be sorted on Price or creation time.

## Engineering Design

Application is written in Go, backed
by [DynamoDB local](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.UsageNotes.html) (
DDB-Local).

For simplicity, there is no long-running server. The CLI app itself is long-running and will be terminated by the user
with ___TODO: ___ SIGTERM signal.

DDB-Local strikes a good balance between the light-weight of a NoSQL database and robustness of a production database.
It also has the benefit of able to be deployed on AWS with minimal changes in the code if scalability is required.

DDB-Local also has the flexibility to either run in-memory only or persist to disk. For this application, it is
configured to be in-memory for ease of testing.

### Auth Design

Simply "Registered => authorized". Authentication is performed on each operation besides Register.

### API Design

- Register(username string)
- CreateListing(username string, title string, description string, price int, category string)
- DeleteListing(username string, listingId string)
- GetListing(listingId string)
- GetListingsByCategory(category string, sortBy enum.SortBy, sortOrder enum.SortOrder)
    - SortBy: Price, CreationTime
- GetTopCategory()
    - Get the top category across with the most listings, across all users.

### Data Schema

Price is stored as int to simplify (since it is fixed at 2 decimal). It is assumed that the price is in cents.
CreatedAt is stored as Epoch Seconds to simplify.

#### Listing table

This table consists of three types of records:

1. User root record

   partition key: `#USER_ROOT`

   sort key: Username

2. Listing record

   partition key: ListingId

   sort key: Username

   attributes:
    - Title
    - Description
    - Price
    - Category
    - CreatedAt
3. Category Metric Record

   partition key: `#CATEGORY_METRIC`

   sort key: Category

   attributes:
    - CategoryCount

LSIs:

partition key: ListingId

sort key: CategoryCount

(To be used for GetTopCategory. Only partition key used will be `#CATEGORY_METRIC`)

GSIs:

1. partition key: Category 
   
   sort key: Price

2. partition key: Category

   sort key: CreatedAt

3. partition key: `1`

   sort key: listingId

   (To be used for generating listingId)

##### Use Cases

- Register (put user root record with special sort key '#ROOT')
- Check if user exist (get user root record)
- Assert username is unique (provided by partition key)
- add listing (put listing record with sort key as listing ID)
- Assert listing ID is unique (provided by sort key)
- Assert listing belong to a single user (provided by partition key)

### Scaling consideration

##### Data scaling
This is a barebone application. Additional use-case can be extended by:

1. Additional indexing: easily added via LSI or GSI in DDB.
2. Additional attributes: easily added to DDB without downtime.

NoSQL is not designed for OLAP use-cases, but can be extended to supported advanced queries
with [Amazon EMR integration](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/EMRforDynamoDB.Querying.html).

For heavy OLAP use-cases, consider integrating with Amazon Redshift or migrate to a relational database.

##### Network and compute scaling
Long-running server can be added to handle the load, in which ECS+Fargate fronted by Application Load Balancer can be
used.

Serverless solution like API-Gateway + Lambda is cost-effective, but with its own set of limitations.
It is assumed Managed-DynamoDB will be used, which will scale automatically.

### Security consideration

Attack vectors to consider:

- Man-in-the-middle attack
- DDOS attack

Proper OAUTH should be used for production. Consider implementing a proper auth server like Auth0.

API-Gateway has built-in DDOS protection, and can be configured to block IP addresses that are attacking.

If long-running compute solution is chosen, Application Load Balancer can be used to protect against DDOS.