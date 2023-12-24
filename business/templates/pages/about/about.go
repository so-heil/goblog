package about

import "github.com/a-h/templ"

type About struct {
	Title    string
	SubTitle string
	Content  []templ.Component
}
