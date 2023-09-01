package model

type CategoryMetric struct {
    Category      string `dynamodbav:"Category" validate:"required"`   // sort key
    CategoryCount int    `dynamodbav:"CategoryCount" validate:"gte=0"` // LSI
}

func NewCategoryMetric(category string) CategoryMetric {
    return CategoryMetric{
        Category:      category,
        CategoryCount: 0,
    }
}

func (c CategoryMetric) Validate() error {
    return validate.Struct(c)
}
