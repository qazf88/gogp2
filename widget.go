package gogp2

// #cgo linux pkg-config: libgphoto2
// #include <gphoto2/gphoto2.h>
// #include <string.h>
// #include <stdio.h>
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

	if retval := C.gp_camera_get_config(c.Camera, (**C.CameraWidget)(unsafe.Pointer(&rootWidget)), c.Context); retval != OK {
		return nil, fmt.Errorf("error initialize camera config :%v", retval)
	}
	return rootWidget, nil
}

func (c *Camera) GetConfig() (*[]string, error) {
	Log.Trace("get config camera")
	var rootWidget *C.CameraWidget

	rootWidget, err := c.getRootWidget()
	if err != nil {
		Log.Error(err.Error())
	}
	//defer C.free(unsafe.Pointer(rootWidget))

	// if res := C.gp_camera_get_config(c.Camera, (**C.CameraWidget)(unsafe.Pointer(&rootWidget)), c.Context); res != OK {
	// 	err := "error get camera config: " + strconv.Itoa(int(res))
	// 	Log.Error(err)
	// 	return nil, fmt.Errorf(err)
	// }

	var arrayWidget []string
	childrenWindow := int(C.gp_widget_count_children(rootWidget))

	var childWindow *C.CameraWidget
	var childSection *C.CameraWidget

	for n := 0; n < childrenWindow; n++ {
		if res := C.gp_widget_get_child(rootWidget, C.int(n), (**C.CameraWidget)(unsafe.Pointer(&childWindow))); res != OK {
			err := "error get child widget: " + strconv.Itoa(int(res))
			Log.Error(err)
			return nil, fmt.Errorf(err)
		}
		childrenSection := int(C.gp_widget_count_children(childWindow))
		for n := 0; n < childrenSection; n++ {
			C.gp_widget_get_child(childWindow, C.int(n), (**C.CameraWidget)(unsafe.Pointer(&childSection)))
			___widget := c.getWidget(childSection, true)
			fullWidjet, err := json.Marshal(___widget)
			if err != nil {
				Log.Error(err.Error())
				return nil, err
			}
			arrayWidget = append(arrayWidget, string(fullWidjet)+",")
		}
	}
	return &arrayWidget, nil
}

func (c *Camera) getWidget(widget *C.CameraWidget, child bool) Widget {

	var _info *C.char
	var _label *C.char
	var _name *C.char
	var _widgetType C.CameraWidgetType
	var _readonly C.int

	C.gp_widget_get_info(widget, (**C.char)(unsafe.Pointer(&_info)))
	C.gp_widget_get_label(widget, (**C.char)(unsafe.Pointer(&_label)))
	C.gp_widget_get_name(widget, (**C.char)(unsafe.Pointer(&_name)))
	C.gp_widget_get_type(widget, (*C.CameraWidgetType)(unsafe.Pointer(&_widgetType)))
	C.gp_widget_get_readonly(widget, &_readonly)

	var choices []string
	var err error

	if _widgetType == typeWidgetToggle {
		choices = []string{"0", "1"}
	} else {
		choices, err = c.getChoice(widget)
		if err != nil {
			Log.Error(err.Error())
			choices = nil
		}
	}
	value, err := c.getValue(widget, _widgetType)
	if err != nil {
		Log.Error(err.Error())
		value = ""
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
	return _cameraWidget
}

func (c *Camera) getChoice(widget *C.CameraWidget) ([]string, error) {
	choicesList := []string{}
	numChoices := C.gp_widget_count_choices(widget)

	for i := 0; i < int(numChoices); i++ {
		var choice *C.char
		res := C.gp_widget_get_choice(widget, C.int(i), (**C.char)(unsafe.Pointer(&choice)))
		if res != OK {
			Log.Info(string(res))
		}
		choicesList = append(choicesList, C.GoString(choice))
	}
	return choicesList, nil
}

func (c *Camera) getValue(widget *C.CameraWidget, _type C.CameraWidgetType) (string, error) {
	var value *C.char

	if retval := C.gp_widget_get_value(widget, (unsafe.Pointer(&value))); retval != OK {
		return "", fmt.Errorf("cannot read widget property value, error code :%d", retval)
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
		return C.GoString(value), nil
	case typeWidgetDate:
		return fmt.Sprintf("%d", value), nil
	case typeWidgetToggle:
		return fmt.Sprintf("%d", value), nil
	}

	return "", fmt.Errorf("err")
}
