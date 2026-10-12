package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// ---------- simulatie (zelfde regels als cpu.go) ----------

var elementsDepthArray [256][4]byte
var moveToIndexArray [256]int

func init() {
	elementsDepthArray['W'] = [4]byte{'L', 'A', 'V', 'W'}
	elementsDepthArray['V'] = [4]byte{'W', 'L', 'A', 'V'}
	elementsDepthArray['A'] = [4]byte{'V', 'W', 'L', 'A'}
	elementsDepthArray['L'] = [4]byte{'A', 'V', 'W', 'L'}
	moveToIndexArray['W'] = 0
	moveToIndexArray['V'] = 1
	moveToIndexArray['A'] = 2
	moveToIndexArray['L'] = 3
	moveToIndexArray['D'] = 4
}

var moveWins = [5][5]uint8{
	{0, 1, 0, 2, 0},
	{2, 0, 1, 0, 0},
	{0, 2, 0, 1, 0},
	{1, 0, 2, 0, 0},
	{0, 0, 0, 0, 0},
}

var depthToElement = [5]byte{'W', 'V', 'A', 'L', 'D'}

type Player struct {
	available [5]int
	moves     [13]byte
	moveCount int
}

func getElementFromCode(depth int) byte {
	if depth < 1 || depth > 5 {
		return 0
	}
	return depthToElement[depth-1]
}

func getElementByDepth(prev byte, depth int) byte {
	if depth == 5 {
		return 'D'
	}
	if prev == 0 {
		return 0
	}
	if prev == 'D' {
		prev = 'L'
	}
	return elementsDepthArray[prev][depth-1]
}

func chooseAvailableElement(target byte, available *[5]int) byte {
	if available[moveToIndexArray[target]] > 0 {
		return target
	}
	current := target
	for i := 0; i < 5; i++ {
		current = elementsDepthArray[current][0]
		if available[moveToIndexArray[current]] > 0 {
			return current
		}
	}
	if available[4] > 0 {
		return 'D'
	}
	return 0
}

func getLastElement(available *[5]int) byte {
	for i, c := range depthToElement {
		if available[i] > 0 {
			return c
		}
	}
	return 0
}

func determineWinner(m1, m2 byte) int {
	return int(moveWins[moveToIndexArray[m1]][moveToIndexArray[m2]])
}

// resolveTarget bepaalt het gewenste element voor zet i: een letter (WVALD) is een vaste zet, een cijfer een diepte
// en de zetten van de tegenstander.
func resolveTarget(ch byte, i int, oppMoves []byte) byte {
	if isMoveLetter(ch) {
		return ch // vaste zet (letter)
	}
	digit := int(ch - '0')
	if digit >= 6 {
		base := digit - 5
		if i < 2 {
			return getElementFromCode(base)
		}
		return getElementByDepth(oppMoves[i-2], base)
	}
	if i == 0 {
		return getElementFromCode(digit)
	}
	return getElementByDepth(oppMoves[i-1], digit)
}

// simulateDepthGame: twee diepte-engines (12 cijfers) tegen elkaar, volledige score.
func simulateDepthGame(e1, e2 string) (s1, s2 int) {
	var p1, p2 Player
	p1.available = [5]int{3, 3, 3, 3, 1}
	p2.available = [5]int{3, 3, 3, 3, 1}
	for i := 0; i < 12; i++ {
		t1 := resolveTarget(e1[i], i, p2.moves[:])
		t2 := resolveTarget(e2[i], i, p1.moves[:])
		m1 := chooseAvailableElement(t1, &p1.available)
		m2 := chooseAvailableElement(t2, &p2.available)
		if m1 == 0 || m2 == 0 {
			return -1, -1
		}
		p1.available[moveToIndexArray[m1]]--
		p1.moves[i] = m1
		p2.available[moveToIndexArray[m2]]--
		p2.moves[i] = m2
		switch determineWinner(m1, m2) {
		case 1:
			s1++
		case 2:
			s2++
		}
	}
	m1 := getLastElement(&p1.available)
	m2 := getLastElement(&p2.available)
	switch determineWinner(m1, m2) {
	case 1:
		s1++
	case 2:
		s2++
	}
	return s1, s2
}

// simulateDepthVsFixed: diepte-engine reageert op een vaste 13-zet engine.
func simulateDepthVsFixed(engine, fixed string) (s1, s2 int) {
	p := Player{available: [5]int{3, 3, 3, 3, 1}}
	var moves [13]byte
	for i := 0; i < 12; i++ {
		t := resolveTarget(engine[i], i, []byte(fixed))
		m := chooseAvailableElement(t, &p.available)
		if m == 0 {
			m = getLastElement(&p.available)
		}
		if m != 0 {
			p.available[moveToIndexArray[m]]--
		}
		moves[i] = m
	}
	m := getLastElement(&p.available)
	if m == 0 {
		m = 'W'
	}
	moves[12] = m
	for i := 0; i < 13; i++ {
		switch determineWinner(moves[i], fixed[i]) {
		case 1:
			s1++
		case 2:
			s2++
		}
	}
	return s1, s2
}

func isMoveLetter(c byte) bool {
	return c == 'W' || c == 'V' || c == 'A' || c == 'L' || c == 'D'
}

// isValidDepthCode aanvaardt 12 tekens: eerst letters (WVALD, vaste zet), dan cijfers (1-9, diepte).
func isValidDepthCode(engine string) bool {
	if len(engine) != 12 {
		return false
	}
	seenDigit := false
	for i := 0; i < len(engine); i++ {
		c := engine[i]
		switch {
		case c >= '1' && c <= '9':
			seenDigit = true
		case isMoveLetter(c):
			if seenDigit {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func validCode(e string) bool {
	return isValidDepthCode(e) || (len(e) == 13 && strings.Trim(e, "WVALD") == "")
}

// parseEngineLine haalt naam en code uit "Naam:level:code", "code" of "code (score: N)".
func parseEngineLine(line string) (name, code string, ok bool) {
	parts := strings.Split(line, ":")
	engine := strings.TrimSpace(line)
	if len(parts) > 2 {
		name = strings.TrimSpace(parts[0])
		engine = strings.TrimSpace(parts[2])
	}
	if isFixed13 := len(engine) == 13 && strings.Trim(engine, "WVALD") == ""; len(engine) > 12 && !isFixed13 {
		engine = engine[:12]
	}
	if !validCode(engine) {
		return "", engine, false
	}
	if name == "" {
		name = engine
	}
	return name, engine, true
}

// engineHits laat een engine de eerste n zetten spelen tegen de tegenstander (opp) en telt de
// hits: de zetten die gelijk zijn aan die van de gezochte speler (target). Een vaste engine
// (13 letters) speelt altijd hetzelfde; een diepte-engine reageert op opp. Zodra de engine
// niet meer op minHits kan komen, stopt de berekening meteen (ok = false) en gaan we door met
// de volgende engine.
func engineHits(engine string, opp, target []byte, n, minHits int) (moves [13]byte, hits int, ok bool) {
	misses := 0
	if len(engine) == 13 {
		for i := 0; i < n; i++ {
			moves[i] = engine[i]
			if moves[i] == target[i] {
				hits++
			} else {
				misses++
				if n-misses < minHits {
					return moves, hits, false
				}
			}
		}
		return moves, hits, true
	}
	p := Player{available: [5]int{3, 3, 3, 3, 1}}
	for i := 0; i < n; i++ {
		var m byte
		if i < 12 {
			t := resolveTarget(engine[i], i, opp)
			m = chooseAvailableElement(t, &p.available)
			if m == 0 {
				m = getLastElement(&p.available)
			}
		} else {
			m = getLastElement(&p.available) // 13e zet: het resterende element
			if m == 0 {
				m = 'W'
			}
		}
		if m != 0 {
			p.available[moveToIndexArray[m]]--
		}
		moves[i] = m
		if m == target[i] {
			hits++
		} else {
			misses++
			if n-misses < minHits {
				return moves, hits, false
			}
		}
	}
	return moves, hits, true
}

// parseMoves leest 1 tot 13 zetten, ook met prefix: "Player1: WAVLDWAVLWAVL".
func parseMoves(input string) (string, error) {
	if i := strings.LastIndex(input, ":"); i >= 0 {
		input = input[i+1:]
	}
	var sb strings.Builder
	for _, c := range strings.ToUpper(input) {
		if c == ' ' || c == ',' || c == '\t' {
			continue
		}
		sb.WriteRune(c)
	}
	s := sb.String()
	if len(s) < 1 || len(s) > 13 || strings.Trim(s, "WVALD") != "" {
		return "", fmt.Errorf("moet 1 tot 13 zetten zijn met alleen W, V, A, L, D (kreeg '%s')", s)
	}
	return s, nil
}

// movesWarning meldt als een speler een element te vaak speelt (max 3x W, V, A, L en 1x D),
// of bij 13 zetten niet precies die verdeling heeft.
func movesWarning(s string) string {
	want := map[byte]int{'W': 3, 'V': 3, 'A': 3, 'L': 3, 'D': 1}
	got := map[byte]int{}
	for i := 0; i < len(s); i++ {
		got[s[i]]++
	}
	for c, n := range want {
		if got[c] > n || (len(s) == 13 && got[c] != n) {
			return "let op: een speler speelt normaal 3x W, 3x V, 3x A, 3x L en 1x D"
		}
	}
	return ""
}

func readLine(sc *bufio.Scanner, prompt string) (string, bool) {
	fmt.Print(prompt)
	if !sc.Scan() {
		return "", false
	}
	return strings.TrimSpace(sc.Text()), true
}

type candidate struct {
	name, code string
	moves      [13]byte
	hits       int
	streak     int // langste reeks hits die direct na elkaar komen
	pairs      int // aantal paren van twee hits direct naast elkaar
}

// streakStats geeft de langste reeks aaneengesloten hits en het aantal aangrenzende hit-paren.
func streakStats(moves [13]byte, target []byte, n int) (longest, pairs int) {
	run := 0
	for i := 0; i < n; i++ {
		if moves[i] == target[i] {
			run++
			if run > longest {
				longest = run
			}
			if run > 1 {
				pairs++
			}
		} else {
			run = 0
		}
	}
	return longest, pairs
}

func main() {
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)

	for {
		l1, ok := readLine(sc, "Zetten van Player 1 (1-13 zetten WVALD, leeg om te stoppen): ")
		if !ok || l1 == "" {
			fmt.Println("Gestopt.")
			return
		}
		p1, err := parseMoves(l1)
		if err != nil {
			fmt.Print("Player 1: ", err, "\n\n")
			continue
		}
		l2, ok := readLine(sc, "Zetten van Player 2: ")
		if !ok {
			return
		}
		p2, err := parseMoves(l2)
		if err != nil {
			fmt.Print("Player 2: ", err, "\n\n")
			continue
		}
		for i, s := range []string{p1, p2} {
			if w := movesWarning(s); w != "" {
				fmt.Printf("Player %d: %s\n", i+1, w)
			}
		}
		// beide spelers hebben evenveel zetten gespeeld; bij een verschil telt de kortste
		n := len(p1)
		if len(p2) < n {
			n = len(p2)
		}
		if len(p1) != len(p2) {
			fmt.Printf("Let op: Player 1 heeft %d zetten en Player 2 %d; we gebruiken de eerste %d.\n", len(p1), len(p2), n)
		}
		p1, p2 = p1[:n], p2[:n]

		var target, opp string
		for {
			l, ok := readLine(sc, "Voor welke speler moet de engine gevonden worden? (1 of 2): ")
			if !ok {
				return
			}
			if l == "1" {
				target, opp = p1, p2
				break
			}
			if l == "2" {
				target, opp = p2, p1
				break
			}
			fmt.Println("Geef 1 of 2.")
		}

		fmt.Println("Voer alle mogelijke engines in (Naam:level:code of code, één per regel; een .txt-bestand mag ook; '.' of lege regel om te stoppen):")
		var cands []candidate
		invalid := 0
		addLine := func(line string) {
			name, code, ok := parseEngineLine(line)
			if !ok {
				invalid++
				return
			}
			cands = append(cands, candidate{name: name, code: code})
		}
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || line == "." {
				break
			}
			// een bestandsnaam: lees daar alle engines uit
			if !strings.Contains(line, ":") && strings.HasSuffix(strings.ToLower(line), ".txt") {
				if f, err := os.Open(line); err == nil {
					fs := bufio.NewScanner(f)
					fs.Buffer(make([]byte, 1024*1024), 1024*1024)
					for fs.Scan() {
						if l := strings.TrimSpace(fs.Text()); l != "" {
							addLine(l)
						}
					}
					f.Close()
					continue
				}
			}
			addLine(line)
		}
		if invalid > 0 {
			fmt.Printf("%d regels overgeslagen (ongeldige engine code).\n", invalid)
		}
		if len(cands) == 0 {
			fmt.Print("Geen engines ingevoerd.\n\n")
			continue
		}

		// Elke engine speelt de eerste n zetten. Kan een engine het huidige beste aantal hits
		// niet meer halen, dan stoppen we meteen en gaan we naar de volgende engine.
		oppB, targetB := []byte(opp), []byte(target)
		best := 0
		var found []candidate // alleen de engines met het hoogste aantal hits
		stopped := 0
		for _, c := range cands {
			moves, hits, ok := engineHits(c.code, oppB, targetB, n, best)
			if !ok {
				stopped++
				continue
			}
			if hits > best {
				best = hits
				found = found[:0]
			}
			if hits == best {
				c.moves, c.hits = moves, hits
				c.streak, c.pairs = streakStats(moves, targetB, n)
				found = append(found, c)
			}
		}
		// bij gelijk aantal hits wint de engine met de langste reeks hits direct na elkaar,
		// daarna die met de meeste aangrenzende hit-paren; blijft het gelijk dan de invoervolgorde
		sort.SliceStable(found, func(i, j int) bool {
			if found[i].streak != found[j].streak {
				return found[i].streak > found[j].streak
			}
			return found[i].pairs > found[j].pairs
		})

		fmt.Printf("\n%d engines getest (%d vroeg gestopt). Beste: %d/%d hits (%d engine(s)).\n", len(cands), stopped, best, n, len(found))
		fmt.Printf("Doel:     %s\n\n", target)
		show := len(found)
		const maxShow = 10
		if show > maxShow {
			show = maxShow
		}
		for i := 0; i < show; i++ {
			c := found[i]
			var marks strings.Builder
			for j := 0; j < n; j++ {
				if c.moves[j] == target[j] {
					marks.WriteByte('|')
				} else {
					marks.WriteByte(' ')
				}
			}
			fmt.Printf("%d. %s (%s)  %d/%d, langste reeks hits: %d\n", i+1, c.name, c.code, c.hits, n, c.streak)
			fmt.Printf("   Gespeeld: %s\n", string(c.moves[:n]))
			fmt.Printf("   Hits:     %s\n", marks.String())
		}
		if len(found) > show {
			fmt.Printf("... en nog %d engines met %d/%d hits\n", len(found)-show, best, n)
		}
		fmt.Println()
	}
}
