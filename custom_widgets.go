// package main

// import (
//     "fyne.io/fyne/v2/widget"
// )

// type MyEntry struct {
//     widget.Entry
//     OnFocusLost func()
// }

// func NewMyEntry() *MyEntry {
//     e := &MyEntry{}
//     e.ExtendBaseWidget(e)
//     return e
// }

// func (e *MyEntry) FocusLost() {
//     if e.OnFocusLost != nil {
//         e.OnFocusLost()
//     }
// }