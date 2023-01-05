/*
SmartIDE - Dev Containers
Copyright (C) 2023 leansoftX.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package model

import "errors"

type FeedbackError struct {
	IsRetry bool  `json:"isRetry"`
	Err     error `json:"err"`
}

func (e *FeedbackError) Error() string {
	/* msg := ""
	if e.IsRetry {
		msg = "disable retry"
	} */
	return e.Err.Error()
}

func CreateFeedbackError(err string, isRetry bool) FeedbackError {
	return FeedbackError{
		IsRetry: isRetry,
		Err:     errors.New(err),
	}
}

func CreateFeedbackError2(err string, isRetry bool) error {
	tmpErr := FeedbackError{
		IsRetry: isRetry,
		Err:     errors.New(err),
	}
	return &tmpErr
}
