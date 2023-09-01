package constant

var (
    TableName              = "Listing"
    CategoryPriceIndex     = "CategoryPriceIndex"
    CategoryCreatedAtIndex = "CategoryCreatedAtIndex"
    CategoryCountIndex     = "CategoryCountIndex"
    ListingIdIndex         = "ListingIdIndex"

    ListingTablePartitionKeyName = "ListingId"
    ListingTableSortKeyName      = "Username"
    UserRootRecordPartitionKey   = -1

    CategoryMetricRecordPartitionKey = -2

    ListingIdIndexPartitionKeyName = "ListingIdIndexAttribute"
    ListingIdIndexPartitionKey     = 1
)
