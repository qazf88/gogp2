package gogp2

// #cgo linux pkg-config: libgphoto2
// #include <gphoto2/gphoto2.h>
// #include <string.h>
import "C"

type GoContext C.GPContext
type GoCamera C.Camera
type WidgetType string

type Camera struct {
	Camera       *C.Camera
	Context      *C.GPContext
	Info         []string
	CameraStatus bool
	//Config  Widget
	//Config2 CameraWidget
}

type configWindow struct {
	section []configSection
}

type configSection struct {
	widget []Widget
}

type Widget struct {
	Label      string   `json:"label"`
	Name       string   `json:"name"`
	Info       string   `json:"info"`
	Value      string   `json:"value"`
	Choise     []string `json:"choise"`
	ReadOnly   bool     `json:"readOnly"`
	widgetType WidgetType
}

// type baseWidget struct {
// 	label      string
// 	info       string
// 	name       string
// 	value      string
// 	choise     []string
// 	readOnly   bool
// 	widgetType WidgetType
// 	//children   []baseWidget
// }

const (
	OK = 0
)

//widget types
const (
	gpWidgetWindow = iota //(0)
	gpWidgetSection
	gpWidgetText
	gpWidgetRange
	gpWidgetToggle
	gpWidgetRadio
	gpWidgetMenu
	gpWidgetButton
	gpWidgetDate
)

//widget types
const (
	//WidgetWindow is the toplevel configuration widget. It should likely contain multiple #WidgetSection entries.
	WidgetWindow WidgetType = "window"
	//WidgetSection : Section widget (think Tab)
	WidgetSection WidgetType = "section"
	//WidgetText : Text widget (string)
	WidgetText WidgetType = "text"
	//WidgetRange : Slider widget (float)
	WidgetRange WidgetType = "range"
	//WidgetToggle : Toggle widget (think check box) (int)
	WidgetToggle WidgetType = "toggle"
	//WidgetRadio : Radio button widget (string)
	WidgetRadio WidgetType = "radio"
	//WidgetMenu : Menu widget (same as RADIO) (string)
	WidgetMenu WidgetType = "menu"
	//WidgetButton : Button press widget ( CameraWidgetCallback )
	WidgetButton WidgetType = "button"
	//WidgetDate : Date entering widget (int)
	WidgetDate WidgetType = "date"
)

func widgetType(gpWidgetType C.CameraWidgetType) WidgetType {
	switch int(gpWidgetType) {
	case gpWidgetButton:
		return WidgetButton
	case gpWidgetDate:
		return WidgetDate
	case gpWidgetMenu:
		return WidgetMenu
	case gpWidgetRadio:
		return WidgetRadio
	case gpWidgetRange:
		return WidgetRange
	case gpWidgetSection:
		return WidgetSection
	case gpWidgetText:
		return WidgetText
	case gpWidgetToggle:
		return WidgetToggle
	case gpWidgetWindow:
		return WidgetWindow
	}
	panic("should not be here")
}
