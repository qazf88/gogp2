package gogp2

// #cgo linux pkg-config: libgphoto2
// #include <gphoto2/gphoto2.h>
// #include <string.h>
import "C"

type GoContext *C.GPContext
type CameraWidget struct {
	widget *C.CameraWidget
}
type WidgetType string
type Abilities *C.CameraAbilitiesList

type Camera struct {
	Camera  *C.Camera
	Context *C.GPContext
}

type widget struct {
	Label    string     `json:"label"`
	Name     string     `json:"name"`
	Info     string     `json:"info"`
	Value    string     `json:"value"`
	Choice   []string   `json:"choise"`
	ReadOnly bool       `json:"readOnly"`
	Type     WidgetType `json:"type"`
}

type Lists struct {
	CameraList      *C.CameraList
	AbilitiesList   *C.CameraAbilitiesList
	PortInfoList    *C.GPPortInfoList
	CameraListCount int
}

type CamerasList struct {
	Name   string `json:"name"`
	Port   string `json:"port"`
	Number int    `json:"number"`
}

type CameraFilePath struct {
	Name     string
	Folder   string
	Isdir    bool
	Children []CameraFilePath
}

const (
	OK = 0
)

//widget types
const (
	typeWidgetWindow = iota //(0)
	typeWidgetSection
	typeWidgetText
	typeWidgetRange
	typeWidgetToggle
	typeWidgetRadio
	typeWidgetMenu
	typeWidgetButton
	typeWidgetDate
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

//File types
const (
	//FileTypePreview is a preview of an image
	FileTypePreview = iota
	//FileTypeNormal is regular normal data of a file
	FileTypeNormal
	//FileTypeRaw usually the same as FileTypeNormal for modern cameras ( left for compatibility purposes)
	FileTypeRaw
	//FileTypeAudio is a audio view of a file. Perhaps an embedded comment or similar
	FileTypeAudio
	//FileTypeExif is the  embedded EXIF data of an image
	FileTypeExif
	//FileTypeMetadata is the metadata of a file, like Metadata of files on MTP devices
	FileTypeMetadata
)

func StringGpError(num int) string {
	stringError := C.gp_port_result_as_string(C.int(num))
	return C.GoString(stringError)
}

func widgetType(_WidgetType C.CameraWidgetType) WidgetType {
	switch int(_WidgetType) {
	case typeWidgetButton:
		return WidgetButton
	case typeWidgetDate:
		return WidgetDate
	case typeWidgetMenu:
		return WidgetMenu
	case typeWidgetRadio:
		return WidgetRadio
	case typeWidgetRange:
		return WidgetRange
	case typeWidgetSection:
		return WidgetSection
	case typeWidgetText:
		return WidgetText
	case typeWidgetToggle:
		return WidgetToggle
	case typeWidgetWindow:
		return WidgetWindow
	}
	panic("should not be here")
}
