package dependencies

import (
	"context"
	"fmt"
	"log/slog"
	"notion-forwarder/models"

	"github.com/jomei/notionapi"
)

var ErrDatabaseNotFound = fmt.Errorf("database not found")
var ErrItemNotFound = fmt.Errorf("item not found")

var ErrForwardingPropNotFound = fmt.Errorf("forwarding property not found")

var ErrForwardingPropNotRichText = fmt.Errorf("forwarding property is not a RichTextProperty")

type NotionForwardingClient struct {
	client             *notionapi.Client
	forwardedDatabases map[string]models.ForwardedDatabase
}

func initializeForwardedDatabase(config models.ForwardedDatabaseConfig) models.ForwardedDatabase {
	return models.ForwardedDatabase{
		Name:              config.Name,
		DatabaseId:        models.NotionDbId(config.DatabaseId),
		ForwardColumnName: config.ForwardColumnName,
		ForwardingDict:    make(map[string]models.ForwardUrls),
	}
}

func richTextToString(richText []notionapi.RichText) string {
	text := ""
	for _, rt := range richText {
		text += rt.PlainText
	}
	return text

}

func addPageUrlToForwardingDict(page notionapi.Page, database *models.ForwardedDatabase) error {
	// Get forwarding from page
	forwarding := page.Properties[database.ForwardColumnName]
	if forwarding == nil {
		return ErrForwardingPropNotFound
	}
	// check that forwarding is a RichTextProperty
	if forwarding.GetType() != notionapi.PropertyTypeRichText {
		return ErrForwardingPropNotRichText
	}

	// convert forwarding property to string
	forwardingRichText := forwarding.(*notionapi.RichTextProperty)
	// Convert RichText array to string
	forwardingStr := richTextToString(forwardingRichText.RichText)
	forwardUrls, ok := database.ForwardingDict[forwardingStr]
	if !ok {
		forwardUrls = models.ForwardUrls{Urls: []string{}}
	}
	forwardUrls.Urls = append(forwardUrls.Urls, page.URL)
	database.ForwardingDict[forwardingStr] = forwardUrls
	return nil
}

func refreshForwardedDatabase(client *notionapi.Client, database *models.ForwardedDatabase) error {
	// Clear database and populate with current entries
	database.ForwardingDict = make(map[string]models.ForwardUrls)

	hasMore := true
	startCursor := notionapi.Cursor("")

	for ok := true; ok; ok = hasMore {
		requestBody := notionapi.DatabaseQueryRequest{
			PageSize:    5,
			StartCursor: startCursor,
		}

		queryResp, err := client.Database.Query(
			context.Background(),
			notionapi.DatabaseID(database.DatabaseId),
			&requestBody)
		if err != nil {
			// notion api returned error
			return err
		}

		for _, page := range queryResp.Results {
			err := addPageUrlToForwardingDict(page, database)
			if err != nil {
				return fmt.Errorf("error adding page to forwarding dict: %v", err)
			}
		}

		hasMore = queryResp.HasMore
		startCursor = queryResp.NextCursor
	}

	return nil
}

func NewNotionForwardingClient(integrationToken string, databases []models.ForwardedDatabaseConfig) *NotionForwardingClient {
	client := notionapi.NewClient(notionapi.Token(integrationToken))

	forwardedDatabases := make(map[string]models.ForwardedDatabase)
	for _, config := range databases {
		forwardedDatabases[config.Name] = initializeForwardedDatabase(config)
	}

	return &NotionForwardingClient{
		client:             client,
		forwardedDatabases: forwardedDatabases,
	}
}

func (n NotionForwardingClient) PopulateForwardedDatabases() error {
	if len(n.forwardedDatabases) == 0 {
		slog.Warn("no databases to populate")
	}
	for _, database := range n.forwardedDatabases {
		err := refreshForwardedDatabase(n.client, &database)
		if err != nil {
			return err
		}
		slog.Info("populated database", "database", database.Name, "entries", len(database.ForwardingDict))
	}
	return nil
}

func (n NotionForwardingClient) GetForwarding(databaseName string, itemId string) ([]string, error) {
	result, err := n.GetForwardingCached(databaseName, itemId)
	if err == nil {
		slog.Debug("Forwarding found in cache", "database", databaseName, "itemId", itemId)
		return result, nil
	}

	result, err = n.GetForwardingFromNotion(databaseName, itemId)
	if err == nil {
		slog.Debug("Forwarding found in Notion", "database", databaseName, "itemId", itemId)
		return result, nil
	}
	slog.Debug("Forwarding not found", "database", databaseName, "itemId", itemId)
	return nil, err
}

func (n NotionForwardingClient) GetForwardingCached(databaseName string, itemId string) ([]string, error) {
	forwDb, ok := n.forwardedDatabases[databaseName]
	if !ok {
		return nil, ErrDatabaseNotFound
	}

	entry, ok := forwDb.ForwardingDict[itemId]
	if !ok {
		return nil, ErrItemNotFound
	}

	return entry.Urls, nil
}

func (n NotionForwardingClient) GetForwardingFromNotion(databaseName string, itemId string) ([]string, error) {
	forwDb, ok := n.forwardedDatabases[databaseName]
	if !ok {
		return nil, ErrDatabaseNotFound
	}

	// Get forwardings from Notion Database
	requestBody := notionapi.DatabaseQueryRequest{
		PageSize: 100,
		Filter:   notionapi.PropertyFilter{Property: forwDb.ForwardColumnName, RichText: &notionapi.TextFilterCondition{Equals: itemId}},
	}
	queryResp, err := n.client.Database.Query(
		context.Background(),
		notionapi.DatabaseID(forwDb.DatabaseId),
		&requestBody)
	if err != nil {
		return nil, err
	}

	if len(queryResp.Results) == 0 {
		return nil, ErrItemNotFound
	}

	for _, page := range queryResp.Results {
		err := addPageUrlToForwardingDict(page, &forwDb)
		if err != nil {
			return nil, fmt.Errorf("error adding page to forwarding dict: %v", err)
		}
	}

	return forwDb.ForwardingDict[itemId].Urls, nil
}
