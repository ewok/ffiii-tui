/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

type LayoutConfig struct {
	TopSize             int
	LeftSize            int
	FullTransactionView bool

	SummarySize int
	TabBarSize  int

	Width  int
	Height int
}

// NewDefaultLayout returns a LayoutConfig with default settings.
// These defaults can be modified using the provided methods.
func NewDefaultLayout() *LayoutConfig {
	return &LayoutConfig{
		Width:       80,
		Height:      24,
		LeftSize:    30,
		TopSize:     3,
		SummarySize: 4,
	}
}

func (lc *LayoutConfig) WithSize(width, height int) *LayoutConfig {
	if lc == nil {
		lc = NewDefaultLayout()
	}
	lc.Width = width
	lc.Height = height
	return lc
}

func (lc *LayoutConfig) WithFullTransactionView(yesNo bool) *LayoutConfig {
	if lc == nil {
		lc = NewDefaultLayout()
	}
	lc.FullTransactionView = yesNo
	return lc
}

func (lc *LayoutConfig) WithLeftSize(size int) *LayoutConfig {
	if lc == nil {
		lc = NewDefaultLayout()
	}
	lc.LeftSize = size
	return lc
}

func (lc *LayoutConfig) WithTopSize(size int) *LayoutConfig {
	if lc == nil {
		lc = NewDefaultLayout()
	}
	lc.TopSize = size
	return lc
}

func (lc *LayoutConfig) WithSummarySize(size int) *LayoutConfig {
	if lc == nil {
		lc = NewDefaultLayout()
	}
	lc.SummarySize = size
	return lc
}

func (lc *LayoutConfig) GetFullTransactionView() bool {
	if lc == nil {
		lc = NewDefaultLayout()
	}
	return lc.FullTransactionView
}

func (lc *LayoutConfig) GetWidth() int {
	if lc == nil {
		lc = NewDefaultLayout()
	}
	return lc.Width
}

func (lc *LayoutConfig) GetHeight() int {
	if lc == nil {
		lc = NewDefaultLayout()
	}
	return lc.Height
}

func (lc *LayoutConfig) GetLeftSize() int {
	if lc == nil {
		lc = NewDefaultLayout()
	}
	return lc.LeftSize
}

func (lc *LayoutConfig) GetTopSize() int {
	if lc == nil {
		lc = NewDefaultLayout()
	}
	return lc.TopSize
}

func (lc *LayoutConfig) GetSummarySize() int {
	if lc == nil {
		lc = NewDefaultLayout()
	}
	return lc.SummarySize
}

func (lc *LayoutConfig) GetTabBarSize() int {
	if lc == nil {
		lc = NewDefaultLayout()
	}
	return lc.TabBarSize
}

func (lc *LayoutConfig) ToggleFullTransactionView() bool {
	if lc == nil {
		lc = NewDefaultLayout()
	}
	lc.FullTransactionView = !lc.FullTransactionView
	return lc.FullTransactionView
}
