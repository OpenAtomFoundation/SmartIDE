package init

type NewTypeBO struct {
	TypeName string    `json:"typename"`
	SubTypes []SubType `json:"subtype"`
	Commands []string  `json:"command"`
}

type TemplateTypeBo struct {
	TypeName string
	SubType  string
	Commands []string
}

type SubType struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
	Logo  string `json:"logo"`
}
