package main

type webUI struct {
}

func (w *webUI) ReadLine(string) (string, error) {
	return "", nil
}

func (w *webUI) Print(...interface{}) {
}

func (w *webUI) PrintErr(...interface{}) {
}

func (w *webUI) IsTerminal() bool {
	return false
}

func (w *webUI) WantBrowser() bool {
	return false
}

func (w *webUI) SetAutoComplete(func(string) string) {
}
