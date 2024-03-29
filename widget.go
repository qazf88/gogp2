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

	err := c.Init()
	if err != nil {
		Log.Error(err.Error())
		return "", err
	}

	var arrayWidget []widget
	var widgetSection, child *C.CameraWidget
	defer C.free(unsafe.Pointer(widgetSection))
	defer C.free(unsafe.Pointer(child))

	childCountWindow := int(C.gp_widget_count_children(c.RootWidget))
	for i := 0; i < childCountWindow; i++ {

		res := C.gp_widget_get_child(c.RootWidget, C.int(i), (**C.CameraWidget)(unsafe.Pointer(&widgetSection)))
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
		err := fmt.Sprintf("error get 'value' from widget by name '%s', error code: %d", wName, _res)
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
		return fmt.Errorf("error widget by name '%s' read-only", _widget.Name)
	}

	oldValue := _widget.Value

	if oldValue == wValue {

		Log.Info("value " + wValue + " is already relevant")
		return nil
	}

	for _, choice := range _widget.Choice {
		if choice == wValue {
			err := c.setValue(&wName, &wValue)
			if err != nil {
				c.setValue(&wName, &oldValue)
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("widget by name '%s' cannot be set value '%s' invalid value", wName, wValue)
}

// SetWiget
func (c *Camera) SetWiget(jsonWidget []byte) error {

	newWidget := widget{}
	err := json.Unmarshal(jsonWidget, &newWidget)
	if err != nil {
		Log.Error(err.Error())
		return err
	}

	_widget, err := c.getWidgetByName(newWidget.Name)
	if err != nil {
		return err
	}

	if _widget.ReadOnly {
		return fmt.Errorf("error widget by name '%s' read-only", _widget.Name)
	}

	if _widget.Value == newWidget.Value {
		Log.Info("value " + newWidget.Value + " is already relevant")
		return nil
	}

	for _, choice := range _widget.Choice {
		if choice == newWidget.Value {
			err := c.setValue(&newWidget.Name, &newWidget.Value)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("widget by name '%s' cannot be set value '%s' invalid value", newWidget.Name, newWidget.Value)
}

// SetWigetArray
//   !!! restoreOld not work if missError set true !!!
//   if value of missError is set to true, set all possible widgets and return all errors as an array
//   if value of missError is set to false, return last error and out
//   if value of restoreOld is set to true and installed widget has an error, it stops working, restores all changed values ​​to old
func (c *Camera) SetWigetArray(widgets []byte, missError bool, restoreOld bool) []error {

	newWidget := []widget{}
	oldWidget := []widget{}
	errors := []error{}

	err := json.Unmarshal(widgets, &newWidget)
	if err != nil {
		Log.Error(err.Error())
		errors = append(errors, err)
		return errors
	}

	widgetLength := len(newWidget)

	for i := 0; i < widgetLength; i++ {
		_widget, err := c.getWidgetByName(newWidget[i].Name)
		if err != nil {
			errors = append(errors, err)
			if missError {
				continue
			} else {
				if restoreOld {
					goto restore
				} else {
					return errors
				}
			}
		}

		if _widget.ReadOnly {
			err = fmt.Errorf("error widget '%s' read-only", _widget.Name)
			errors = append(errors, err)
			if missError {
				continue
			} else {
				if restoreOld {
					goto restore
				} else {
					return errors
				}
			}
		}

		if _widget.Value == newWidget[i].Value {
			Log.Info("value " + newWidget[i].Value + " is already relevant")
			continue
		}

		flagChoice := false
		for _, choice := range _widget.Choice {
			if choice == newWidget[i].Value {
				err := c.setValue(&newWidget[i].Name, &newWidget[i].Value)
				if err != nil {
					errors = append(errors, err)
					if missError {
						break
					} else {
						if restoreOld {
							goto restore
						} else {
							return errors
						}
					}
				}
				flagChoice = true
				oldWidget = append(oldWidget, _widget)
			}
		}

		if !flagChoice {
			err = fmt.Errorf("widget by name '%s' cannot be set value '%s' invalid value", newWidget[i].Name, newWidget[i].Value)
			errors = append(errors, err)
		}

		if missError {
			continue
		} else {
			if restoreOld {
				goto restore
			} else {
				return errors
			}
		}
	}

	if len(errors) < 1 {
		return nil
	}

	return errors

restore:
	for i := 0; i < len(oldWidget); i++ {
		_widget, err := c.getWidgetByName(oldWidget[i].Name)
		fmt.Println(oldWidget[i].Name)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		if _widget.ReadOnly {
			err = fmt.Errorf("error widget '%s' read-only", _widget.Name)
			errors = append(errors, err)
			continue
		}

		if _widget.Value == oldWidget[i].Value {
			Log.Info("value " + oldWidget[i].Value + " is already relevant")
			continue
		}

		for _, choice := range _widget.Choice {
			if choice == oldWidget[i].Value {
				err := c.setValue(&oldWidget[i].Name, &oldWidget[i].Value)
				if err != nil {
					errors = append(errors, err)
					continue
				}
			}
		}

		err = fmt.Errorf("could not retrieveor alredy installed widget by name '%s'", oldWidget[i].Name)
		errors = append(errors, err)
	}
	return errors
}

// getRootWidget
func (c *Camera) getRootWidget() (**C.CameraWidget, error) {

	var rootWidget *C.CameraWidget
	defer C.free(unsafe.Pointer(rootWidget))

	res := C.gp_camera_get_config(c.Camera, (**C.CameraWidget)(unsafe.Pointer(&rootWidget)), c.Context)
	if res != OK {
		return nil, fmt.Errorf("error initialize camera config, error code: %d", res)
	}

	return &rootWidget, nil
}

// getStringWidgetName
func getStringWidgetName(_widget *C.CameraWidget) (string, error) {

	var C_name *C.char

	res := C.gp_widget_get_name(_widget, (**C.char)(unsafe.Pointer(&C_name)))
	if res != OK {
		return "", fmt.Errorf("error get widget name, error code: %d", res)
	}

	return C.GoString(C_name), nil
}

// getWidgetByName
func (c *Camera) getWidgetByName(wName string) (widget, error) {

	childWidget, err := c.getGpWidgetByName(wName)
	if err != nil {
		return widget{}, err
	}

	_widget, err := getWidget(childWidget)
	if err != nil {
		return widget{}, err
	}

	return _widget, nil
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
		return widget{}, fmt.Errorf(err)
	}

	wType, err := getWidgetType(_widget)
	if err != nil {
		return widget{}, err
	}

	res = C.gp_widget_get_readonly(_widget, &C_readonly)
	if res != OK {
		return widget{}, fmt.Errorf("error get 'read-only' value from widget by name '%s', error code: %d", C.GoString(C_name), res)
	}

	res = C.gp_widget_get_info(_widget, (**C.char)(unsafe.Pointer(&C_info)))
	if res != OK {
		return widget{}, fmt.Errorf("error get 'info' value from widget by name '%s', error code: %d", C.GoString(C_name), res)
	}

	res = C.gp_widget_get_label(_widget, (**C.char)(unsafe.Pointer(&C_label)))
	if res != OK {
		return widget{}, fmt.Errorf("error get 'label' value from widget by name '%s', error code: %d", C.GoString(C_name), res)
	}

	value, _res := getWidgetValue(_widget, wType)
	if _res != OK {
		return widget{}, fmt.Errorf("error get 'value' from widget by name '%s', error code: %d", C.GoString(C_name), res)
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
			Log.Warning(fmt.Sprintf("error the list of options is not complete, from widget by name '%s', error code: %d ", C.GoString(C_name), _res))
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
		wName, err := getStringWidgetName(_widget)
		if err != nil {
			return _widgetType, fmt.Errorf("could not retrieve widget type , error code: %d", res)
		}
		return _widgetType, fmt.Errorf("could not retrieve widget type by name %s, error code: %d", wName, res)
	}

	return _widgetType, nil
}

// getGpWidgetByName
func (c *Camera) getGpWidgetByName(wName string) (*C.CameraWidget, error) {

	var childWidget *C.CameraWidget
	defer C.free(unsafe.Pointer(childWidget))

	res := C.gp_widget_get_child_by_name(c.RootWidget, C.CString(wName), (**C.CameraWidget)(unsafe.Pointer(&childWidget)))
	if res != OK {
		return nil, fmt.Errorf("could not retrieve widget by name '%s', error code: %d", wName, res)
	}
	return childWidget, nil
}

// setValue
func (c *Camera) setValue(wName *string, wValue *string) error {

	_widget, err := c.getGpWidgetByName(*wName)
	if err != nil {
		return err
	}

	C_value := C.CString(*wValue)
	defer C.free(unsafe.Pointer(C_value))

	res := C.gp_widget_set_value(_widget, unsafe.Pointer(C_value))
	if res != OK {
		return fmt.Errorf("error setting the value for widget by name '%s', error code: %d", *wName, res)
	}

	C_name := C.CString(*wName)
	defer C.free(unsafe.Pointer(C_name))

	res = C.gp_camera_set_single_config(c.Camera, C_name, _widget, c.Context)
	if res != OK {
		return fmt.Errorf("error save widget by name '%s', error code: %d", *wName, res)
	}
	return nil
}
