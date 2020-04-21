// Code generated by "stringer -linecomment -type PDBVersion"; DO NOT EDIT.

package pdb

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[PDBVersionVC2-19941610]
	_ = x[PDBVersionVC4-19950623]
	_ = x[PDBVersionVC41-19950814]
	_ = x[PDBVersionVC50-19960307]
	_ = x[PDBVersionVC98-19970604]
	_ = x[PDBVersionVC70-20000404]
	_ = x[PDBVersionVC70Dep-19990604]
	_ = x[PDBVersionVC80-20030901]
	_ = x[PDBVersionVC110-20091201]
	_ = x[PDBVersionVC140-20140508]
}

const (
	_PDBVersion_name_0 = "VC 2 (1994-10-16)"
	_PDBVersion_name_1 = "VC 4 (1995-06-23)"
	_PDBVersion_name_2 = "VC 4.1 (1995-08-14)"
	_PDBVersion_name_3 = "VC 5.0 (1996-03-07)"
	_PDBVersion_name_4 = "VC 98 (1997-06-04)"
	_PDBVersion_name_5 = "VC 7.0 - deprecated (1999-06-04)"
	_PDBVersion_name_6 = "VC 7.0 (2000-04-04)"
	_PDBVersion_name_7 = "VC 8.0 (2003-09-01)"
	_PDBVersion_name_8 = "VC 11.0 (2009-12-01)"
	_PDBVersion_name_9 = "VC 14.0 (2014-05-08)"
)

func (i PDBVersion) String() string {
	switch {
	case i == 19941610:
		return _PDBVersion_name_0
	case i == 19950623:
		return _PDBVersion_name_1
	case i == 19950814:
		return _PDBVersion_name_2
	case i == 19960307:
		return _PDBVersion_name_3
	case i == 19970604:
		return _PDBVersion_name_4
	case i == 19990604:
		return _PDBVersion_name_5
	case i == 20000404:
		return _PDBVersion_name_6
	case i == 20030901:
		return _PDBVersion_name_7
	case i == 20091201:
		return _PDBVersion_name_8
	case i == 20140508:
		return _PDBVersion_name_9
	default:
		return "PDBVersion(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
