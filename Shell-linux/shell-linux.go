package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var exit string

type Shell struct {
	users        string
	config       string
	socketServer *net.Conn
	reader       *bufio.Reader
	writer       *bufio.Writer
}

func InitShell() Shell {
	return Shell{users: "/etc/shell/users.bd",
		config: "/etc/shell/shell.conf"}
}

func (s *Shell) getPort() string {
	// Consulta el archivo de configuración
	archivo2, _ := os.ReadFile(s.config)
	str_archivo2 := string(archivo2)

	// Divide el archivo en un arreglo
	info := strings.Split(str_archivo2, "\n")

	var port string

	for i := 0; i < len(info); i++ {
		if strings.HasPrefix(info[i], "port=") {
			line_port := strings.Split(info[i], "=")
			port = line_port[1]
		}
	}
	return port
}

func (s *Shell) is_ip_allowed() bool {
	// Consulta el archivo de configuración
	archivo2, _ := os.ReadFile(s.config)
	str_archivo2 := string(archivo2)

	// Divide el archivo en un arreglo
	info := strings.Split(str_archivo2, "\n")

	var ips []string

	for i := 0; i < len(info); i++ {
		if strings.HasPrefix(info[i], "ips=") {
			line_ips := strings.Split(info[i], "=")
			ips = strings.Split(line_ips[1], ",")

		}
	}

	ip_connected := (*s.socketServer).RemoteAddr().String()
	ip := strings.Split(ip_connected, ":")[0]

	for j := 0; j < len(ips); j++ {
		if ip == ips[j] {
			s.writer.WriteString("1\n")
			s.writer.Flush()
			return true
		}
	}
	s.writer.WriteString("0\n")
	s.writer.Flush()
	return false
}

func (s *Shell) getMaxTry() int {
	// Consulta el archivo de configuración
	archivo2, _ := os.ReadFile(s.config)
	str_archivo2 := string(archivo2)

	// Divide el archivo en un arreglo
	info := strings.Split(str_archivo2, "\n")

	var x int

	for i := 0; i < len(info); i++ {
		if strings.HasPrefix(info[i], "x=") {
			line_x := strings.Split(info[i], "=")
			x, _ = strconv.Atoi(line_x[1])
		}
	}
	return x
}

func (s *Shell) sendMaxtry() {
	s.writer.WriteString(fmt.Sprint(s.getMaxTry()) + "\n")
	s.writer.Flush()
}

func (s *Shell) setSocket(socketServer *net.Conn) {
	s.socketServer = socketServer
	s.reader = bufio.NewReader(*socketServer)
	s.writer = bufio.NewWriter(*socketServer)
}

func (s *Shell) logIn() bool {
	// Consulta la base de datos de usuarios
	archivo, _ := os.ReadFile(s.users) // Lee el archivo y lo guarda en la variable
	strArchivo := string(archivo)      // castea la variable a string

	// Agrupa todo la informacion de los usuarios en un arreglo para su posterior consulta
	paramLogin := strings.Split(strArchivo, "\n") // Divide el string por cada salto de linea y lo guarda en un arreglo

	var newparamLogin []string
	for i := 0; i < len(paramLogin)-1; i++ {
		new := strings.Split(paramLogin[i], ":") // divide el arreglo anterior en un array con cada elemento que este separado con ":"
		newparamLogin = append(newparamLogin, new...)
	}

	// Consulta los usuarios permitidos
	archivo2, _ := os.ReadFile(s.config)
	str_archivo2 := string(archivo2)

	// Agrupa los usuarios permitidos en un arreglo
	info := strings.Split(str_archivo2, "\n")

	var us_allowed []string

	for i := 0; i < len(info); i++ {
		if strings.HasPrefix(info[i], "users_allowed=") {
			users_allowed := strings.Split(info[i], "=")
			us_allowed = strings.Split(users_allowed[1], ",")
		}
	}

	// Recibe las credenciales por el socket
	credentials, _ := s.reader.ReadString('\n')
	credentials = strings.TrimRight(credentials, "\r\n")

	credentialsFormated := strings.Split(credentials, ":")

	// Consulta la base de datos de usuarios
	for j := 0; j < len(newparamLogin); j += 2 {
		if newparamLogin[j] == credentialsFormated[0] && newparamLogin[j+1] == credentialsFormated[1] {
			for k := 0; k < len(us_allowed); k++ {
				if us_allowed[k] == credentialsFormated[0] {
					s.writer.WriteString("1\n")
					s.writer.Flush()
					fmt.Println("*** USER: " + credentialsFormated[0] + " ***")
					return true
				}
			}
			s.writer.WriteString("2\n")
			s.writer.Flush()
			return false
		}
	}
	s.writer.WriteString("0\n")
	s.writer.Flush()
	return false
}

func (s *Shell) obtenerUsoCPU() float64 {
	contents, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0
	}

	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 5 || fields[0] != "cpu" {
			continue
		}

		var total, idle uint64
		for i := 1; i < len(fields); i++ {
			value, err := strconv.ParseUint(fields[i], 10, 64)
			if err != nil {
				return 0
			}
			total += value
			if i == 4 {
				idle = value
			}
		}
		cpuUsage := 100 * (float64(total-idle) / float64(total))
		return cpuUsage
	}

	return 0
}

func (s *Shell) sendStatus(n time.Duration) {
	for exit != "bye" {
		// ----- memoria ------
		inf_mem := exec.Command("free", "-m")
		res_mem, _ := inf_mem.Output()
		inf_mem_array := strings.Fields(string(res_mem))
		mem := inf_mem_array[8]
		// ----- disk ------
		inf_disk := exec.Command("df", "-m")
		res_disk, _ := inf_disk.Output()
		inf_disk_array := strings.Fields(string(res_disk))
		disk := inf_disk_array[23]
		// ----- procesador -----
		proc := fmt.Sprintf("%f", s.obtenerUsoCPU())
		//proc := strconv.FormatFloat(s.obtenerUsoCPU(), 'f', 2, 64)

		s.writer.WriteString("\v1REPORTE: [ Proc=" + proc + "% - Mem=" + mem + "M - Disk=" + disk + "M ]\v")
		s.writer.Flush()

		time.Sleep(n * time.Second)
	}
}

func (s Shell) processCommand() {
	var datoIn string
	var array_datoIn []string
	var shell *exec.Cmd
	var res []byte
	for exit != "bye" {
		datoIn = ""
		datoIn, _ = s.reader.ReadString('\n')
		datoIn = strings.TrimRight(datoIn, "\r\n")

		if datoIn != "" {

			array_datoIn = strings.Fields(datoIn)
			exit = datoIn
			if len(array_datoIn) > 1 {
				shell = exec.Command(array_datoIn[0], array_datoIn[1:]...) // el ... indica el resto de elementos del arraglo
			} else {
				shell = exec.Command(datoIn)
			}
			//shell = exec.Command(array_datoIn[0], array_datoIn[1:]...) // el ... indica el resto de elementos del arraglo
			res, _ = shell.Output()

			//fmt.Println(datoIn + ":" + string(res))

			s.writer.WriteString("\v2" + string(res) + "\v")
			s.writer.Flush()
		}
	}
}

func (s *Shell) getNValue() int {
	n, _ := s.reader.ReadString('\n')
	n = strings.TrimRight(n, "\r\n")
	nInt, _ := strconv.Atoi(n)
	return nInt
}

func (s *Shell) mainLoop() {
	res := false
	exit = " "

	fmt.Print("\nSe ha conectado al server un Cliente (ip:port) : (", (*s.socketServer).RemoteAddr(), ")\n\n")

	n := s.getNValue()

	// Autenticación
	s.sendMaxtry()
	for i := 0; !res && i < s.getMaxTry(); i++ {
		res = s.logIn()
	}

	if res {
		fmt.Println("Successful login")

		go s.sendStatus(time.Duration(n))
		go s.processCommand()

		for exit != "bye" {

		}

	} else {
		fmt.Print("\n------ Cantidad máxima de intentos de autenticación alcanzada ------\n")
	}
	fmt.Print("\n +++++ GRACIAS POR USAR NUESTRO SERVIDOR, HASTA PRONTO +++++\n")
	(*s.socketServer).Close()
}

func main() {
	shell := InitShell()
	var err error

	tcpAddress, _ := net.ResolveTCPAddr("tcp4", ":"+shell.getPort())
	SocketS, _ := net.ListenTCP("tcp", tcpAddress) // Crea el socket servidor
	var socketServer net.Conn

	fmt.Print("\n--------SHELL OPERATIVOS (LINUX)------\n\n")
	for {
		socketServer, err = SocketS.Accept() // Pone en modo escucha el socket y aca almacena todo lo que llega desde el cliente

		shell.setSocket(&socketServer)

		if err == nil && shell.is_ip_allowed() {
			shell.mainLoop()
		} else {
			fmt.Println("---IP NO PERMITIDA---")
		}
	}

}
