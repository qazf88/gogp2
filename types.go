package gogp2

// #cgo linux pkg-config: libgphoto2
// #include <gphoto2/gphoto2.h>
// #include <string.h>
import "C"

type GoContext C.context
type GoCamera C.Camera
type WidgetType string

type Camera struct {
	Camera       *C.Camera
	Context      *C.GPContext
	Info         []string
	CameraStatus bool
	Config       []string
}

type Widget struct {
	Label    string     `json:"label"`
	Name     string     `json:"name"`
	Info     string     `json:"info"`
	Value    string     `json:"value"`
	Choice   []string   `json:"choise"`
	ReadOnly bool       `json:"readOnly"`
	Type     WidgetType `json:"type"`
}

type CameraFilePath struct {
	Name     string
	Folder   string
	Isdir    bool
	Children []CameraFilePath

	camera *Camera
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
