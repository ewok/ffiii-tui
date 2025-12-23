/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import "time"

func daysIn(m int, year int) int {
	month := time.Month(m)
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
