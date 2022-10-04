package main

type webFlagSet struct {
	strings map[string]string
	args    []string
}

func (w *webFlagSet) Bool(_ string, def bool, _ string) *bool {
	return &def
}

func (w *webFlagSet) Int(_ string, def int, _ string) *int {
	return &def
}

func (w *webFlagSet) Float64(_ string, def float64, _ string) *float64 {
	return &def
}

func (w *webFlagSet) String(name string, def, _ string) *string {
	if b, ok := w.strings[name]; ok {
		return &b
	}
	return &def
}

func (w *webFlagSet) StringList(_, _, _ string) *[]*string {
	return &[]*string{}
}

func (w *webFlagSet) ExtraUsage() string {
	return ""
}

func (w *webFlagSet) AddExtraUsage(string) {
}

func (w *webFlagSet) Parse(func()) []string {
	return w.args
}
