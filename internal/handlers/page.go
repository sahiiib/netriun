package handlers

const assetVersion = "20260910-1"

func NewPageData(title, description, canonical string) PageData {
	return PageData{
		Title:        title,
		Description:  description,
		Canonical:    canonical,
		AssetVersion: assetVersion,
	}
}
