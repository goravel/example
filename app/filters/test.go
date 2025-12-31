package filters

import "context"

type Test struct {
}

// Signature The signature of the filter.
func (receiver *Test) Signature() string {
	return "test"
}

// Handle defines the filter function to apply.
//
// The Handle method should return a function that processes an input and
// returns a transformed value. The function can either return the
// transformed value alone or a tuple of the transformed value and an error.
// The input to the filter function is flexible: the first input is the value
// of the key on which the filter is applied, and the rest of the inputs are
// the arguments passed to the filter.
//
// Example usages:
//
//  1. Return only the transformed value:
//     func (val string) int {
//     // conversion logic
//     return 1
//     }
//
//  2. Return the transformed value and an error:
//     func (val int) (int, error) {
//     // conversion logic with error handling
//     return 1, nil
//     }
//
//  3. Take additional arguments:
//     func (val string, def ...string) string {
//     if val == "" && len(def) > 0 {
//     return def[0]
//     }
//     return val
//     }
func (receiver *Test) Handle(ctx context.Context) any {
	return nil
}
