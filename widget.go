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

func (c *Camera) GetConfig() (*[]string, error) {

	rootWidget, err := c.getRootWidget()
	if err != nil {
		Log.Error(err.Error())
		return nil, err
	}

	var arrayWidget []string
	var widgetSection, child *C.CameraWidget
	defer C.free(unsafe.Pointer(widgetSection))
	defer C.free(unsafe.Pointer(child))

	childCountWindow := int(C.gp_widget_count_children(rootWidget))
	for i := 0; i < childCountWindow; i++ {

		res := C.gp_widget_get_child(rootWidget, C.int(i), (**C.CameraWidget)(unsafe.Pointer(&widgetSection)))
		if res != OK {
			// always return OK.
			continue
		}

		childCountSection := int(C.gp_widget_count_children(widgetSection))
		for j := 0; j < childCountSection; j++ {

			res = C.gp_widget_get_child(widgetSection, C.int(j), (**C.CameraWidget)(unsafe.Pointer(&child)))
			if res != OK {
				// always return OK.
				continue
			}

			widget, err := getWidget(child)
			if err != nil {
				Log.Error(err.Error())
				continue
			}

			fullWidjet, err := json.Marshal(widget)
			if err != nil {
				Log.Error(err.Error())
				continue
			}

			if j == (childCountSection-1) && i == (childCountWindow-1) {
				arrayWidget = append(arrayWidget, string(fullWidjet))
			} else {
				arrayWidget = append(arrayWidget, string(fullWidjet)+",")
			}
		}
	}

	return &arrayWidget, nil
}

func (c *Camera) getRootWidget() (*C.CameraWidget, error) {

	var rootWidget *C.CameraWidget
	defer C.free(unsafe.Pointer(rootWidget))
	res := C.gp_camera_get_config(c.Camera, (**C.CameraWidget)(unsafe.Pointer(&rootWidget)), c.Context)
	if res != OK {
		return nil, fmt.Errorf("error initialize camera config :%v", res)
	}

	return rootWidget, nil
}

func getWidget(_widget *C.CameraWidget) (widget, error) {

	var _info *C.char
	var _label *C.char
	var _name *C.char
	var _readonly C.int

	res := C.gp_widget_get_name(_widget, (**C.char)(unsafe.Pointer(&_name)))
	if res != OK {
		err := "error get widget name: " + strconv.Itoa(int(res))
		Log.Error(err)
		empty := widget{}
		return empty, fmt.Errorf(err)
	}

	wType, err := getWidGetType(_widget)
	if err != nil {
		Log.Error(err.Error())
		empty := widget{}
		return empty, err
	}

	res = C.gp_widget_get_readonly(_widget, &_readonly)
	if res != OK {
		err := "widget " + C.GoString(_name) + " error read 'read only': " + strconv.Itoa(int(res))
		Log.Error(err)
		empty := widget{}
		return empty, fmt.Errorf(err)
	}

	C.gp_widget_get_info(_widget, (**C.char)(unsafe.Pointer(&_info)))
	C.gp_widget_get_label(_widget, (**C.char)(unsafe.Pointer(&_label)))

	value, err := getWidgetValue(_widget, wType)
	if err != nil {
		_err := "widget " + C.GoString(_name) + " error read value: " + err.Error()
		Log.Error(_err)
		empty := widget{}
		return empty, fmt.Errorf(_err)
	}

	var choices []string
	if wType == typeWidgetToggle {
		if value == "2" {
			choices = []string{"not supported"}
			value = ""
		} else {
			choices = []string{"0", "1"}
		}
	} else {
		choices, err = getWidgetChoices(_widget)
		if err != nil {
			Log.Error(err.Error())
			empty := widget{}
			return empty, err
		}
	}

	_cameraWidget := widget{
		Label:    C.GoString(_label),
		Info:     C.GoString(_info),
		Name:     C.GoString(_name),
		Type:     widgetType(wType),
		Choice:   choices,
		Value:    value,
		ReadOnly: (int(_readonly) == 1),
	}

	return _cameraWidget, nil
}

func getWidgetChoices(_widget *C.CameraWidget) ([]string, error) {

	choicesList := []string{}
	numChoices := C.gp_widget_count_choices(_widget)

	for i := 0; i < int(numChoices); i++ {
		var choice *C.char
		res := C.gp_widget_get_choice(_widget, C.int(i), (**C.char)(unsafe.Pointer(&choice)))
		if res != OK {
			Log.Info(string(res))
			continue
		}
		choicesList = append(choicesList, C.GoString(choice))
	}

	return choicesList, nil
}

func getWidgetValue(_widget *C.CameraWidget, _type C.CameraWidgetType) (string, error) {

	var value *C.char

	res := C.gp_widget_get_value(_widget, (unsafe.Pointer(&value)))
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

func (c *Camera) GetWidgetChoicesByName(wName string) ([]string, error) {

	childWidget, err := c.getGpWidgetByName(wName)
	if err != nil {
		return nil, err
	}

	choices, err := getWidgetChoices(childWidget)
	if err != nil {
		Log.Error(err.Error())
		return nil, err
	}

	return choices, nil
}

func (c *Camera) GetWidgetByName(wName string) (*widget, error) {

	childWidget, err := c.getGpWidgetByName(wName)
	if err != nil {
		Log.Error(err.Error())
		return nil, err
	}

	_widget, err := getWidget(childWidget)
	if err != nil {
		return nil, err
	}

	return &_widget, nil
}

func (c *Camera) GetWidgetValueByName(wName string) (string, error) {

	childWidget, err := c.getGpWidgetByName(wName)
	if err != nil {
		Log.Error(err.Error())
		return "", err
	}

	wType, err := getWidGetType(childWidget)
	if err != nil {
		Log.Error(err.Error())
		return "", err
	}

	value, err := getWidgetValue(childWidget, wType)
	if err != nil {
		Log.Error(err.Error())
		return "", err
	}
	return value, nil
}

func getWidGetType(_widget *C.CameraWidget) (C.CameraWidgetType, error) {

	var _widgetType C.CameraWidgetType

	res := C.gp_widget_get_type(_widget, (*C.CameraWidgetType)(unsafe.Pointer(&_widgetType)))
	if res != OK {
		err := "could not retrieve widget type, error code" + strconv.Itoa(int(res))
		Log.Error(err)
		return _widgetType, fmt.Errorf(err)
	}
	return _widgetType, nil
}

func (c *Camera) getGpWidgetByName(_name string) (*C.CameraWidget, error) {

	_rootWidget, err := c.getRootWidget()
	//defer C.free(unsafe.Pointer(_rootWidget))
	if err != nil {
		return nil, err
	}

	var childWidget *C.CameraWidget
	defer C.free(unsafe.Pointer(childWidget))

	res := C.gp_widget_get_child_by_name(_rootWidget, C.CString(_name), (**C.CameraWidget)(unsafe.Pointer(&childWidget)))
	if res != OK {
		err := "could not retrieve widget with name " + _name + ", error code" + strconv.Itoa(int(res))
		Log.Error(err)
		return nil, fmt.Errorf(err)
	}
	return childWidget, nil
}

// func (w *CameraWidget) freeChildWidget() error{
// 	var _rootWidget *C.CameraWidget
// 	defer C.free(unsafe.Pointer(_rootWidget))
// 	C.gp_widget_get_root(*w, (**C.CameraWidget)(unsafe.Pointer(&_rootWidget)))
// 	C.free(unsafe.Pointer(_rootWidget))
// }

func (c *Camera) SetValueWigetByName(wName string, wValue string) error {

	choices, err := c.GetWidgetChoicesByName(wName)
	if err != nil {
		Log.Error(err.Error())
		return err
	}
	for _, choice := range choices {
		if choice == wValue {

			_widget, err := c.getGpWidgetByName(wName)
			if err != nil {
				Log.Error(err.Error())
				return err
			}
			_value := C.CString(wValue)
			defer C.free(unsafe.Pointer(_value))
			res := C.gp_widget_set_value(_widget, unsafe.Pointer(_value))
			//res2 := C.gp_camera_set_config(c.Camera, _widget, c.Context)
			//	fmt.Println(res2)

			if res != OK {
				return fmt.Errorf("error set value ")
			} else {
				fmt.Println(res)

			}
			return nil
		}
	}
	return fmt.Errorf("erterer")
}

// 	var value *C.char

// 	res := C.gp_widget_get_value(widget, (unsafe.Pointer(&value)))
// 	if res != OK {
// 		return fmt.Errorf(strconv.Itoa(int(res)))
// 	}

// 	// switch _type {
// 	// case typeWidgetText:
// 	// 	return C.GoString(value), nil
// 	// case typeWidgetRadio:
// 	// 	return C.GoString(value), nil
// 	// case typeWidgetMenu:
// 	// 	return C.GoString(value), nil
// 	// case typeWidgetButton:
// 	// 	return C.GoString(value), nil
// 	// case typeWidgetRange:
// 	// 	//return C.GoString(value), nil
// 	// case typeWidgetDate:
// 	// 	return fmt.Sprintf("%d", value), nil
// 	// case typeWidgetToggle:
// 	// 	return fmt.Sprintf("%d", value), nil
// 	// }

// 	// return "", fmt.Errorf("widget type not faund")
// 	return nil
// }
