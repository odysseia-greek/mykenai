package seeder

type ExampleModel struct {
    Examples []Examples `json:"examples"`
}

type Examples struct {
	Author          string  `json:"author"`
	Book            string  `json:"book"`
	Type            string  `json:"type"`
	Reference       string  `json:"reference"`
	PerseusTextLink string  `json:"perseusTextLink"`
}
