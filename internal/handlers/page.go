package handlers

const assetVersion = "20260519-2"

func NewPageData(title, description, canonical string) PageData {
	return PageData{
		Title:        title,
		Description:  description,
		Canonical:    canonical,
		AssetVersion: assetVersion,
	}
}
