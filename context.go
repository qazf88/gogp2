package gogp2

// #cgo linux pkg-config: libgphoto2
// #include <gphoto2/gphoto2.h>
// #include <string.h>
import "C"
import (
	"fmt"

	Log "github.com/qazf88/golog"
)

//NewContext Initialize new context
func (c *Camera) NewContext() error {
	if c.Context != nil {
		err := "context is already initialized"
		Log.Error(err)
		return fmt.Errorf(err)
	}
	Context := C.gp_context_new()
	if Context == nil {
		err := "error initialize context"
		Log.Error(err)
		return fmt.Errorf(err)
	}
	c.Context = Context
	Log.Trace("initialize context")
	return nil
}

//FreeContext Free context
func (c *Camera) FreeContext() error {
	if c.Context != nil {
		C.gp_context_unref(c.Context)
		c = nil
		Log.Trace("free context")
		return nil
	}
	err := "can not free context is empty"
	Log.Error(err)
	return fmt.Errorf(err)
}
