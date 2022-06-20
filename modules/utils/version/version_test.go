package version

import (
	"fmt"
	"testing"
)

func TestGreater(t *testing.T) {
	v1 := "2.11.3"
	v2 := "11.2.3"
	greater, err := GreaterOrEqual(v1, v2)
	if err != nil {
		fmt.Println("err is :", err)
	}
	fmt.Println(greater)
}
