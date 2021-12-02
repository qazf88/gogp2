package gogp2

// #cgo linux pkg-config: libgphoto2
// #include <gphoto2/gphoto2.h>
// #include <string.h>
// #include <stdlib.h>
import "C"
import (
	"encoding/json"
	"fmt"
	"strconv"
	"unsafe"

	Log "github.com/qazf88/golog"
)

func (c *Camera) getRootWidget() (*C.CameraWidget, error) {
	var rootWidget *C.CameraWidget
	defer C.free(unsafe.Pointer(rootWidget))
	if retval := C.gp_camera_get_config(c.Camera, (**C.CameraWidget)(unsafe.Pointer(&rootWidget)), c.Context); retval != OK {
		return nil, fmt.Errorf("error initialize camera config :%v", retval)
	}
	return rootWidget, nil
}

func (c *Camera) getChildWidget() (*C.CameraWidget, error) {
	var rootWidget *C.CameraWidget

	if retval := C.gp_camera_get_config(c.Camera, (**C.CameraWidget)(unsafe.Pointer(&rootWidget)), c.Context); retval != OK {
		return nil, fmt.Errorf("error initialize camera config :%v", retval)
	}
	return rootWidget, nil
}

func (camera *Camera) freeChildWidget(input *C.CameraWidget) {
	var rootWidget *C.CameraWidget
	C.gp_widget_get_root(input, (**C.CameraWidget)(unsafe.Pointer(&rootWidget)))
	C.free(unsafe.Pointer(rootWidget))
}

func (c *Camera) GetConfig() (*[]string, error) {
	Log.Trace("get config camera")
	var rootWidget *C.CameraWidget

	rootWidget, err := c.getRootWidget()
	if err != nil {
		Log.Error(err.Error())
	}

	var arrayWidget []string
	childrencountWindow := int(C.gp_widget_count_children(rootWidget))

	var widgetSection, child *C.CameraWidget

	for n := 0; n < childrencountWindow; n++ {
		if res := C.gp_widget_get_child(rootWidget, C.int(n), (**C.CameraWidget)(unsafe.Pointer(&widgetSection))); res != OK {
			err := "error get child widget: " + strconv.Itoa(int(res))
			Log.Error(err)
			return nil, fmt.Errorf(err)
		}
		childrenCountSection := int(C.gp_widget_count_children(widgetSection))
		for n := 0; n < childrenCountSection; n++ {
			C.gp_widget_get_child(widgetSection, C.int(n), (**C.CameraWidget)(unsafe.Pointer(&child)))
			_widget, err := c.getWidget(child)
			if err != nil {
				Log.Error(err.Error())
				continue
			}
			fullWidjet, err := json.Marshal(_widget)
			if err != nil {
				Log.Error(err.Error())
				continue
			}
			arrayWidget = append(arrayWidget, string(fullWidjet)+",")
		}
	}
	return &arrayWidget, nil
}

func (c *Camera) getWidget(widget *C.CameraWidget) (Widget, error) {

	var _info *C.char
	var _label *C.char
	var _name *C.char
	var _widgetType C.CameraWidgetType
	var _readonly C.int

	res := C.gp_widget_get_name(widget, (**C.char)(unsafe.Pointer(&_name)))
	if res != OK {
		err := "error get widget name: " + strconv.Itoa(int(res))
		Log.Error(err)
		empty := Widget{}
		return empty, fmt.Errorf(err)
	}

	res = C.gp_widget_get_type(widget, (*C.CameraWidgetType)(unsafe.Pointer(&_widgetType)))
	if res != OK {
		err := "widget " + C.GoString(_name) + " type not defined: " + strconv.Itoa(int(res))
		Log.Error(err)
		empty := Widget{}
		return empty, fmt.Errorf(err)
	}

	res = C.gp_widget_get_readonly(widget, &_readonly)
	if res != OK {
		err := "widget " + C.GoString(_name) + " error read 'read only': " + strconv.Itoa(int(res))
		Log.Error(err)
		empty := Widget{}
		return empty, fmt.Errorf(err)
	}

	C.gp_widget_get_info(widget, (**C.char)(unsafe.Pointer(&_info)))
	C.gp_widget_get_label(widget, (**C.char)(unsafe.Pointer(&_label)))

	value, err := c.getValue(widget, _widgetType)
	if err != nil {
		_err := "widget " + C.GoString(_name) + " error read 'value': " + err.Error()
		Log.Error(_err)
		empty := Widget{}
		return empty, fmt.Errorf(_err)
	}

	var choices []string

	if _widgetType == typeWidgetToggle {
		if value == "2" {
			choices = []string{"not supported"}
			value = ""
		} else {
			choices = []string{"0", "1"}
		}
	} else {
		choices, err = c.getChoice(widget)
		if err != nil {
			Log.Error(err.Error())
			empty := Widget{}
			return empty, err
		}
	}

	_cameraWidget := Widget{
		Label:    C.GoString(_label),
		Info:     C.GoString(_info),
		Name:     C.GoString(_name),
		Type:     widgetType(_widgetType),
		Choice:   choices,
		Value:    value,
		ReadOnly: (int(_readonly) == 1),
	}
	return _cameraWidget, nil
}

func (c *Camera) getChoice2(widget *C.CameraWidget) ([]string, error) {
	choicesList := []string{}
	numChoices := C.gp_widget_count_choices(widget)

	for i := 0; i < int(numChoices); i++ {
		var choice *C.int
		res := C.gp_widget_get_choice(widget, C.int(i), (**C.char)(unsafe.Pointer(&choice)))
		if res != OK {
			Log.Info(string(res))
		}
		choicesList = append(choicesList, fmt.Sprintf("%d", choice))
	}
	return choicesList, nil
}

func (c *Camera) getChoice(widget *C.CameraWidget) ([]string, error) {
	choicesList := []string{}
	numChoices := C.gp_widget_count_choices(widget)

	for i := 0; i < int(numChoices); i++ {
		var choice *C.char
		res := C.gp_widget_get_choice(widget, C.int(i), (**C.char)(unsafe.Pointer(&choice)))
		if res != OK {
			Log.Info(string(res))
			continue
		}
		choicesList = append(choicesList, C.GoString(choice))
	}
	return choicesList, nil
}

func (c *Camera) getValue(widget *C.CameraWidget, _type C.CameraWidgetType) (string, error) {
	var value *C.char

	res := C.gp_widget_get_value(widget, (unsafe.Pointer(&value)))
	if res != OK {
		return "", fmt.Errorf(strconv.Itoa(int(res)))
	}

	switch _type {
	case typeWidgetText:
		return C.GoString(value), nil
	case typeWidgetRadio:
		return C.GoString(value), nil
	case typeWidgetMenu:
		return C.GoString(value), nil
	case typeWidgetButton:
		return C.GoString(value), nil
	case typeWidgetRange:
		//return C.GoString(value), nil
	case typeWidgetDate:
		return fmt.Sprintf("%d", value), nil
	case typeWidgetToggle:
		return fmt.Sprintf("%d", value), nil
	}

	return "", fmt.Errorf("widget type not faund")
}

// func (camera *Camera) GetWidgetByName(name *string) (*Widget, error) {
// 	var rootWidget, childWidget *C.CameraWidget
// 	var err error
// 	if rootWidget, err = camera.getRootWidget(); err != nil {
// 		return nil, err
// 	}

// 	gpChildWidgetName := C.CString(*name)
// 	defer C.free(unsafe.Pointer(gpChildWidgetName))

// 	if retval := C.gp_widget_get_child_by_name(rootWidget, gpChildWidgetName, (**C.CameraWidget)(unsafe.Pointer(&childWidget))); retval != gpOk {
// 		return nil, fmt.Errorf("Could not retrieve child widget with name %s, error code %d", *name, retval)
// 	}
// 	return childWidget, nil
// }

// func (camera *Camera) GetWidgetChoiceByName(name *string) (*Widget, error) {
// 	var rootWidget, childWidget *C.CameraWidget
// 	var err error
// 	if rootWidget, err = camera.getRootWidget(); err != nil {
// 		return nil, err
// 	}

// 	gpChildWidgetName := C.CString(*name)
// 	defer C.free(unsafe.Pointer(gpChildWidgetName))

// 	if retval := C.gp_widget_get_child_by_name(rootWidget, gpChildWidgetName, (**C.CameraWidget)(unsafe.Pointer(&childWidget))); retval != gpOk {
// 		return nil, fmt.Errorf("Could not retrieve child widget with name %s, error code %d", *name, retval)
// 	}
// 	return childWidget, nil
// }

// func (camera *Camera) GetWidgetValueByName(name *string) (*Widget, error) {
// 	var rootWidget, childWidget *C.CameraWidget
// 	var err error
// 	if rootWidget, err = camera.getRootWidget(); err != nil {
// 		return nil, err
// 	}

// 	gpChildWidgetName := C.CString(*name)
// 	defer C.free(unsafe.Pointer(gpChildWidgetName))

// 	if retval := C.gp_widget_get_child_by_name(rootWidget, gpChildWidgetName, (**C.CameraWidget)(unsafe.Pointer(&childWidget))); retval != gpOk {
// 		return nil, fmt.Errorf("Could not retrieve child widget with name %s, error code %d", *name, retval)
// 	}
// 	return childWidget, nil
// }
