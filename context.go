package gogp2

// #cgo linux pkg-config: libgphoto2
// #include <gphoto2/gphoto2.h>
// #include <string.h>
import "C"
import (
	"fmt"

	Log "github.com/qazf88/golog"
)

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

	return nil
}

func (c *Camera) FreeContext() error {

	if c.Context != nil {
		C.gp_context_unref(c.Context)
		c = nil
		Log.Trace("free context")
		return nil
	}
	c.Context = nil
	err := "can not free context is empty"
	Log.Error(err)
	return fmt.Errorf(err)
}
