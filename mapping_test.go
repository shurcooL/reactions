package reactions_test

import (
	"fmt"
	"reflect"
	"strings"
)

/*func TestMapping(t *testing.T) {
	tests := []string{
		"32.5% 55%",
		"32.5% 42.5%",
	}
	for _, test := range tests {
		foo(test)
	}
}*/

func ExampleBar() {
	in := "32.5% 55%"
	want := []string{"32.5%", "55%"}
	got := bar(in)

	fmt.Println(reflect.DeepEqual(got, want))

	// Output: true
}

func bar(in string) []string {
	return strings.Fields(in)
}

func ExampleBaz() {
	in := "32.5%"
	want := 13
	got := baz(in)

	fmt.Println(reflect.DeepEqual(got, want))

	// Output: true
}
