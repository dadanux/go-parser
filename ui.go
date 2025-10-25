package main

import (
	"bufio"
	"image/color"
	"regexp"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// regex pour trouver "machine: name" ou "machine = name" (tolérant, capture le name)
var machineRe = regexp.MustCompile(`(?i)machine\s*[:=]\s*"?([^"\n]+?)"?\b`)
type MyEntry struct {
    widget.Entry
    OnFocusLost func()
}

func NewMyEntry() *MyEntry {
    e := &MyEntry{}
    e.ExtendBaseWidget(e)
    return e
}

func (e *MyEntry) FocusLost() {
    if e.OnFocusLost != nil {
        e.OnFocusLost()
    }
}
func main() {
	a := app.New()
	w := a.NewWindow("Creation d'un template pour Ansible à partir d'un fichier Autosys")

	initialText := "variable ex: machine_iis" // texte affiché quand pas de valeur définie
	// affichage read-only du fichier avec contraste augmenté (fond blanc, texte noir)
	contentLbl := widget.NewLabel("")
	contentLbl.Wrapping = fyne.TextWrapOff
	bg := canvas.NewRectangle(color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF})
	// bord sombre pour bien délimiter
	bg.StrokeColor = color.NRGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xFF}
	bg.StrokeWidth = 1
	// Scroll contenant le rectangle de fond et le label (label au-dessus)
	inputView := container.NewScroll(container.NewMax(bg, contentLbl))
	// optionnel : forcer taille minimale visible
	inputView.SetMinSize(fyne.NewSize(200, 200))

	output := widget.NewMultiLineEntry()
	output.SetPlaceHolder("Prévisualisation après remplacements")
	output.Wrapping = fyne.TextWrapOff

	// fonction de traitement réutilisable (conserve votre logique start_times)
	process := func(text string) {
		lines := strings.Split(text, "\n")
		var outputLines []string
		for _, line := range lines {
			if strings.HasPrefix(line, "owner:") {
				outputLines = append(outputLines, "owner: {{ owner }}")
			} else if strings.HasPrefix(line, "insert_job:") {
				re := regexp.MustCompile(`insert_job:\s*([^\s]+)`)
				m := re.FindStringSubmatch(line)
				if len(m) >= 2 {
					outputLines = append(outputLines, "insert_job: {{ prefix }}_"+m[1])
				}
			} else if strings.HasPrefix(line, "box_name:") {
				re := regexp.MustCompile(`box_name\s*:\s*"?([^"\s]+)"?`)
				m := re.FindStringSubmatch(line)
				if len(m) >= 2 {
					outputLines = append(outputLines, "box_name: {{ prefix }}_"+m[1])
				}

			} else {
				outputLines = append(outputLines, line)
			}
		}
		// afficher le résultat traité dans la zone output (prévisualisation)
		output.SetText(strings.Join(outputLines, "\n"))
	}

	// données pour la liste de machines
	var machineList []string
	replacements := map[string]string{} // map machine -> valeur de remplacement
	sourceText := ""

	// container qui affiche toutes les machines (sans scroll)
	machinesContainer := container.NewVBox()

	// applique toutes les replacements sur sourceText et met à jour output (prévisualisation)
	applyReplacements := func() {
		newText := machineRe.ReplaceAllStringFunc(sourceText, func(m string) string {
			sub := machineRe.FindStringSubmatch(m)
			if len(sub) >= 2 {
				name := strings.TrimSpace(sub[1])
				if r, ok := replacements[name]; ok && r != "" {
					return strings.Replace(m, sub[1], "{{ prefix | lower }}_{{"+r+"}}_a", 1)
				}
			}
			return m
		})
		// NE PAS modifier input (fichier original). Mettre la prévisualisation dans output.
		process(newText)
	}

	// reconstruit la liste visible des machines (sans scrollbar)
	updateMachineList := func(text string) {
		matches := machineRe.FindAllStringSubmatch(text, -1)
		set := make(map[string]struct{})
		for _, m := range matches {
			if len(m) >= 2 {
				name := strings.TrimSpace(m[1])
				if name != "" {
					set[name] = struct{}{}
				}
			}
		}
		// reconstruire machineList triée
		machineList = machineList[:0]
		for k := range set {
			machineList = append(machineList, k)
		}
		sort.Strings(machineList)

		// garantir les clés dans replacements
		for _, k := range machineList {
			if _, ok := replacements[k]; !ok {
				replacements[k] = ""
			}
		}

		// vider et reconstruire le container (chaque ligne : nom + valeur éditable)
		machinesContainer.Objects = nil
		for _, name := range machineList {
			// label nom
			nameLbl := widget.NewLabel(name)
			nameLbl.Alignment = fyne.TextAlignLeading
			// wrapper pour forcer largeur colonne "nom" via GridWrapLayout (largeur fixe)
			nameBox := container.New(layout.NewGridWrapLayout(fyne.NewSize(220, 30)), nameLbl)

			// label affichage valeur (ou valeur par défaut)
			val := replacements[name]
			var valText string
			if val == "" {
				// valeur par défaut demandée
				valText = initialText
			} else {
				valText = ""
			}
			valLbl := widget.NewButton(valText, nil) // bouton pour activer édition
			// entry pour édition inline, cachée initialement
			valEntry := NewMyEntry()
			valEntry.SetText(val)
			valEntry.Hide()

			// stack valLbl + valEntry puis wrapper pour largeur de colonne "remplacement"
			valStack := container.NewMax(valLbl, valEntry)
			valBox := container.New(layout.NewGridWrapLayout(fyne.NewSize(200, 30)), valStack)

			// activation de l'édition : swap label -> entry
			valLbl.OnTapped = func(n string, lbl *widget.Button, ent *MyEntry) func() {
				return func() {
					lbl.Hide()
					ent.Show()
					// focus sur l'entry
					w.Canvas().Focus(ent)
				}
			}(name, valLbl, valEntry)

			// on submit : enregistrer et appliquer, revenir affichage label
			valEntry.OnSubmitted = func(n string, ent *MyEntry, lbl *widget.Button) func(string) {
				return func(s string) {
					replacements[n] = s
					if s == "" {
						lbl.SetText(initialText)
					} else {
						lbl.SetText(s)
					}
					ent.Hide()
					lbl.Show()
					applyReplacements()
				}
			}(name, valEntry, valLbl)

valEntry.OnFocusLost = func() {
    replacements[name] = valEntry.Text
    if valEntry.Text == "" {
        valLbl.SetText(initialText)
    } else {
        valLbl.SetText(valEntry.Text)
    }
    valEntry.Hide()
    valLbl.Show()
    applyReplacements()
}
    
			// perte de focus / modification : mise à jour live (optionnelle) et appliquer
			valEntry.OnChanged = func(n string, ent *MyEntry, lbl *widget.Button) func(string) {
				return func(s string) {
					replacements[n] = s
					if s == "" {
						lbl.SetText(initialText)
					} else {
						lbl.SetText(s)
					}
	    			applyReplacements()
				}
			}(name, valEntry, valLbl)


			// ligne : nom | remplacement
			row := container.NewHBox(
				nameBox,
				valBox,
			)
			row.Refresh()
			machinesContainer.Add(row)
		}
		machinesContainer.Refresh()
	}


	loadBtn := widget.NewButton("Charger fichier jil", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			scanner := bufio.NewScanner(reader)
			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			text := strings.Join(lines, "\n")
			sourceText = text

			contentLbl.SetText(sourceText)
			contentLbl.Refresh()

			replacements = map[string]string{}
			updateMachineList(sourceText)
			applyReplacements()
		}, w)
	})



	machineControls := container.NewBorder(nil, nil, nil, nil,
		container.NewHBox(widget.NewLabel("Machines:")), // header visuel
	)
	controlsWithList := container.NewBorder(machineControls, nil, nil, nil, machinesContainer)

	// gauche : load button en haut + controls (liste) puis input rempli le reste
	leftTop := container.NewVBox(loadBtn, controlsWithList)

	left := container.NewBorder(leftTop, nil, nil, nil, inputView)

	// split 50/50
	content := container.NewHSplit(left, output)
	content.SetOffset(0.5)

	w.SetContent(content)
	w.Resize(fyne.NewSize(1000, 700))
	w.ShowAndRun()
}
