package godgets

import (
	"fmt"
	"reflect"
)

func assertEqual(found, expected interface{}) {
	if !reflect.DeepEqual(found, expected) {
		panic(fmt.Sprintf("found %#v, expected %#v", found, expected))
	}
}
