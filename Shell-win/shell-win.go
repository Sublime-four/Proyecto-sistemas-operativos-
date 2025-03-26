package main

import (
	"bufio"         // buffers
	"crypto/sha256" // hash
	"fmt"           // Imprimir por pantalla
	"net"           // Socket
	"os"            // Entrada y salida
	"os/exec"       // Ejecutar comandos
	"strconv"       // Convertir a string
	"strings"       // Para la manipulación de strings
	"syscall"       // Para recibir la contraseña
	"time"          // Para el tiempo de espera"

	"github.com/gdamore/tcell/v2" // teclas
	"github.com/rivo/tview"       // TUI
	"golang.org/x/term"           // Para recibir la contraseña
)

var exit string
var datoIn string
var reporte string
var res string
var report *tview.TextView
var userInput *tview.InputField
var output *tview.TextView

type Shell struct {
	SocketClient *net.Conn
	readerInput  *bufio.Reader
	reader       *bufio.Reader
	writer       *bufio.Writer
	tui          *tview.Application
}

func InitShell(SocketClient *net.Conn, tui *tview.Application) Shell {
	return Shell{SocketClient: SocketClient,
		readerInput: bufio.NewReader(os.Stdin),
		reader:      bufio.NewReader(*SocketClient),
		writer:      bufio.NewWriter(*SocketClient),
		tui:         tui}
}

func (s *Shell) is_ip_allowed() bool {
	time.Sleep(2 * time.Second)
	is_ip_allowed, _ := s.reader.ReadString('\n')
	is_ip_allowed = strings.TrimRight(is_ip_allowed, "\r\n")
	if is_ip_allowed == "1" {
		return true
	} else {
		return false
	}
}

func (s *Shell) getMaxTry() int {
	maxTry, _ := s.reader.ReadString('\n')
	maxTry = strings.TrimRight(maxTry, "\r\n")
	maxTryI, _ := strconv.Atoi(maxTry)
	return maxTryI
}

func (s Shell) askCredentials() (string, string) {
	var user string
	var password string

	fmt.Print("User: ")
	user, _ = s.readerInput.ReadString('\n')
	user = strings.TrimRight(user, "\r\n")
	//fmt.Scanf("%s", &user)

	fmt.Print("Password: ")
	estado_anterior, _ := term.MakeRaw(int(syscall.Stdin))
	defer term.Restore(int(syscall.Stdin), estado_anterior) // Se ejecuta cuando la función askCredentials() termina

	// Leer la contraseña digitada sin que se vea en la terminal
	bytePassword, _ := term.ReadPassword(int(syscall.Stdin))

	password = string(bytePassword)
	fmt.Println()

	// Lee la contraseña normalmente (se visualiza en la terminal)
	//password, _ = s.readerInput.ReadString('\n')
	//password = strings.TrimRight(password, "\r\n")
	//fmt.Scanf("%s", &password)

	hash_password := sha256.Sum256([]byte(password + user)) // Realiza el hash de la contraseña
	hash := fmt.Sprintf("%x", hash_password)                // Convierte a hexadecimal
	return user, hash
}

func (s *Shell) logIn(user string, password string) string {
	s.writer.WriteString(user + ":" + password + "\n")
	s.writer.Flush()

	res, _ := s.reader.ReadString('\n')
	res = strings.TrimRight(res, "\r\n")
	return res + user
}

func (s *Shell) get_trafic() {
	var msg *bufio.Reader
	resp := ""
	for exit != "bye" {
		msg = s.reader
		resp, _ = msg.ReadString('\v')
		resp = strings.TrimRight(resp, "\v")
		resp = strings.TrimLeft(resp, "\v")

		if len(resp) > 0 {
			if resp[0] == '1' {
				reporte = resp[1:]
			} else {
				res = resp[1:]
			}
		}
	}

}

func (s *Shell) getStatus(n time.Duration, report *tview.TextView) {
	for exit != "bye" {
		s.tui.QueueUpdateDraw(func() {
			report.Clear()
			report.SetText(reporte)
		})
		time.Sleep(n * time.Second)
	}
}

func (s *Shell) command() {
	time.Sleep(2 * time.Second)

	for exit != "bye" {
		//datoIn = ""

		if datoIn != "" {
			s.writer.WriteString(datoIn + "\n")
			s.writer.Flush()
			datoIn = ""

			//fmt.Println(res)
			time.Sleep(60 * time.Millisecond)
			s.tui.QueueUpdateDraw(func() {
				output.Clear()
				output.SetText(res)
			})
			res = ""
		}

	}
}

func (s *Shell) sendNValue(n int) {
	s.writer.WriteString(fmt.Sprint(n) + "\n")
	s.writer.Flush()
}

func (s *Shell) sendExit() {
	s.writer.WriteString("bye\n")
	s.writer.Flush()
}

// Función para limpiar la pantalla de la terminal
func (s *Shell) clearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func (s *Shell) mainLoop(n int) {
	isUser := "0"

	title := tview.NewTextView().SetText("----------------------------SHELL OPERATIVOS (WINDOWS)--------------------------").SetTextAlign(tview.AlignCenter).SetSize(1, 0).SetTextColor(tcell.Color99)
	header := tview.NewTextView().SetText("El Cliente se ha conectado a un server con (ip:port) : ("+(*s.SocketClient).RemoteAddr().String()+")").SetTextAlign(tview.AlignCenter).SetSize(1, 0).SetTextColor(tcell.ColorAqua)
	report = tview.NewTextView().SetText(reporte).SetTextAlign(tview.AlignCenter).SetTextColor(tcell.Color200)
	userInput = tview.NewInputField().SetLabel("-Oper# ").SetPlaceholder("Type comands...                          Type 'bye' to quit").SetPlaceholderTextColor(tcell.Color250)
	output = tview.NewTextView().SetText("").SetScrollable(true)
	user := tview.NewTextView().SetText("").SetTextAlign(tview.AlignCenter).SetTextColor(tcell.Color19)

	flex := tview.NewFlex().AddItem(
		tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(title, 2, 1, false).
			AddItem(header, 2, 1, false).
			AddItem(user, 2, 1, false).
			AddItem(report, 2, 1, false).
			AddItem(userInput, 2, 3, true).
			AddItem(output, 0, 1, false), 0, 2, true)

	s.sendNValue(n)

	// Autenticacion
	maxTry := s.getMaxTry()
	for i := 0; isUser[0] == '0' && i < maxTry; i++ {
		isUser = s.logIn(s.askCredentials())
		if isUser[0] == '1' {
			user.Clear()
			user.SetText("User: " + isUser[1:])
			fmt.Print("Successful login\n\n")
		} else {
			fmt.Print("Login incorrect\n\n")
		}
	}
	//fmt.Println(isUser)

	if isUser[0] == '1' {
		go s.get_trafic()
		go s.getStatus(time.Duration(n), report)
		go s.command()

		userInput.SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				datoIn = userInput.GetText()
				exit = datoIn
				userInput.SetText("") // Limpiar el campo de entrada
				if exit == "bye" {
					s.sendExit()
					s.clearScreen()
					s.tui.Stop() // Termina la aplicación
				}
			}
		})

		if err := s.tui.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
			panic(err.Error())
		}
	} else if isUser[0] == '2' {
		fmt.Print("------ Usuario no permitido ------\n\n")
	} else {
		fmt.Print("------ Cantidad máxima de intentos de autenticación alcanzada ------\n\n")
	}
	fmt.Print("\n +++++ GRACIAS POR USAR NUESTRO SERVIDOR, HASTA PRONTO +++++\n\n")
	s.sendExit()
	(*s.SocketClient).Close() // Cierra el socket
}

func main() {
	tui := tview.NewApplication()
	args := os.Args

	if len(args) >= 4 {
		ip := args[1]
		port := args[2]
		n, _ := strconv.Atoi(args[3])
		var err error

		tcpAddress, _ := net.ResolveTCPAddr("tcp4", ip+":"+port) // ip destino : puerto de comunicacion del server
		var socketClient net.Conn
		socketClient, err = net.DialTCP("tcp", nil, tcpAddress) // Crea el socket cliente
		shell := InitShell(&socketClient, tui)

		if err == nil && shell.is_ip_allowed() {
			shell.mainLoop(n)
		} else {
			fmt.Println("---IP NO PERMITIDA---")
		}
	} else {
		fmt.Print("\n---Parámetros insuficientes---\n\n")
	}

}
