// Code generated by "stringer -linecomment -type TPIVersion"; DO NOT EDIT.

package pdb

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TPIVersionV40-19950410]
	_ = x[TPIVersionV41-19951122]
	_ = x[TPIVersionV50Interim-19960307]
	_ = x[TPIVersionV50-19961031]
	_ = x[TPIVersionV70-19990903]
	_ = x[TPIVersionV80-20040203]
}

const (
	_TPIVersion_name_0 = "V 4.0 (1995-04-10)"
	_TPIVersion_name_1 = "V 4.1 (1995-11-22)"
	_TPIVersion_name_2 = "V 5.0 - interim (1996-03-07)"
	_TPIVersion_name_3 = "V 5.0 (1996-10-31)"
	_TPIVersion_name_4 = "V 7.0 (1999-09-03)"
	_TPIVersion_name_5 = "V 8.0 (2004-02-03)"
)

func (i TPIVersion) String() string {
	switch {
	case i == 19950410:
		return _TPIVersion_name_0
	case i == 19951122:
		return _TPIVersion_name_1
	case i == 19960307:
		return _TPIVersion_name_2
	case i == 19961031:
		return _TPIVersion_name_3
	case i == 19990903:
		return _TPIVersion_name_4
	case i == 20040203:
		return _TPIVersion_name_5
	default:
		return "TPIVersion(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
