// Code generated by "stringer -linecomment -type StreamID"; DO NOT EDIT.

package pdb

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[StreamIDPrevStreamTable-0]
	_ = x[StreamIDPDBStream-1]
	_ = x[StreamIDTPIStream-2]
}

const _StreamID_name = "previous stream tablePDB streamTPI stream"

var _StreamID_index = [...]uint8{0, 21, 31, 41}

func (i StreamID) String() string {
	if i >= StreamID(len(_StreamID_index)-1) {
		return "StreamID(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _StreamID_name[_StreamID_index[i]:_StreamID_index[i+1]]
}
