package main

import (
	"encoding/json"
	"os"
)

const saveFile = "zashecoins_save.json"
const startBalance = 1000
const depositAmount = 1000     // сумма пополнения по кнопке «Внести додеп»
const depositMaxBalance = 2000 // додеп доступен, только если баланс строго ниже этого

// State — постоянное состояние игрока (сохраняется на диск).
type State struct {
	Balance     int `json:"balance"`
	TotalSpins  int `json:"total_spins"`
	BiggestWin  int `json:"biggest_win"`
	CasesOpened int `json:"cases_opened"`
}

func loadState() *State {
	s := &State{Balance: startBalance}
	data, err := os.ReadFile(saveFile)
	if err == nil {
		_ = json.Unmarshal(data, s)
	}
	if s.Balance <= 0 && err != nil {
		s.Balance = startBalance
	}
	return s
}

func (s *State) save() {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(saveFile, data, 0o644)
}

func (s *State) recordWin(amount int) {
	if amount > s.BiggestWin {
		s.BiggestWin = amount
	}
}
