package main

// type HypothesisSearch struct {
// 	Limit        int    `json:"limit"`
// 	Offset       int    `json:"offset"`
// 	Search_after string `json:"search_after"`
// 	Order        string `json:"order"`
// 	User         string `json:"user"`
// }

type HypothesisResponseRow struct {
	Uri      string   `json:"uri"`
	Tags     []string `json:"tags"`
	Text     string   `json:"text"` // user note
	Updated  string   `json:"updated"`
	Document struct {
		Title []string
	}
	Target []struct {
		Selector []struct {
			Exact string
		}
	}
}

type HypothesisResponse struct {
	Total string                  `json:"total"`
	Rows  []HypothesisResponseRow `json:"rows"`
}

type HypothesisNotation struct {
	Title string
	Url   string
	Cite  []struct {
		Note  string
		Quote string
	}
	Tags []string
}

type OutlineArticle struct {
	Title        string `json:"title"`
	Text         string `json:"text"`
	CollectionId string `json:"collectionId"`
	Publish      string `json:"publish"`
}

type OutlineResponse struct {
	Ok      string `json:"ok"`
	Error   string `json:"error"`
	Message string `json:"message"`
	Status  string `json:"status"`
}
