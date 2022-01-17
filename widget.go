package gogp2

// #cgo linux pkg-config: libgphoto2
// #include <gphoto2/gphoto2.h>
// #include <string.h>
// #include <stdlib.h>
import "C"
import (
	"encoding/json"

	"fmt"
	"unsafe"

	Log "github.com/qazf88/golog"
)

// GetConfig
func (c *Camera) GetConfig() (string, error) {

	rootWidget, err := c.getRootWidget()
	if err != nil {
		Log.Error(err.Error())
		return "", err
	}

	var arrayWidget []widget
	var widgetSection, child *C.CameraWidget
	defer C.free(unsafe.Pointer(widgetSection))
	defer C.free(unsafe.Pointer(child))

	childCountWindow := int(C.gp_widget_count_children(rootWidget))
	for i := 0; i < childCountWindow; i++ {

		res := C.gp_widget_get_child(rootWidget, C.int(i), (**C.CameraWidget)(unsafe.Pointer(&widgetSection)))
		if res != OK {
			Log.Trace("'gp_widget_get_child()' always return OK.")
			continue
		}

		childCountSection := int(C.gp_widget_count_children(widgetSection))
		for j := 0; j < childCountSection; j++ {

			res = C.gp_widget_get_child(widgetSection, C.int(j), (**C.CameraWidget)(unsafe.Pointer(&child)))
			if res != OK {
				Log.Trace("'gp_widget_get_child()' always return OK.")
				continue
			}

			widget, err := getWidget(child)
			if err != nil {
				Log.Error(err.Error())
				continue
			}

			arrayWidget = append(arrayWidget, widget)
		}
	}

	configJson, err := json.Marshal(arrayWidget)
	if err != nil {
		Log.Error(err.Error())
		return "", err
	}

	return string(configJson), nil
}

// GetWidgetChoicesByName
func (c *Camera) GetWidgetChoicesByName(wName string) ([]string, error) {

	childWidget, err := c.getGpWidgetByName(wName)
	if err != nil {
		return nil, err
	}

	choices, _res := getWidgetChoices(childWidget)
	if _res != OK {
		Log.Warning(fmt.Sprintf("error the list of options is not complete, from widget by name %s, error code: %d ", wName, _res))
		Log.Error(fmt.Sprintf("error the list of options is not complete, from widget by name %s, error code: %d ", wName, _res))
	}

	return choices, nil
}

// GetWidgetByName
func (c *Camera) GetWidgetByName(wName string) (string, error) {

	childWidget, err := c.getGpWidgetByName(wName)
	if err != nil {
		Log.Error(err.Error())
		return "", err
	}

	_widget, err := getWidget(childWidget)
	if err != nil {
		return "", err
	}

	result, err := json.Marshal(_widget)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// getWidgetByName
func (c *Camera) getWidgetByName(wName string) (widget, error) {

	childWidget, err := c.getGpWidgetByName(wName)
	if err != nil {
		Log.Error(err.Error())
		return widget{}, err
	}

	_widget, err := getWidget(childWidget)
	if err != nil {
		return widget{}, err
	}

	return _widget, nil
}

// GetWidgetValueByName
func (c *Camera) GetWidgetValueByName(wName string) (string, error) {

	childWidget, err := c.getGpWidgetByName(wName)
	if err != nil {
		Log.Error(err.Error())
		return "", err
	}

	wType, err := getWidgetType(childWidget)
	if err != nil {
		Log.Error(err.Error())
		return "", err
	}

	value, _res := getWidgetValue(childWidget, wType)
	if _res != OK {
		err := fmt.Sprintf("error get 'value' from widget by name %s, error code: %d", wName, _res)
		Log.Error(err)
		return "", fmt.Errorf(err)
	}
	return value, nil
}

// SetWigetValueByName
func (c *Camera) SetWigetValueByName(wName string, wValue string) error {

	_widget, err := c.getWidgetByName(wName)
	if err != nil {
		return err
	}

	if _widget.ReadOnly {
		return fmt.Errorf("error widget '%s' read-only", _widget.Name)
	}

	if _widget.Value == wValue {
		Log.Info("value " + wValue + " is already relevant")
		return nil
	}

	for _, choice := range _widget.Choice {
		if choice == wValue {
			err := c.setValue(&wName, &wValue)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("value '%s' cannot be set", wValue)
}

// SetWiget
func (c *Camera) SetWiget(jsonWidget string) error {

	newWidget := []widget{}
	err := json.Unmarshal([]byte("["+jsonWidget+"]"), &newWidget)
	if err != nil {
		Log.Error(err.Error())
		return err
	}

	_widget, err := c.getWidgetByName(newWidget[0].Name)
	if err != nil {
		return err
	}

	if _widget.ReadOnly {
		return fmt.Errorf("error widget '%s' read-only", _widget.Name)
	}

	if _widget.Value == newWidget[0].Value {
		Log.Info("value " + newWidget[0].Value + " is already relevant")
		return nil
	}

	for _, choice := range _widget.Choice {
		if choice == newWidget[0].Value {
			err := c.setValue(&newWidget[0].Name, &newWidget[0].Value)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("name '%s' cannot be set widget", newWidget[0].Name)
}

// getRootWidget
func (c *Camera) getRootWidget() (*C.CameraWidget, error) {

	var rootWidget *C.CameraWidget
	defer C.free(unsafe.Pointer(rootWidget))
	res := C.gp_camera_get_config(c.Camera, (**C.CameraWidget)(unsafe.Pointer(&rootWidget)), c.Context)
	if res != OK {
		return nil, fmt.Errorf("error initialize camera config, error code: %d", res)
	}

	return rootWidget, nil
}

// getStringWidgetName
func getStringWidgetName(_widget *C.CameraWidget) string {

	var C_name *C.char

	res := C.gp_widget_get_name(_widget, (**C.char)(unsafe.Pointer(&C_name)))
	if res != OK {
		Log.Error(fmt.Sprintf("error get widget name, error code: %d", res))
		return ""
	}

	return C.GoString(C_name)
}

// getWidget
func getWidget(_widget *C.CameraWidget) (widget, error) {

	var C_info *C.char
	var C_label *C.char
	var C_name *C.char
	var C_readonly C.int

	res := C.gp_widget_get_name(_widget, (**C.char)(unsafe.Pointer(&C_name)))
	if res != OK {
		err := fmt.Sprintf("error get widget name, error code: %d", res)
		Log.Error(err)
		empty := widget{}
		return empty, fmt.Errorf(err)
	}

	wType, err := getWidgetType(_widget)
	if err != nil {
		Log.Error(err.Error())
		empty := widget{}
		return empty, err
	}

	res = C.gp_widget_get_readonly(_widget, &C_readonly)
	if res != OK {
		err := fmt.Sprintf("error get 'read-only' value from widget by name %s, error code: %d", C.GoString(C_name), res)
		Log.Error(err)
		empty := widget{}
		return empty, fmt.Errorf(err)
	}

	res = C.gp_widget_get_info(_widget, (**C.char)(unsafe.Pointer(&C_info)))
	if res != OK {
		err := fmt.Sprintf("error get 'info' value from widget by name %s, error code: %d", C.GoString(C_name), res)
		Log.Error(err)
		empty := widget{}
		return empty, fmt.Errorf(err)
	}

	res = C.gp_widget_get_label(_widget, (**C.char)(unsafe.Pointer(&C_label)))
	if res != OK {
		err := fmt.Sprintf("error get 'label' value from widget by name %s, error code: %d", C.GoString(C_name), res)
		Log.Error(err)
		empty := widget{}
		return empty, fmt.Errorf(err)
	}

	value, _res := getWidgetValue(_widget, wType)
	if _res != OK {
		err := fmt.Sprintf("error get 'value' from widget by name %s, error code: %d", C.GoString(C_name), res)
		Log.Error(err)
		empty := widget{}
		return empty, fmt.Errorf(err)
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
		choices, _res = getWidgetChoices(_widget)
		if _res != OK {
			Log.Warning(fmt.Sprintf("error the list of options is not complete, from widget by name %s, error code: %d ", C.GoString(C_name), _res))
			Log.Error(fmt.Sprintf("error the list of options is not complete, from widget by name %s, error code: %d ", C.GoString(C_name), _res))
		}
	}

	_cameraWidget := widget{
		Label:    C.GoString(C_label),
		Info:     C.GoString(C_info),
		Name:     C.GoString(C_name),
		Type:     widgetType(wType),
		Choice:   choices,
		Value:    value,
		ReadOnly: (int(C_readonly) == 1),
	}

	return _cameraWidget, nil
}

// getWidgetValue
func getWidgetValue(_widget *C.CameraWidget, _type C.CameraWidgetType) (string, int) {

	var C_value *C.char

	res := C.gp_widget_get_value(_widget, (unsafe.Pointer(&C_value)))
	if res != OK {

		return "", int(res)
	}

	switch _type {
	case typeWidgetText:
		return C.GoString(C_value), 0
	case typeWidgetRadio:
		return C.GoString(C_value), 0
	case typeWidgetMenu:
		return C.GoString(C_value), 0
	case typeWidgetButton:
		return C.GoString(C_value), 0
	case typeWidgetRange:
		//return C.GoString(C_value), 0
	case typeWidgetDate:
		return fmt.Sprintf("%d", C_value), 0
	case typeWidgetToggle:
		return fmt.Sprintf("%d", C_value), 0
	}

	return "", -999
}

// getWidgetChoices
func getWidgetChoices(_widget *C.CameraWidget) ([]string, int) {

	_res := 0

	choicesList := []string{}
	numChoices := C.gp_widget_count_choices(_widget)

	for i := 0; i < int(numChoices); i++ {
		var choice *C.char
		res := C.gp_widget_get_choice(_widget, C.int(i), (**C.char)(unsafe.Pointer(&choice)))
		if res != OK {
			_res = int(res)
			continue
		}
		choicesList = append(choicesList, C.GoString(choice))
	}

	return choicesList, _res
}

// getWidgetType
func getWidgetType(_widget *C.CameraWidget) (C.CameraWidgetType, error) {

	var _widgetType C.CameraWidgetType

	res := C.gp_widget_get_type(_widget, (*C.CameraWidgetType)(unsafe.Pointer(&_widgetType)))
	if res != OK {
		wName := getStringWidgetName(_widget)
		err := fmt.Sprintf("could not retrieve widget type by name %s, error code: %d", wName, res)
		Log.Error(err)
		return _widgetType, fmt.Errorf(err)
	}
	return _widgetType, nil
}

// getGpWidgetByName
func (c *Camera) getGpWidgetByName(wName string) (*C.CameraWidget, error) {

	rootWidget, err := c.getRootWidget()
	defer C.free(unsafe.Pointer(rootWidget))
	if err != nil {
		return nil, err
	}

	var childWidget *C.CameraWidget
	defer C.free(unsafe.Pointer(childWidget))

	res := C.gp_widget_get_child_by_name(rootWidget, C.CString(wName), (**C.CameraWidget)(unsafe.Pointer(&childWidget)))
	if res != OK {
		err := fmt.Sprintf("could not retrieve widget by name %s, error code: %d", wName, res)
		Log.Error(err)
		return nil, fmt.Errorf(err)
	}
	return childWidget, nil
}

// setValue
func (c *Camera) setValue(wName *string, wValue *string) error {

	_widget, err := c.getGpWidgetByName(*wName)
	if err != nil {
		Log.Error(err.Error())
		return err
	}

	C_value := C.CString(*wValue)
	defer C.free(unsafe.Pointer(C_value))
	res := C.gp_widget_set_value(_widget, unsafe.Pointer(C_value))
	if res != OK {
		return fmt.Errorf("error setting the value for widget by name %s, error code: %d", *wName, res)
	}

	C_name := C.CString(*wName)
	defer C.free(unsafe.Pointer(C_name))
	res = C.gp_camera_set_single_config(c.Camera, C_name, _widget, c.Context)
	if res != OK {
		return fmt.Errorf("error save widget, error code: %d", res)
	}
	return nil
}
