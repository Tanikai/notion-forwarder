package models

type NotionDbId string

type ForwardUrls struct {
	Urls []string
}

type ForwardedDatabase struct {
	Name              string     // name of database
	DatabaseId        NotionDbId // id of database
	ForwardColumnName string     // name of column to forward

	ForwardingDict map[string]ForwardUrls
}
