package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// Stack type aliases
type Bstack []byte
type Fstack []Frag

// Cool stack that works shut up
// Based on https://github.com/golang/go/wiki/SliceTricks
func (bs *Bstack) push(b byte) {
	*bs = append(*bs, b)
}

func (bs *Bstack) pop() (b byte) {
	b, *bs = (*bs)[len(*bs)-1], (*bs)[:len(*bs)-1]
	return
}

func (bs *Bstack) peek() (b byte) {
	return (*bs)[len(*bs)-1]
}

// TODO: Refactor and figure out interfaces
func (fs *Fstack) push(f Frag) {
	*fs = append(*fs, f)
}

func (fs *Fstack) pop() (f Frag) {
	f, *fs = (*fs)[len(*fs)-1], (*fs)[:len(*fs)-1]
	return
}

// Basic preprocessing:
// convert {n} to n appearances of preceding character,
// addes explicit '.' concat operator
//
// TODO: Make {n} work with () groups
func preProcess(s string) (string, error) {
	for {
		// Current block to modify
		openIdx := strings.Index(s, "{")
		closeIdx := strings.Index(s, "}")

		// Nothing to do
		if openIdx == -1 || closeIdx == -1 || openIdx == 0 {
			break
		}

		// Get the number inside the {}
		num, err := strconv.Atoi(s[openIdx+1 : closeIdx])
		if err != nil || num <= 0 {
			break
		}

		// Build replacement string
		newStr := ""
		for i := 1; i <= num; i++ {
			newStr += s[openIdx-1 : openIdx]
		}

		//fmt.Printf("newStr: %v\n", newStr)

		s = s[:openIdx-1] + newStr + s[closeIdx+1:]

		//return s
	}

	// check for stray {}
	openIdx := strings.Index(s, "{")
	closeIdx := strings.Index(s, "}")
	if openIdx != -1 || closeIdx != -1 {
		err := errors.New("Malformed or stray {}'s")
		return "", err
	}

	// Now add explicit concatenation:
	// That is, any character followed by
	// anything other than ),+,*,| add a .
	// dontconcat := ")+*|"

	// Start with second char
	newStr := string(s[0])
	for i := 1; i < len(s); i++ {
		switch s[i] {
		// Never precede with a '.'
		case ')', '+', '*', '|':
			newStr += string(s[i])

		default:
			switch s[i-1] {
			// Don't append a '.' after these
			case '(', '|':
				newStr += string(s[i])

			default:
				newStr += "." + string(s[i])
			}

		}
	}

	return newStr, nil

}

// Converts infix regex to postfix
// Uses Shunting-yard algorithm to convert
// to Reverse Polish Notation
func toPostfix(infix string) (string, error) {

	strLen := len(infix)

	// String to build
	// acts like a push-only stack
	postfix := ""
	// operator stack
	stack := make(Bstack, 0, strLen)

	// Iterate through string
	for i := 0; i < strLen; i++ {

		// current char
		c := infix[i]

		// if-else chains are for nerds
		// cool kids use switches
		switch c {
		case '(':
			stack.push(c)

		// Add the stuff in the () group
		case ')':
			for stack.peek() != '(' {
				postfix += string(stack.pop())
				// malformed parens
				if len(stack) == 0 {
					return "", errors.New("Mismatched parentheses")
				}
			}
			stack.pop() // toss out the '('

		// Deal with the stack
		default:
			for len(stack) != 0 {
				// Top element
				peekChar := stack.peek()

				// Get the precedence of read char and peeked char
				peekedCharPrec := precedenceOf(peekChar)
				currCharPrec := precedenceOf(c)

				// if the peeked char has a >= precedence
				// to the read char, pop the peaked char
				// and add to the output string
				if peekedCharPrec >= currCharPrec {
					postfix += string(stack.pop())
				} else {
					break
				}
			}
			// Add the current char to the stack
			stack.push(c)
		}

	}

	// Deal with the remaining operators in the stack
	for len(stack) != 0 {
		if stack.peek() == '(' || stack.peek() == ')' {
			return "", errors.New("Mismatched parentheses")
		}
		postfix += string(stack.pop())
	}

	return postfix, nil

}

// Regex precedence 'chart'
func precedenceOf(c byte) int {
	switch c {
	case '(':
		return 1
	case '|':
		return 2
	case '.':
		return 3
	case '?', '*', '+':
		return 4
	case '^':
		return 5
	default:
		return 6
	}
}

// NFA Data structures
type State struct {
	char     byte
	stype    int
	out      *State
	out1     *State
	lastlist int
}

// stype's
const (
	NORMAL = iota
	SPLIT
	ACCEPT
)

// NFA Fragment,
// a state and dangling
// out arrows
type Frag struct {
	start *State
	out   OutList
}

// A list of 'out' pointers,
// that is, dangling arrows during
// NFA constructions
type OutList []**State

// Create a list with just one out pointer
func oneOut(outp **State) OutList {
	l := make(OutList, 1)
	l[0] = outp
	return l
}

// Patches the out pointers in the
// list, that is, make all state pointers
// in the list point to s
func (l *OutList) patch(s *State) {
	for i := range *l {
		*((*l)[i]) = s
	}
}

// Make the NFA
// Returns a pointer to the first State of the NFA
func makeNFA(postfix string) (*State, error) {
	if len(postfix) == 0 {
		return nil, errors.New("Invalid input to makeNFA")
	}

	// var e1, e2, e Frag
	var frag1, frag2 Frag
	stack := make(Fstack, 0, len(postfix))

	var s *State
	//s := new(State)
	//fmt.Println(s)

	// See https://swtch.com/~rsc/regexp/regexp1.html
	// For detailed diagrams
	for _, c := range postfix {
		switch c {
		// Concat
		case '.':
			frag2 = stack.pop()
			frag1 = stack.pop()
			frag1.out.patch(frag2.start)
			stack.push(Frag{frag1.start, frag2.out})

		// OR
		case '|':
			frag2 = stack.pop()
			frag1 = stack.pop()
			s = &State{stype: SPLIT, out: frag1.start, out1: frag2.start}
			stack.push(Frag{s, append(frag1.out, frag2.out...)})

		// Zero or one
		case '?':
			frag1 = stack.pop()
			s = &State{stype: SPLIT, out: frag1.start}
			stack.push(Frag{s, append(frag1.out, oneOut(&s.out)...)})

		// Zero or more
		case '*':
			frag1 = stack.pop()
			s = &State{stype: SPLIT, out: frag1.start}
			frag1.out.patch(s)
			stack.push(Frag{s, oneOut(&s.out1)})

		// One or more
		case '+':
			frag1 = stack.pop()
			s = &State{stype: SPLIT, out: frag1.start}
			frag1.out.patch(s)
			stack.push(Frag{frag1.start, oneOut(&s.out1)})

		// Non-op characters
		default:
			s = &State{char: byte(c)}
			//fmt.Printf("Add state w/ char = %v : %v\n", string(c), s)
			stack.push(Frag{s, oneOut(&s.out)})
			//fmt.Printf("%v\n", &s.out)
			//fmt.Printf("Push Frag : %v\n", Frag{s, oneOut(&s.out)})
			//fmt.Println(len(oneOut(s.out)))
		}
	}

	frag1 = stack.pop()
	if len(stack) != 0 {
		return nil, errors.New("Error building NFA. Most likely mismatched parentheses")
	}

	frag1.out.patch(&State{stype: ACCEPT})

	return frag1.start, nil
}

// Matching State
//const matchState = State{0, MATCH}

// Time to simulate the NFA
type StateList []*State

// Package variables for NFA simulation
var (
	listId = 0
	l1, l2 StateList
)

// Initial state list (just the
// starting state)
func initList(start *State) *StateList {
	list := make(StateList, 0, 1)
	// fmt.Printf("initList: list = %v; start = %v\n", list, start)
	listId++
	list.addState(start)
	//fmt.Printf("initList, addState: list = %v\n", list)
	return &list
}

// Returns true if the given StateList
// contains an accept state.
func (sl StateList) isMatch() bool {
	for i := range sl {
		if sl[i].stype == ACCEPT {
			return true
		}
	}

	return false
}

// Adds a State to a StateList while
// avoiding duplicates
func (sl *StateList) addState(s *State) {
	//fmt.Printf("addState: sl = %v; s = %v\n", sl, s)
	// Base case:
	// If the state is nil or is the same
	// as the last one, be done.
	if s == nil || s.lastlist == listId {
		//fmt.Println("State is nil or s.lastlist == listId")
		return
	}
	s.lastlist = listId
	// If it's split, recurse
	if s.stype == SPLIT {
		// Epsilon transitions
		sl.addState(s.out)
		sl.addState(s.out1)
	}
	// Append the new state
	*sl = append(*sl, s)
}

// Step forward in the NFA with given char
// Using clist for current states and
// building nlist for the next iteration
func step(clist *StateList, c byte) *StateList {

	listId++
	nlist := make(StateList, 0, len(*clist))

	// var s *State

	for i := range *clist {
		s := (*clist)[i]
		//fmt.Println(s)
		if s.char == c {
			nlist.addState(s.out)
		}
	}

	return &nlist
}

func runNFA(start *State, str string) bool {
	clist := initList(start)
	//fmt.Println(clist)
	// nlist := make(StateList, 0, len(clist))
	//var nlist StateList

	for i := range str {
		nlist := step(clist, byte(str[i]))
		// nlist becomes clist
		clist = nlist
	}

	return clist.isMatch()
}

func main() {

	var inRegex, inStr string
	var err error

	if len(os.Args) != 3 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter regex: ")
		inRegex, err = reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print("Enter string: ")
		inStr, err = reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

	} else {
		inRegex = os.Args[1]
		inStr = os.Args[2]
	}

	regex := inRegex
	str := inStr

	preProc, err := preProcess(regex)
	if err != nil {
		log.Fatal(err)
	}

	postfix, err := toPostfix(preProc)
	if err != nil {
		log.Fatal(err)
	}

	nfa, err := makeNFA(postfix)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(runNFA(nfa, str))

	if !runNFA(nfa, str) {
		os.Exit(1)
	}

	return

}
