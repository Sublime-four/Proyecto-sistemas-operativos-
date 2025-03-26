package main

import (
	"bufio"         // buffers
	"crypto/sha256" // hash
	"fmt"           // Imprimir por pantalla
	"net"           // Socket
	"os"            // Entrada y salida
	"runtime"       // Habilita secuencias de escape ANSI
	"strconv"       // Convertir a string
	"strings"       // Para la manipulación de strings
	"syscall"       //
	"time"          // Para el tiempo de espera"

	"github.com/gosuri/uilive"
	"golang.org/x/sys/windows"
	"golang.org/x/term" // Para recibir la contraseña
)

var datoIn string
var reporte string
var res string

type Shell struct {
	SocketClient *net.Conn
	readerInput  *bufio.Reader
	reader       *bufio.Reader
	writer       *bufio.Writer
	tui          *uilive.Writer
}

func InitShell(SocketClient *net.Conn) Shell {
	return Shell{SocketClient: SocketClient,
		readerInput: bufio.NewReader(os.Stdin),
		reader:      bufio.NewReader(*SocketClient),
		writer:      bufio.NewWriter(*SocketClient),
		tui:         uilive.New()}
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

	hash_password := sha256.Sum256([]byte(password)) // Realiza el hash de la contraseña
	hash := fmt.Sprintf("%x", hash_password)         // Convierte a hexadecimal
	return user, hash
}

func (s *Shell) logIn(user string, password string) string {
	s.writer.WriteString(user + ":" + password + "\n")
	s.writer.Flush()

	res, _ := s.reader.ReadString('\n')
	res = strings.TrimRight(res, "\r\n")
	return res
}

func (s *Shell) get_trafic() {
	var msg *bufio.Reader
	resp := ""
	for {
		msg = s.reader
		resp, _ = msg.ReadString('\v')
		resp = strings.TrimRight(resp, "\v")
		if len(resp) == 0 {
			resp, _ = msg.ReadString('\n')
			resp = strings.TrimRight(resp, "\r\t")
		}
		resp = strings.TrimRight(resp, "\v\r\t")

		if len(resp) > 0 {
			if resp[0] == '1' {
				reporte = resp[1:]
			} else {
				res = resp[1:]
			}
		}
	}
}

func (s *Shell) getStatus(n time.Duration) {
	s.tui.Start()
	for datoIn != "bye" {
		s.moveCursor(100, 100)
		s.clearLine()
		fmt.Fprintf(s.tui, "%s\n", reporte)
		s.tui.Flush()
		//s.moveCursor(100, 100)
		//fmt.Print(reporte)
		time.Sleep(n * time.Second)
	}
}

func (s *Shell) command() {
	time.Sleep(2 * time.Second)
	for datoIn != "bye" {
		s.moveCursor(99, 0)
		fmt.Print("-Oper# ")
		datoIn, _ = s.readerInput.ReadString('\n') // el _ indica una variable que almacena el mensaje de error, sin embargo la variable no se va a usar
		datoIn = strings.TrimRight(datoIn, "\r\n")

		if datoIn != "" {
			s.writer.WriteString(datoIn + "\n")
			s.writer.Flush()
			time.Sleep(50 * time.Millisecond)
			fmt.Println(res)
			res = ""
		}

	}
}

func (s *Shell) sendNValue(n int) {
	s.writer.WriteString(string(n) + "\n")
	s.writer.Flush()
}

func (s *Shell) sendExit() {
	s.writer.WriteString("bye\n")
	s.writer.Flush()
}

func (s *Shell) mainLoop(n int) {
	isUser := "0"
	fmt.Print("\n----------------------------SHELL OPERATIVOS (WINDOWS)--------------------------\n\n")
	fmt.Print("El Cliente se ha conectado a un server con (ip:port) : (", (*s.SocketClient).RemoteAddr(), ")\n\n")

	s.sendNValue(n)

	for i := 0; isUser == "1" && i < 3; i++ {
		isUser = s.logIn(s.askCredentials())
		if isUser == "1" {
			fmt.Print("Successful login\n\n")
		} else {
			fmt.Print("Login incorrect\n\n")
		}
	}

	if isUser != "1" {
		go s.get_trafic()
		go s.getStatus(time.Duration(n))
		go s.command()

		for datoIn != "bye" {

		}
	} else {
		fmt.Print("------ Cantidad máxima de intentos de autenticación alcanzada ------\n\n")
	}

	fmt.Print("\n +++++ GRACIAS POR USAR NUESTRO SERVIDOR, HASTA PRONTO +++++\n\n")
	s.sendExit()
	(*s.SocketClient).Close() // Cierra el socket
}

// enableVirtualTerminalProcessing habilita el procesamiento de secuencias de escape ANSI en Windows.
func enableVirtualTerminalProcessing() {
	if runtime.GOOS == "windows" {
		var mode uint32
		h := windows.Handle(os.Stdout.Fd())
		windows.GetConsoleMode(h, &mode)
		mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
		windows.SetConsoleMode(h, mode)
	}
}

// moveCursor mueve el cursor a la posición especificada.
func (s *Shell) moveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}

// clearLine limpia la línea actual de la terminal.
func (s *Shell) clearLine() {
	fmt.Print("\033[2K")
}

func main() {
	enableVirtualTerminalProcessing()
	args := os.Args

	if len(args) >= 4 {
		ip := args[1]
		port := args[2]
		n, _ := strconv.Atoi(args[3])
		var err error

		tcpAddress, _ := net.ResolveTCPAddr("tcp4", ip+":"+port) // ip destino : puerto de comunicacion del server
		var socketClient net.Conn
		socketClient, err = net.DialTCP("tcp", nil, tcpAddress) // Crea el socket cliente

		if err == nil {
			shell := InitShell(&socketClient)
			shell.mainLoop(n)
		}
	} else {
		fmt.Print("\n---Parámetros insuficientes---\n\n")
	}

}
