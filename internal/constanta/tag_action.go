package constanta

type TagServiceAction int8

const (
	// trigger when article version is published or archived
	CalculateTagUsageAndPairFrequency TagServiceAction = iota
	// trigger when article version is drafted or published
	CalculateArticleTagRelation
)
