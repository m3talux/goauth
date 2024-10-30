package mongo

import (
	"net/url"

	"go.mongodb.org/mongo-driver/mongo"
)

type Document interface {
	Indexes() []mongo.IndexModel
	NameSingular() string
	NamePlural() string
	CollectionName() string
}

type UpdateResult struct {
	NotFound    bool
	UniqueError bool
	Inserted    bool
}

func uriForLog(uri string) string {
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return ""
	}

	hiddenPassword := "xxxx"

	if parsedURL.User != nil {
		parsedURL.User = url.UserPassword(parsedURL.User.Username(), hiddenPassword)
	}

	return parsedURL.String()
}
