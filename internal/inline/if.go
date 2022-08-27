/*
Depcrecated: Package inline provides the trivial functionality of inline if-else expressions.
*/
package inline

// If returns:
//
//	then      if condition == true
//	otherwise if condition != true
func If[T any](condition bool, then T, otherwise T) T {
	if condition {
		return then
	}
	return otherwise
}
