package model

type CategoryMetric struct {
    Category      string `dynamodbav:"Username" validate:"required"`   // sort key
    CategoryCount int    `dynamodbav:"CategoryCount" validate:"gte=0"` // LSI
}

func (c CategoryMetric) Validate() error {
    return validate.Struct(c)
}
