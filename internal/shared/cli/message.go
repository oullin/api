package cli

import "fmt"

func Error(message string) {
	fmt.Print(RedColour + message + Reset)
}

func Errorln(message string) {
	fmt.Println(RedColour + message + Reset)
}

func Success(message string) {
	fmt.Print(GreenColour + message + Reset)
}

func Successln(message string) {
	fmt.Println(GreenColour + message + Reset)
}

func Warning(message string) {
	fmt.Print(YellowColour + message + Reset)
}

func Warningln(message string) {
	fmt.Println(YellowColour + message + Reset)
}

func Magenta(message string) {
	fmt.Print(MagentaColour + message + Reset)
}

func Magentaln(message string) {
	fmt.Println(MagentaColour + message + Reset)
}

func Blue(message string) {
	fmt.Print(BlueColour + message + Reset)
}

func Blueln(message string) {
	fmt.Println(BlueColour + message + Reset)
}

func Cyan(message string) {
	fmt.Print(CyanColour + message + Reset)
}

func Cyanln(message string) {
	fmt.Println(CyanColour + message + Reset)
}

func Gray(message string) {
	fmt.Print(GrayColour + message + Reset)
}

func Grayln(message string) {
	fmt.Println(GrayColour + message + Reset)
}
