/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/

package ui

import (
	"errors"
	"reflect"
	"testing"

	"ffiii-tui/internal/firefly"
	"ffiii-tui/internal/ui/notify"
	"ffiii-tui/internal/ui/prompt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mockAssetAPI struct {
	updateAccountsFunc       func(accountType string) error
	accountsByTypeFunc       func(accountType string) []firefly.Account
	accountBalanceFunc       func(accountID string) float64
	createAssetAccountFunc   func(name, currencyCode string) error
	updateAccountsCalledWith []string
	createAssetCalledWith    []struct {
		name, currency string
	}
}

func (m *mockAssetAPI) UpdateAccounts(accountType string) error {
	m.updateAccountsCalledWith = append(m.updateAccountsCalledWith, accountType)
	if m.updateAccountsFunc != nil {
		return m.updateAccountsFunc(accountType)
	}
	return nil
}

func (m *mockAssetAPI) AccountsByType(accountType string) []firefly.Account {
	if m.accountsByTypeFunc != nil {
		return m.accountsByTypeFunc(accountType)
	}
	return nil
}

func (m *mockAssetAPI) AccountBalance(accountID string) float64 {
	if m.accountBalanceFunc != nil {
		return m.accountBalanceFunc(accountID)
	}
	return 0
}

func (m *mockAssetAPI) CreateAssetAccount(name, currencyCode string) error {
	m.createAssetCalledWith = append(m.createAssetCalledWith, struct {
		name, currency string
	}{name: name, currency: currencyCode})
	if m.createAssetAccountFunc != nil {
		return m.createAssetAccountFunc(name, currencyCode)
	}
	return nil
}

func collectMsgsFromCmd(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	return collectMsgsFromMsg(cmd())
}

func collectMsgsFromMsg(msg tea.Msg) []tea.Msg {
	if msg == nil {
		return nil
	}

	// Both tea.BatchMsg and tea.sequenceMsg are slices of tea.Cmd.
	rv := reflect.ValueOf(msg)
	if rv.IsValid() && rv.Kind() == reflect.Slice {
		var out []tea.Msg
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i).Interface()
			cmd, ok := elem.(tea.Cmd)
			if !ok {
				continue
			}
			out = append(out, collectMsgsFromCmd(cmd)...)
		}
		return out
	}

	return []tea.Msg{msg}
}

func newFocusedAssetsModelWithAccount(t *testing.T, acc firefly.Account) modelAssets {
	t.Helper()

	api := &mockAssetAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			if accountType != "asset" {
				t.Fatalf("expected accountType 'asset', got %q", accountType)
			}
			return []firefly.Account{acc}
		},
		accountBalanceFunc: func(accountID string) float64 { return 0 },
	}

	m := newModelAssets(api)
	(&m).Focus()
	return m
}

var _ = prompt.PromptMsg{}

func TestGetAssetsItems_UsesBalanceAPI(t *testing.T) {
	api := &mockAssetAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			if accountType != "asset" {
				t.Fatalf("expected accountType 'asset', got %q", accountType)
			}
			return []firefly.Account{
				{ID: "a1", Name: "Checking", CurrencyCode: "USD", Type: "asset"},
				{ID: "a2", Name: "Savings", CurrencyCode: "EUR", Type: "asset"},
			}
		},
		accountBalanceFunc: func(accountID string) float64 {
			switch accountID {
			case "a1":
				return 10.5
			case "a2":
				return 99
			default:
				return 0
			}
		},
	}

	items := getAssetsItems(api)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	first, ok := items[0].(assetItem)
	if !ok {
		t.Fatalf("expected item type assetItem, got %T", items[0])
	}
	if first.account.ID != "a1" {
		t.Errorf("expected first account ID 'a1', got %q", first.account.ID)
	}
	if first.balance != 10.5 {
		t.Errorf("expected first balance 10.5, got %v", first.balance)
	}
	if first.Description() != "Balance: 10.50 USD" {
		t.Errorf("unexpected description: %q", first.Description())
	}
	if first.Title() != "Checking" {
		t.Errorf("expected title 'Checking', got %q", first.Title())
	}
}

func TestModelAssets_RefreshAssets_Success(t *testing.T) {
	api := &mockAssetAPI{}
	m := newModelAssets(api)

	_, cmd := m.Update(RefreshAssetsMsg{})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msg := cmd()
	if _, ok := msg.(AssetsUpdateMsg); !ok {
		t.Fatalf("expected AssetsUpdateMsg, got %T", msg)
	}

	if len(api.updateAccountsCalledWith) != 1 || api.updateAccountsCalledWith[0] != "asset" {
		t.Fatalf("expected UpdateAccounts called with 'asset', got %v", api.updateAccountsCalledWith)
	}
}

func TestModelAssets_RefreshAssets_Error(t *testing.T) {
	expectedErr := errors.New("boom")
	api := &mockAssetAPI{
		updateAccountsFunc: func(accountType string) error {
			return expectedErr
		},
	}
	m := newModelAssets(api)

	_, cmd := m.Update(RefreshAssetsMsg{})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msg := cmd()
	notifyMsg, ok := msg.(notify.NotifyMsg)
	if !ok {
		t.Fatalf("expected notify.NotifyMsg, got %T", msg)
	}
	if notifyMsg.Level != notify.Warn {
		t.Fatalf("expected warn level, got %v", notifyMsg.Level)
	}
	if notifyMsg.Message != expectedErr.Error() {
		t.Fatalf("expected message %q, got %q", expectedErr.Error(), notifyMsg.Message)
	}

	if len(api.updateAccountsCalledWith) != 1 || api.updateAccountsCalledWith[0] != "asset" {
		t.Fatalf("expected UpdateAccounts called with 'asset', got %v", api.updateAccountsCalledWith)
	}
}

func TestModelAssets_NewAsset_Error(t *testing.T) {
	expectedErr := errors.New("create failed")
	api := &mockAssetAPI{
		createAssetAccountFunc: func(name, currencyCode string) error {
			return expectedErr
		},
	}
	m := newModelAssets(api)

	_, cmd := m.Update(NewAssetMsg{Account: "My Asset", Currency: "usd"})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msg := cmd()
	notifyMsg, ok := msg.(notify.NotifyMsg)
	if !ok {
		t.Fatalf("expected notify.NotifyMsg, got %T", msg)
	}
	if notifyMsg.Level != notify.Warn {
		t.Fatalf("expected warn level, got %v", notifyMsg.Level)
	}
	if notifyMsg.Message != expectedErr.Error() {
		t.Fatalf("expected message %q, got %q", expectedErr.Error(), notifyMsg.Message)
	}

	if len(api.createAssetCalledWith) != 1 {
		t.Fatalf("expected CreateAssetAccount called once, got %d", len(api.createAssetCalledWith))
	}
	if api.createAssetCalledWith[0].name != "My Asset" || api.createAssetCalledWith[0].currency != "usd" {
		t.Fatalf("unexpected CreateAssetAccount args: %+v", api.createAssetCalledWith[0])
	}
}

func TestModelAssets_NewAsset_Success(t *testing.T) {
	api := &mockAssetAPI{}
	m := newModelAssets(api)

	_, cmd := m.Update(NewAssetMsg{Account: "My Asset", Currency: "usd"})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msgs := collectMsgsFromCmd(cmd)

	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d (%T)", len(msgs), msgs)
	}

	msg1 := msgs[0]
	if _, ok := msg1.(RefreshAssetsMsg); !ok {
		t.Fatalf("expected RefreshAssetsMsg, got %T", msg1)
	}
	msg2 := msgs[1]
	n, ok := msg2.(notify.NotifyMsg)
	if !ok {
		t.Fatalf("expected notify.NotifyMsg, got %T", msg2)
	}
	if n.Level != notify.Log {
		t.Fatalf("expected log notify level, got %v", n.Level)
	}
	if n.Message != "Asset account 'My Asset' created" {
		t.Fatalf("unexpected notify message: %q", n.Message)
	}

	if len(api.createAssetCalledWith) != 1 {
		t.Fatalf("expected CreateAssetAccount called once, got %d", len(api.createAssetCalledWith))
	}
	if api.createAssetCalledWith[0].name != "My Asset" || api.createAssetCalledWith[0].currency != "usd" {
		t.Fatalf("unexpected CreateAssetAccount args: %+v", api.createAssetCalledWith[0])
	}
}

func TestModelAssets_AssetsUpdate_EmitsDataLoadCompleted(t *testing.T) {
	api := &mockAssetAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			if accountType != "asset" {
				t.Fatalf("expected accountType 'asset', got %q", accountType)
			}
			// return []firefly.Account{}
			return []firefly.Account{
				{ID: "a1", Name: "Checking", CurrencyCode: "USD", Type: "asset"},
				{ID: "a2", Name: "Savings", CurrencyCode: "EUR", Type: "asset"},
			}
		},
		accountBalanceFunc: func(accountID string) float64 {
			switch accountID {
			case "a1":
				return 10.5
			case "a2":
				return 99
			default:
				return 0
			}
		},
	}
	m := newModelAssets(api)
	_, cmd := m.Update(AssetsUpdateMsg{})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msg := cmd()
	loader, ok := msg.(DataLoadCompletedMsg)
	if !ok {
		t.Fatalf("expected DataLoadCompletedMsg, got %T", msg)
	}
	if loader.DataType != "assets" {
		t.Fatalf("expected DataType 'assets', got %q", loader)
	}

	listItems := m.list.Items()
	if len(listItems) != 2 {
		t.Fatalf("expected 2 list items, got %d", len(listItems))
	}
}

func TestModelAssets_UpdatePositions_SetsListSize(t *testing.T) {
	globalWidth = 100
	globalHeight = 40
	topSize = 5
	summarySize = 10

	api := &mockAssetAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account { return nil },
	}
	m := newModelAssets(api)

	updated, _ := m.Update(UpdatePositions{})
	m2 := updated.(modelAssets)

	h, v := m2.styles.Base.GetFrameSize()
	wantW := globalWidth - h
	wantH := globalHeight - v - topSize - summarySize
	if m2.list.Width() != wantW {
		t.Fatalf("expected width %d, got %d", wantW, m2.list.Width())
	}
	if m2.list.Height() != wantH {
		t.Fatalf("expected height %d, got %d", wantH, m2.list.Height())
	}
}

func TestModelAssets_IgnoresKeysWhenNotFocused(t *testing.T) {
	api := &mockAssetAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{{ID: "a1", Name: "Checking", CurrencyCode: "USD", Type: "asset"}}
		},
		accountBalanceFunc: func(accountID string) float64 { return 0 },
	}
	m := newModelAssets(api) // focus is false by default
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	if cmd != nil {
		t.Fatalf("expected nil cmd when not focused, got %T", cmd)
	}
}

func TestModelAssets_KeyFilter_EmitsFilterMsgWithSelectedAccount(t *testing.T) {
	acc := firefly.Account{ID: "a1", Name: "Checking", CurrencyCode: "USD", Type: "asset"}
	api := &mockAssetAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account { return []firefly.Account{acc} },
		accountBalanceFunc: func(accountID string) float64 { return 0 },
	}
	m := newModelAssets(api)
	(&m).Focus()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	msgs := collectMsgsFromCmd(cmd)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 messages, got %d (%T)", len(msgs), msgs)
	}

	filter, ok := msgs[0].(FilterMsg)
	if !ok {
		t.Fatalf("expected FilterMsg, got %T", msgs[0])
	}
	if filter.Account.ID != "a1" {
		t.Fatalf("expected account ID 'a1', got %q", filter.Account.ID)
	}
}

func TestModelAssets_KeyRefresh_BatchesAssetsAndSummaryRefresh(t *testing.T) {
	api := &mockAssetAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account { return nil },
		accountBalanceFunc: func(accountID string) float64 { return 0 },
	}
	m := newModelAssets(api)

	(&m).Focus()

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	msgs := collectMsgsFromCmd(cmd)
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d (%T)", len(msgs), msgs)
	}

	_, ok := msgs[0].(RefreshAssetsMsg)
	if !ok {
		t.Fatalf("expected SetFocusedViewMsg, got %T", msgs[0])
	}
	if _, ok := msgs[1].(RefreshSummaryMsg); !ok {
		t.Fatalf("expected UpdatePositions, got %T", msgs[1])
	}
}

func TestModelAssets_KeyResetFilter_EmitsResetFilterMsg(t *testing.T) {
	api := &mockAssetAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account { return nil },
		accountBalanceFunc: func(accountID string) float64 { return 0 },
	}
	m := newModelAssets(api)
	(&m).Focus()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	if cmd == nil {
		t.Fatal("expected cmd")
	}
	msg := cmd()
	filter, ok := msg.(FilterMsg)
	if !ok {
		t.Fatalf("expected FilterMsg, got %T", msg)
	}
	if !filter.Reset {
		t.Fatalf("expected Reset=true, got %+v", filter)
	}
}

func TestModelAssets_View_UsesLeftPanelStyle(t *testing.T) {
	m := newFocusedAssetsModelWithAccount(t, firefly.Account{ID: "a1", Name: "Checking", CurrencyCode: "USD", Type: "asset"})

	// Make the style effect obvious.
	m.styles.LeftPanel = lipgloss.NewStyle().PaddingLeft(2).PaddingRight(3)

	got := m.View()
	want := m.styles.LeftPanel.Render(m.list.View())

	if got != want {
		t.Fatalf("unexpected view output")
	}
	if got == m.list.View() {
		t.Fatalf("expected left panel styling to change output")
	}
}

func TestModelAssets_KeyQuit_SetsTransactionsView(t *testing.T) {
	m := newFocusedAssetsModelWithAccount(t, firefly.Account{ID: "a1", Name: "Checking", CurrencyCode: "USD", Type: "asset"})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	msgs := collectMsgsFromCmd(cmd)

	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d (%T)", len(msgs), msgs)
	}

	focused, ok := msgs[0].(SetFocusedViewMsg)
	if !ok {
		t.Fatalf("expected SetFocusedViewMsg, got %T", msgs[0])
	}
	if focused.state != transactionsView {
		t.Fatalf("expected transactionsView, got %v", focused.state)
	}
	if _, ok := msgs[1].(UpdatePositions); !ok {
		t.Fatalf("expected UpdatePositions, got %T", msgs[1])
	}
}

func TestModelAssets_KeySelect_SequencesFilterAndView(t *testing.T) {
	acc := firefly.Account{ID: "a1", Name: "Checking", CurrencyCode: "USD", Type: "asset"}
	m := newFocusedAssetsModelWithAccount(t, acc)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	msgs := collectMsgsFromCmd(cmd)

	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d (%T)", len(msgs), msgs)
	}

	filter, ok := msgs[0].(FilterMsg)
	if !ok {
		t.Fatalf("expected FilterMsg, got %T", msgs[0])
	}
	if filter.Account.ID != acc.ID {
		t.Fatalf("expected account ID %q, got %q", acc.ID, filter.Account.ID)
	}

	focused, ok := msgs[1].(SetFocusedViewMsg)
	if !ok {
		t.Fatalf("expected SetFocusedViewMsg, got %T", msgs[1])
	}
	if focused.state != transactionsView {
		t.Fatalf("expected transactionsView, got %v", focused.state)
	}
	if _, ok := msgs[2].(UpdatePositions); !ok {
		t.Fatalf("expected UpdatePositions, got %T", msgs[2])
	}
}

func TestModelAssets_KeyNew_ReturnsPromptMsg(t *testing.T) {
	m := newFocusedAssetsModelWithAccount(t, firefly.Account{ID: "a1", Name: "Checking", CurrencyCode: "USD", Type: "asset"})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msg := cmd()
	p, ok := msg.(prompt.PromptMsg)
	if !ok {
		t.Fatalf("expected prompt.PromptMsg, got %T", msg)
	}
	if p.Prompt != "New Asset(<name>,<currency>): " {
		t.Fatalf("unexpected prompt: %q", p.Prompt)
	}
	if p.Callback == nil {
		t.Fatal("expected callback")
	}
}

func TestModelAssets_KeyViewCategories_SetsCategoriesView(t *testing.T) {
	m := newFocusedAssetsModelWithAccount(t, firefly.Account{ID: "a1", Name: "Checking", CurrencyCode: "USD", Type: "asset"})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	msgs := collectMsgsFromCmd(cmd)

	focused, ok := msgs[0].(SetFocusedViewMsg)
	if !ok {
		t.Fatalf("expected SetFocusedViewMsg, got %T", msgs[0])
	}
	if focused.state != categoriesView {
		t.Fatalf("expected categoriesView, got %v", focused.state)
	}
}

func TestModelAssets_KeyViewExpenses_SetsExpensesView(t *testing.T) {
	m := newFocusedAssetsModelWithAccount(t, firefly.Account{ID: "a1", Name: "Checking", CurrencyCode: "USD", Type: "asset"})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	msgs := collectMsgsFromCmd(cmd)

	focused, ok := msgs[0].(SetFocusedViewMsg)
	if !ok {
		t.Fatalf("expected SetFocusedViewMsg, got %T", msgs[0])
	}
	if focused.state != expensesView {
		t.Fatalf("expected expensesView, got %v", focused.state)
	}
}

func TestModelAssets_KeyViewRevenues_SetsRevenuesView(t *testing.T) {
	m := newFocusedAssetsModelWithAccount(t, firefly.Account{ID: "a1", Name: "Checking", CurrencyCode: "USD", Type: "asset"})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	msgs := collectMsgsFromCmd(cmd)

	focused, ok := msgs[0].(SetFocusedViewMsg)
	if !ok {
		t.Fatalf("expected SetFocusedViewMsg, got %T", msgs[0])
	}
	if focused.state != revenuesView {
		t.Fatalf("expected revenuesView, got %v", focused.state)
	}
}

func TestModelAssets_KeyViewLiabilities_SetsLiabilitiesView(t *testing.T) {
	m := newFocusedAssetsModelWithAccount(t, firefly.Account{ID: "a1", Name: "Checking", CurrencyCode: "USD", Type: "asset"})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	msgs := collectMsgsFromCmd(cmd)

	focused, ok := msgs[0].(SetFocusedViewMsg)
	if !ok {
		t.Fatalf("expected SetFocusedViewMsg, got %T", msgs[0])
	}
	if focused.state != liabilitiesView {
		t.Fatalf("expected liabilitiesView, got %v", focused.state)
	}
}

func TestModelAssets_KeyViewTransactions_SetsTransactionsView(t *testing.T) {
	m := newFocusedAssetsModelWithAccount(t, firefly.Account{ID: "a1", Name: "Checking", CurrencyCode: "USD", Type: "asset"})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	msgs := collectMsgsFromCmd(cmd)

	focused, ok := msgs[0].(SetFocusedViewMsg)
	if !ok {
		t.Fatalf("expected SetFocusedViewMsg, got %T", msgs[0])
	}
	if focused.state != transactionsView {
		t.Fatalf("expected transactionsView, got %v", focused.state)
	}
}

func TestModelAssets_KeyViewAssets_IsDisabled(t *testing.T) {
	m := newFocusedAssetsModelWithAccount(t, firefly.Account{ID: "a1", Name: "Checking", CurrencyCode: "USD", Type: "asset"})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if cmd == nil {
		return
	}

	msgs := collectMsgsFromCmd(cmd)
	for _, msg := range msgs {
		if _, ok := msg.(SetFocusedViewMsg); ok {
			t.Fatalf("expected no view change, got %T", msg)
		}
	}
}

func TestCmdPromptNewAsset_EmitsPromptMsgWithCallback(t *testing.T) {
	backCmd := Cmd(SetFocusedViewMsg{state: assetsView})
	cmd := CmdPromptNewAsset(backCmd)

	msg := cmd()
	p, ok := msg.(prompt.PromptMsg)
	if !ok {
		t.Fatalf("expected prompt.PromptMsg, got %T", msg)
	}
	if p.Prompt != "New Asset(<name>,<currency>): " {
		t.Fatalf("unexpected prompt: %q", p.Prompt)
	}
	if p.Callback == nil {
		t.Fatal("expected callback")
	}
}

func TestCmdPromptNewAsset_CallbackValid_EmitsNewAssetMsgAndBackCmd(t *testing.T) {
	backCmd := Cmd(SetFocusedViewMsg{state: assetsView})
	cmd := CmdPromptNewAsset(backCmd)

	p := cmd().(prompt.PromptMsg)
	msgs := collectMsgsFromCmd(p.Callback("  My Asset  , usd  "))

	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d (%T)", len(msgs), msgs)
	}

	newAsset, ok := msgs[0].(NewAssetMsg)
	if !ok {
		t.Fatalf("expected NewAssetMsg, got %T", msgs[0])
	}
	if newAsset.Account != "My Asset" {
		t.Fatalf("expected account 'My Asset', got %q", newAsset.Account)
	}
	if newAsset.Currency != "usd" {
		t.Fatalf("expected currency 'usd', got %q", newAsset.Currency)
	}

	focused, ok := msgs[1].(SetFocusedViewMsg)
	if !ok {
		t.Fatalf("expected SetFocusedViewMsg, got %T", msgs[1])
	}
	if focused.state != assetsView {
		t.Fatalf("expected assetsView, got %v", focused.state)
	}
}

func TestCmdPromptNewAsset_CallbackInvalid_EmitsWarnAndBackCmd(t *testing.T) {
	backCmd := Cmd(SetFocusedViewMsg{state: assetsView})
	cmd := CmdPromptNewAsset(backCmd)

	p := cmd().(prompt.PromptMsg)
	msgs := collectMsgsFromCmd(p.Callback("invalid"))

	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d (%T)", len(msgs), msgs)
	}

	n, ok := msgs[0].(notify.NotifyMsg)
	if !ok {
		t.Fatalf("expected notify.NotifyMsg, got %T", msgs[0])
	}
	if n.Level != notify.Warn {
		t.Fatalf("expected warn notify level, got %v", n.Level)
	}
	if n.Message != "Invalid asset name or currency" {
		t.Fatalf("unexpected notify message: %q", n.Message)
	}

	focused, ok := msgs[1].(SetFocusedViewMsg)
	if !ok {
		t.Fatalf("expected SetFocusedViewMsg, got %T", msgs[1])
	}
	if focused.state != assetsView {
		t.Fatalf("expected assetsView, got %v", focused.state)
	}
}

func TestCmdPromptNewAsset_CallbackNone_EmitsOnlyBackCmd(t *testing.T) {
	backCmd := Cmd(SetFocusedViewMsg{state: assetsView})
	cmd := CmdPromptNewAsset(backCmd)

	p := cmd().(prompt.PromptMsg)
	msgs := collectMsgsFromCmd(p.Callback("None"))

	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d (%T)", len(msgs), msgs)
	}
	focused, ok := msgs[0].(SetFocusedViewMsg)
	if !ok {
		t.Fatalf("expected SetFocusedViewMsg, got %T", msgs[0])
	}
	if focused.state != assetsView {
		t.Fatalf("expected assetsView, got %v", focused.state)
	}
}
