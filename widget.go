package gogp2

// #cgo linux pkg-config: libgphoto2
// #include <gphoto2/gphoto2.h>
// #include <string.h>
import "C"
import (
	"fmt"
	"strconv"
	"unsafe"

	Log "github.com/qazf88/GoLog"
)

func (c *Camera) GetConfig() (error, string) {
	Log.Trace("get config camera")
	var rootWidget *C.CameraWidget
	var child *C.CameraWidget
	var _child *C.CameraWidget

	if res := C.gp_camera_get_config(c.Camera, (**C.CameraWidget)(unsafe.Pointer(&rootWidget)), c.Context); res != OK {
		err := "error get camera config: " + strconv.Itoa(int(res))
		Log.Error(err)
		return fmt.Errorf(err), ""
	}
	fmt.Println(c.getWidget(rootWidget, false))

	childrenCount := int(C.gp_widget_count_children(rootWidget))

	for n := 0; n < childrenCount; n++ {
		if res := C.gp_widget_get_child(rootWidget, C.int(n), (**C.CameraWidget)(unsafe.Pointer(&child))); res != OK {
			err := "error get child widget: " + strconv.Itoa(int(res))
			Log.Error(err)
			return fmt.Errorf(err), ""
		}
		widgetSection := c.getWidget(child, false)
		fmt.Println(widgetSection)
		childrenCount2 := int(C.gp_widget_count_children(child))
		for n := 0; n < childrenCount2; n++ {
			C.gp_widget_get_child(child, C.int(n), (**C.CameraWidget)(unsafe.Pointer(&_child)))
			___widget := c.getWidget(_child, true)
			fmt.Println(___widget)
		}

	}
	return nil, "ok"
}

func (c *Camera) getWidget(widget *C.CameraWidget, child bool) Widget {
	//_cameraWidget := cameraWidget{}
	var gpInfo *C.char
	var gpLabel *C.char
	var gpName *C.char
	var gpWidgetType C.CameraWidgetType
	//var child *C.CameraWidget
	var readonly C.int

	C.gp_widget_get_info(widget, (**C.char)(unsafe.Pointer(&gpInfo)))
	C.gp_widget_get_label(widget, (**C.char)(unsafe.Pointer(&gpLabel)))
	C.gp_widget_get_name(widget, (**C.char)(unsafe.Pointer(&gpName)))
	C.gp_widget_get_type(widget, (*C.CameraWidgetType)(unsafe.Pointer(&gpWidgetType)))
	C.gp_widget_get_readonly(widget, &readonly)
	c.getChoice(widget)
	c.getValue(widget)
	_cameraWidget := Widget{
		widgetType: widgetType(gpWidgetType),
		Label:      C.GoString(gpLabel),
		Info:       C.GoString(gpInfo),
		Name:       C.GoString(gpName),
		ReadOnly:   (int(readonly) == 1),
	}
	return _cameraWidget
}

func (c *Camera) getChoice(widget *C.CameraWidget) (error, []string) {
	choicesList := []string{}
	numChoices := C.gp_widget_count_choices(widget)
	for i := 0; i < int(numChoices); i++ {
		var choice *C.char
		C.gp_widget_get_choice(widget, C.int(i), (**C.char)(unsafe.Pointer(&choice)))
		choicesList = append(choicesList, C.GoString(choice))
	}
	return nil, choicesList
}

func (c *Camera) getValue(widget *C.CameraWidget) (string, error) {
	var value *C.char
	if retval := C.gp_widget_get_value(widget, (unsafe.Pointer(&value))); retval != OK {
		return "", fmt.Errorf("Cannot read widget property value, error code :%d", retval)
	}
	fmt.Println(C.GoString(value))
	if value != nil {
		//widgetValue := C.GoString(value)
		//return widgetValue, nil
	}

	return "", nil
}
