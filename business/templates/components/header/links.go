package header

type Nav struct {
	Title string
	Href  string
}

var navs = []Nav{
	{
		Title: "ABOUT",
		Href:  "/",
	},
	{
		Title: "BLOG",
		Href:  "/blog",
	},
	{
		Title: "CV",
		Href:  "/cv",
	},
}
