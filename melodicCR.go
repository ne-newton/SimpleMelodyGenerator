//program generates a 4 measure melody and outputs a midi file of that melody.

package main

import (
	"github.com/moutend/go-midi"
	"github.com/moutend/go-midi/constant"
	"github.com/moutend/go-midi/deltatime"
	"github.com/moutend/go-midi/event"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"time"
)

//Returns a one rhythm from a slice of rhymths
//H = half note, "dQ" = dotted quarter note, "Q" = quarter note, "E" = eighth note
func firstRhythm() []string {
	rhythmList := [][]string{
		{"H", "H"},
		{"H", "Q", "Q"},
		{"Q", "Q", "H"},
		{"Q", "Q", "Q", "Q"},
		{"Q", "Q", "E", "E", "E", "E"},
		{"E", "E", "E", "E", "Q", "Q"},
		{"H", "E", "E", "E", "E"},
		{"dQ", "E", "dQ", "E"},
		{"H", "dQ", "E"},
		{"dQ", "E", "Q", "E", "E"},
		{"Q", "Q", "dQ", "E"},
	}
	rand.Seed(time.Now().UnixNano())
	rhythm := rhythmList[rand.Intn(len(rhythmList))]
	return rhythm
}

//same as firstRhythm(), but with a different slice of rhythms
func thirdRhythm() []string {
	rhythmList := [][]string{
		{"Q", "Q", "Q", "Q"},
		{"Q", "Q", "E", "E", "E", "E"},
		{"E", "E", "E", "E", "Q", "Q"},
		{"Q", "Q", "Q", "E", "E"},
		{"E", "E", "Q", "E", "E", "E", "E"},
		{"dQ", "E", "Q", "Q"},
		{"dQ", "E", "E", "E", "E", "E"},
	}
	rand.Seed(time.Now().UnixNano())
	rhythm := rhythmList[rand.Intn(len(rhythmList))]
	return rhythm
}

func sliceIndex(inE int8, inL []int8) int8{
	var outI int8
	for n, listE := range inL {
		if inE == listE {
			outI = int8(n)
		}
	}
	return outI
}

//for the key of c major, takes a pitch as input and outputs the pitch one step lower in the scale.
func downScale(note int8) int8 {
	inList := []int8{45, 47, 48, 50, 52, 53, 55, 57, 59, 60, 62, 64, 65, 67, 69, 71, 72, 74, 76, 77}
	outList := []int8{43, 45, 47, 48, 50, 52, 53, 55, 57, 59, 60, 62, 64, 65, 67, 69, 71, 72, 74, 76}
	outI := sliceIndex(note, inList)
	return outList[outI]
}

//the same as downScale(), but outputs a pitch one step higher in the scale.
func upScale(note int8) int8 {
	inList := []int8{45, 47, 48, 50, 52, 53, 55, 57, 59, 60, 62, 64, 65, 67, 69, 71, 72, 74, 76, 77}
	outList := []int8{47, 48, 50, 52, 53, 55, 57, 59, 60, 62, 64, 65, 67, 69, 71, 72, 74, 76, 77, 79}
	outI := sliceIndex(note, inList)
	return outList[outI]
}

//takes harmony (tonic, predominant, or dominant) and a rhythm produced by firstRhythm() or thirdRhythm()
//then outputs random pitches from that harmony for that rhythm length
func createMelody(harmony string, rhythm []string) []int8 {
	harmonyDict := map[string][]int8{
		"tonic": {60, 64, 67},
		"predominant": {60, 62, 65, 69},
		"dominant": {62, 65, 67, 71},
	}
	var melody []int8
	rand.Seed(time.Now().UnixNano())
	harmonySlice := harmonyDict[harmony]
	for i := 0; i < len(rhythm); i++ {
		melody = append(melody, harmonySlice[rand.Intn(len(harmonySlice))])
	}
	return melody
}

//takes a melody and checks it for large leaps.
//if a leap between two pitches is greater than a perfect fifth, the second pitch is shifted by an octave
//up or down, to make the leap smaller
func checkLeaps(melody []int8) []int8 {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < len(melody) - 1; i++ {
		if int8(math.Abs(float64(melody[i]) - float64(melody[i + 1]))) > 7 {
			if melody[i] > melody[i + 1] {
				melody[i + 1] = melody[i + 1] + 12
			} else {
				melody[i + 1] = melody[i + 1] - 12
			}
		}
		if rand.Float32() > 0.8 {
			if int8(math.Abs(float64(melody[i]) - float64(melody[i + 1]))) == 7 {
				if melody[i] > melody[i + 1] {
					melody[i + 1] = melody[i + 1] + 12
				} else {
					melody[i + 1] = melody[i + 1] - 12
				}
			}
		}
	}
	return melody
}

//Takes a melody and inserts passing notes. If two pitches are a third (3 or 4 semitones) apart and separated by one note
//there is a chance that the middle note will be replaced by a passing tone.
//Eg: n, x, m - if n and m are a third apart, x might be changed to a passing tone that connects n and m.
func addPassing(melody []int8) []int8 {
	for i := 0; i < len(melody) - 2; i++ {
		switch int8(math.Abs(float64(melody[i] - melody[i + 2]))) {
		case 3, 4:
			rand.Seed(time.Now().UnixNano())
			if rand.Float32() > 0.3 {
				if melody[i] > melody[i + 2] {
					melody[i + 1] = downScale(melody[i])
				} else {
					melody[i + 1] = upScale(melody[i])
				}
			}
		}
	}
	return melody
}

//similar to addPassing() except adds a neighbor tone rather than a passing tone.
//give n, x, m, if n and m are equal, there is a chance that x will be replaced with a neighbor tone
//within this, there is a chance that the neighbor will be upper or lower.
func addNeighbor(melody []int8) []int8 {
	for i := 0; i < len(melody) - 2; i++ {
		if melody[i] == melody[i + 2] {
			rand.Seed(time.Now().UnixNano())
			if rand.Float32() < 0.4 {
				if rand.Float32() > 0.7 {
					melody[i + 1] = upScale(melody[i])
				} else {
					melody[i + 1] = downScale(melody[i])
				}
			}
		}
	}
	return melody
}

//creates the data for the first measure.
//retrieves a rhythm string from firstRhythm(), than uses that to create a tonic melody with createMelody().
//passing and neighboring notes are then added, after if it is checked for leaps.
func measure1() ([]int8, []string){
	measureRhythm := firstRhythm()
	measureNotes := createMelody("tonic", measureRhythm)
	measureNotes = addNeighbor(measureNotes)
	measureNotes = addPassing(checkLeaps(measureNotes))
	return measureNotes, measureRhythm
}

//takes a melody that is based around a tonic harmony and adapts it for a dominant harmony.
func tonToDom(melody []int8) []int8 {
	var newMelody []int8
	inList := []int8{43, 45, 47, 48, 50, 52, 53, 55, 57, 59, 60, 62, 64, 65, 67, 69, 71, 72, 74, 76, 77}
	var outList []int8
	if rand.Float32() >= 0.5 {
		outList = []int8{43, 45, 47, 47, 48, 50, 53, 55, 57, 59, 59, 60, 62, 65, 67, 69, 71, 71, 72, 74, 77}
	} else {
		outList = []int8{43, 45, 47, 50, 42, 53, 53, 55, 57, 59, 62, 64, 65, 65, 67, 69, 71, 74, 76, 77, 77}
	}
	for _, i := range melody {
		outI := sliceIndex(i, inList)
		newMelody = append(newMelody, outList[outI])
	}
	return newMelody
}

//takes a melody that is based around a tonic harmonic and adapts if for a subdominant harmony.
func tonToSDom(melody []int8) []int8 {
	var newMelody []int8
	inList := []int8{43, 45, 47, 48, 50, 52, 53, 55, 57, 59, 60, 62, 64, 65, 67, 69, 71, 72, 74, 76, 77}
	outList := []int8{45, 45, 47, 48, 50, 53, 55, 57, 57, 59, 60, 62, 65, 67, 69, 69, 71, 72, 74, 77, 77}
	for _, i := range melody {
		outI := sliceIndex(i, inList)
		newMelody = append(newMelody, outList[outI])
	}
	return newMelody
}

//creates a second measure based on the first measure. The second measure could be tonic, subdominant, or dominant.
func measure2(melody []int8, rhythm []string) ([]int8, []string) {
	rand.Seed(time.Now().UnixNano())
	n := rand.Float32()
	measureRhythm := rhythm
	var measureNotes []int8
	if n <= 0.3 {
		measureNotes = melody
	} else if n > 0.3 && n <= 0.8 {
		measureNotes = tonToDom(melody)
	} else {
		measureNotes = tonToSDom(melody)
	}
	return measureNotes, measureRhythm
}

func inSlice(sliceE int8, sliceI []int8) bool {
	isIn := false
	for _, e := range sliceI {
		if sliceE == e {
			isIn = true
		}
	}
	return isIn
}

//makes sure that the penultimate note of the melody moves to the final note by step.
func penultimateNote(melody []int8) []int8 {
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(2)
	preTonic := []int8{62, 71}
	lastI := len(melody) - 1
	lastE := melody[lastI]
	if inSlice(lastE, preTonic) == false {
		melody[lastI] = preTonic[n]
	}
	return melody
}

//there is a chanced that the transition between measures will contain a large leap. This function smooths that out.
func smoothMeasures(finalNote int8, firstNote int8) int8 {
	if math.Abs(float64(finalNote) - float64(firstNote)) > 6 {
		if finalNote > firstNote {
			firstNote = firstNote + 12
		} else {
			firstNote = firstNote - 12
		}
	}
	return firstNote
}

//checks melodies for augmented fourth or diminished fifth leaps, if it finds any, it alters the leap
func checkTT(melody []int8) []int8 {
	for i := 0; i < len(melody) - 1; i++ {
		if math.Abs(float64(melody[i]) - float64(melody[i + 1])) == 6 {
			if melody[i] % 12 == 5 {
				melody[i] = melody[i] + 2
			} else if melody[i + 1] % 12 == 5{
				melody[i + 1] = melody[i + 1] + 2
			}
		}
	}
	return melody
}

var m1melody, m1rhythm = measure1()
var m2melody, m2rhythm = measure2(m1melody, m1rhythm)

//similar to measure 1, but for measure three.
func measure3() ([]int8, []string) {
	melLen := len(m2melody)
	measureRhythm := thirdRhythm()
	measureNotes := createMelody("dominant", measureRhythm)
	measureNotes[0] = smoothMeasures(m2melody[melLen - 1], measureNotes[0])
	measureNotes = penultimateNote(measureNotes)
	measureNotes = checkLeaps(measureNotes)
	measureNotes = checkTT(measureNotes)
	measureNotes = addNeighbor(measureNotes)
	measureNotes = addPassing(measureNotes)
	return measureNotes, measureRhythm
}

var m3melody, m3rhythm = measure3()

//the final measure is just a root whole note.
func measure4() ([]int8, []string) {
	measureRhythm := []string{"W"}
	melLen := len(m3melody)
	leadingTone := m3melody[melLen - 1]
	var measureNotes []int8
	switch leadingTone {
	case 71, 74:
		measureNotes = append(measureNotes, 72)
	default :
		measureNotes = append(measureNotes, 60)
	}
	return measureNotes, measureRhythm
}

var m4melody, m4rhythm = measure4()

func joinIntSlice(slice1 []int8, slice2 []int8) []int8 {
	sliceOut := slice1
	for _, i := range slice2 {
		sliceOut = append(sliceOut, int8(i))
	}
	return sliceOut
}

func createPhraseMelody() []int8 {
	var phrase []int8
	phrase = joinIntSlice(phrase, m1melody)
	phrase = joinIntSlice(phrase, m2melody)
	phrase = joinIntSlice(phrase, m3melody)
	phrase = joinIntSlice(phrase, m4melody)
	return phrase
}

func joinStrSlice(slice1 []string, slice2 []string) []string {
	sliceOut := slice1
	for _, i := range slice2 {
		sliceOut = append(sliceOut, string(i))
	}
	return sliceOut
}

func createPhraseRhythm() []string {
	var phrase []string
	phrase = joinStrSlice(phrase, m1rhythm)
	phrase = joinStrSlice(phrase, m2rhythm)
	phrase = joinStrSlice(phrase, m3rhythm)
	phrase = joinStrSlice(phrase, m4rhythm)
	return phrase
}

var phraseMelody = createPhraseMelody()
var phraseRhythm = createPhraseRhythm()

//creates the midi file
func createMidi() {
	var trackOn []*event.NoteOnEvent
	var trackOff []*event.NoteOffEvent
	callMidi := midi.MIDI{}
	callMidi.TimeDivision().SetBPM(480)
	deltaZero, _ := deltatime.New(0)
	bpm, _ := event.NewSetTempoEvent(deltaZero, 600000)
	numerator := 4
	timeSig, _ := event.NewTimeSignatureEvent(deltaZero, uint8(numerator), 2, 24, 8)
	callTrack := midi.NewTrack()
	rhythmValues := map[string]int {
		"W":  1920,
		"H":  960,
		"dQ": 720,
		"Q": 480,
		"E": 240,
	}
	for i := 0; i < len(phraseMelody); i++ {
		deltaOff, _ := deltatime.New(rhythmValues[phraseRhythm[i]])
		note := byte(phraseMelody[i])
		noteOn, _ := event.NewNoteOnEvent(deltaZero, 0, constant.Note(note), 100)
		noteOff, _ := event.NewNoteOffEvent(deltaOff, 0, constant.Note(note), 0)
		trackOn = append(trackOn, noteOn)
		trackOff = append(trackOff, noteOff)
	}
	endTrack, _ := event.NewEndOfTrackEvent(deltaZero)
	callTrack.Events = append(callTrack.Events, bpm)
	callTrack.Events = append(callTrack.Events, timeSig)
	for i := 0; i < len(trackOn); i++ {
		callTrack.Events = append(callTrack.Events, trackOn[i])
		callTrack.Events = append(callTrack.Events, trackOff[i])
	}
	callTrack.Events = append(callTrack.Events, endTrack)
	callMidi.Tracks = append(callMidi.Tracks, callTrack)
	err := ioutil.WriteFile("melody.mid", callMidi.Serialize(), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	createMidi()
}
