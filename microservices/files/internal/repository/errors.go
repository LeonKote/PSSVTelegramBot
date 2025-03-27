package repository

import "errors"

var NoAffectedError error = errors.New("Rows not affected")
